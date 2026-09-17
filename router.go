package discordkit

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
)

// Handler handles a single interaction.
type Handler func(*Context) error

// Middleware wraps a Handler. The first middleware registered runs outermost.
type Middleware func(Handler) Handler

type routeKind uint8

const (
	commandRoute routeKind = iota
	componentRoute
	modalRoute
	autocompleteRoute
)

type routeEntry struct {
	kind     routeKind
	path     string
	segments []pathSegment
	handler  Handler
}

// Router is a shared routing core for commands, components, modals and autocomplete.
// Recovery is always outermost. Registration and dispatch are concurrency safe.
type Router struct {
	mu         sync.RWMutex
	routes     []routeEntry
	middleware []Middleware
	onError    func(*Context, error)
}

// NewRouter creates a router. Middleware executes in the order supplied.
func NewRouter(middleware ...Middleware) *Router {
	return &Router{middleware: append([]Middleware(nil), middleware...)}
}

// Use appends global middleware, including for previously registered routes.
func (r *Router) Use(m ...Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middleware = append(r.middleware, m...)
}

// OnError sets the error hook. A nil hook uses structured server-side logging.
// Error hooks should not panic. No stack trace or automatic error reply is sent.
func (r *Router) OnError(h func(*Context, error)) { r.mu.Lock(); defer r.mu.Unlock(); r.onError = h }

// Command registers a slash or context command path, e.g. "jobs search".
func (r *Router) Command(path string, h Handler, m ...Middleware) error {
	return r.add(commandRoute, path, h, m)
}

// Component registers a custom ID route, with optional named path parameters.
func (r *Router) Component(path string, h Handler, m ...Middleware) error {
	return r.add(componentRoute, path, h, m)
}

// Modal registers a modal submission route.
func (r *Router) Modal(path string, h Handler, m ...Middleware) error {
	return r.add(modalRoute, path, h, m)
}

// Autocomplete registers a command path and focused option name.
func (r *Router) Autocomplete(path, option string, h Handler, m ...Middleware) error {
	if option == "" || strings.ContainsAny(option, " \t\n") {
		return fmt.Errorf("%w: invalid focused option", ErrRouteConflict)
	}
	return r.add(autocompleteRoute, path+" "+option, h, m)
}
func chain(h Handler, m []Middleware) Handler {
	for i := len(m) - 1; i >= 0; i-- {
		if m[i] != nil {
			h = m[i](h)
		}
	}
	return h
}
func (r *Router) add(k routeKind, path string, h Handler, m []Middleware) error {
	if h == nil || path == "" {
		return fmt.Errorf("%w: empty route or nil handler", ErrRouteConflict)
	}
	entry := routeEntry{kind: k, path: path, handler: chain(h, m)}
	if k == componentRoute || k == modalRoute {
		p, e := parsePattern(path)
		if e != nil {
			return e
		}
		entry.segments = p
	} else if strings.Join(strings.Fields(path), " ") != path {
		return fmt.Errorf("%w: noncanonical command path", ErrRouteConflict)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, old := range r.routes {
		if old.kind != k {
			continue
		}
		conflict := old.path == path
		if entry.segments != nil && len(entry.segments) == len(old.segments) && strings.HasPrefix(path, "/") == strings.HasPrefix(old.path, "/") {
			conflict = true
			for i, s := range entry.segments {
				o := old.segments[i]
				if (s.param == "") != (o.param == "") || s.literal != o.literal {
					conflict = false
					break
				}
			}
		}
		if conflict {
			return fmt.Errorf("%w: %s", ErrRouteConflict, path)
		}
	}
	r.routes = append(r.routes, entry)
	return nil
}
func commandPath(d discordgo.ApplicationCommandInteractionData) string {
	p := d.Name
	opts := d.Options
	for {
		var child *discordgo.ApplicationCommandInteractionDataOption
		for _, o := range opts {
			if o != nil && (o.Type == 1 || o.Type == 2) {
				child = o
				break
			}
		}
		if child == nil {
			return p
		}
		p += " " + child.Name
		opts = child.Options
	}
}
func interactionRoute(c *Context) (routeKind, string, error) {
	i := c.interaction()
	if i == nil {
		return 0, "", ErrInvalidInteraction
	}
	switch i.Type {
	case discordgo.InteractionApplicationCommand, discordgo.InteractionApplicationCommandAutocomplete:
		d, ok := c.commandData()
		if !ok {
			return 0, "", ErrInvalidInteraction
		}
		p := commandPath(d)
		if i.Type == discordgo.InteractionApplicationCommandAutocomplete {
			o, ok := c.FocusedOption()
			if !ok {
				return 0, "", ErrInvalidInteraction
			}
			return autocompleteRoute, p + " " + o.Name, nil
		}
		return commandRoute, p, nil
	case discordgo.InteractionMessageComponent:
		switch d := i.Data.(type) {
		case discordgo.MessageComponentInteractionData:
			return componentRoute, d.CustomID, nil
		case *discordgo.MessageComponentInteractionData:
			if d != nil {
				return componentRoute, d.CustomID, nil
			}
		}
	case discordgo.InteractionModalSubmit:
		switch d := i.Data.(type) {
		case discordgo.ModalSubmitInteractionData:
			return modalRoute, d.CustomID, nil
		case *discordgo.ModalSubmitInteractionData:
			if d != nil {
				return modalRoute, d.CustomID, nil
			}
		}
	}
	return 0, "", ErrInvalidInteraction
}

// Dispatch routes an interaction and returns handler errors. It does not call OnError.
func (r *Router) Dispatch(c *Context) error {
	k, path, e := interactionRoute(c)
	if e != nil {
		return e
	}
	var parts []string
	if k == componentRoute || k == modalRoute {
		parts, e = splitID(path)
		if e != nil {
			return e
		}
	}
	r.mu.RLock()
	var best *routeEntry
	var params map[string]string
	for idx := range r.routes {
		entry := &r.routes[idx]
		if entry.kind != k {
			continue
		}
		if parts == nil {
			if entry.path == path {
				best = entry
				break
			}
			continue
		}
		if len(entry.segments) != len(parts) || strings.HasPrefix(path, "/") != strings.HasPrefix(entry.path, "/") {
			continue
		}
		p := map[string]string{}
		match := true
		for j, s := range entry.segments {
			if s.param != "" {
				p[s.param] = parts[j]
			} else if s.literal != parts[j] {
				match = false
				break
			}
		}
		if !match {
			continue
		}
		better := best == nil
		if best != nil {
			for j, s := range entry.segments {
				a, b := s.param == "", best.segments[j].param == ""
				if a != b {
					better = a
					break
				}
			}
		}
		if better {
			best = entry
			params = p
		}
	}
	var h Handler
	if best != nil {
		h = best.handler
	}
	mw := append([]Middleware(nil), r.middleware...)
	r.mu.RUnlock()
	if h == nil {
		return fmt.Errorf("%w: %s", ErrRouteNotFound, path)
	}
	c.params = params
	return Recovery()(func(c *Context) error { return chain(h, mw)(c) })(c)
}

// Handle is a discordgo InteractionCreate handler. Attach using Session.AddHandler.
func (r *Router) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	c := NewContext(s, i)
	if err := r.Dispatch(c); err != nil {
		r.mu.RLock()
		hook := r.onError
		r.mu.RUnlock()
		if hook != nil {
			if hookErr := Recovery()(func(c *Context) error { hook(c, err); return nil })(c); hookErr != nil {
				slog.Error("discordkit error hook panicked", "error", hookErr)
			}
		} else {
			slog.Error("discordkit interaction failed", "error", err)
		}
	}
}

// Group shares a route prefix and middleware. Command prefixes use spaces;
// custom ID prefixes use slashes. A group does not create a separate router.
type Group struct {
	router     *Router
	prefix     string
	middleware []Middleware
}

// Group creates a group whose registrations share this router.
func (r *Router) Group(prefix string, m ...Middleware) *Group {
	return &Group{r, prefix, append([]Middleware(nil), m...)}
}
func (g *Group) path(p, sep string) string {
	if g.prefix == "" {
		return p
	}
	return strings.TrimRight(g.prefix, sep) + sep + strings.TrimLeft(p, sep)
}

// Command registers a command in the group.
func (g *Group) Command(p string, h Handler, m ...Middleware) error {
	return g.router.Command(g.path(p, " "), h, append(append([]Middleware(nil), g.middleware...), m...)...)
}

// Component registers a component in the group.
func (g *Group) Component(p string, h Handler, m ...Middleware) error {
	return g.router.Component(g.path(p, "/"), h, append(append([]Middleware(nil), g.middleware...), m...)...)
}

// Modal registers a modal in the group.
func (g *Group) Modal(p string, h Handler, m ...Middleware) error {
	return g.router.Modal(g.path(p, "/"), h, append(append([]Middleware(nil), g.middleware...), m...)...)
}

// Autocomplete registers an autocomplete handler in the group.
func (g *Group) Autocomplete(p, o string, h Handler, m ...Middleware) error {
	return g.router.Autocomplete(g.path(p, " "), o, h, append(append([]Middleware(nil), g.middleware...), m...)...)
}
