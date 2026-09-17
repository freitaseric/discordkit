package discordkit

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
	"sync"
	"testing"
)

func interaction(t discordgo.InteractionType, d discordgo.InteractionData) *Context {
	return NewContext(nil, &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{Type: t, Data: d}})
}
func component(id string) *Context {
	return interaction(discordgo.InteractionMessageComponent, discordgo.MessageComponentInteractionData{CustomID: id})
}
func TestCustomID(t *testing.T) {
	for _, value := range []string{"a/b", "%2F", "hello world", "?x=#&", "ação", "雪", ":id", ".."} {
		id, err := Route("/jobs/:id/save").Param("id", value).Build()
		if err != nil {
			t.Fatal(err)
		}
		r := NewRouter()
		if err = r.Component("/jobs/:id/save", func(c *Context) error {
			got, ok := c.Param("id")
			if !ok || got != value {
				t.Errorf("got %q want %q", got, value)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		if err = r.Dispatch(component(string(id))); err != nil {
			t.Fatal(err)
		}
	}
	for _, pattern := range []string{"", "/", "/a//b", "/:x/:x", "/:x/:", "/a/%zz"} {
		if _, err := Route(pattern).Build(); err == nil {
			t.Errorf("accepted %q", pattern)
		}
	}
	for _, id := range []string{"/bad/%XX", "/bad//x", "/x/\x00", "/x/\xff"} {
		if err := NewRouter().Dispatch(component(id)); !errors.Is(err, ErrInvalidCustomID) {
			t.Errorf("%q: %v", id, err)
		}
	}
	if err := CustomID(strings.Repeat("a", 100)).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := CustomID(strings.Repeat("a", 101)).Validate(); err == nil {
		t.Fatal("accepted overlong ID")
	}
	if _, err := Route("/:x").Param("x", strings.Repeat("a", 100)).Build(); err == nil {
		t.Fatal("accepted overlong built ID")
	}
	if _, err := Route("/:x").Param("y", "a").Build(); err == nil {
		t.Fatal("missing binding")
	}
	if _, err := Route("/x").Param("y", "a").Build(); err == nil {
		t.Fatal("unused binding")
	}
}
func TestRouterPrecedenceAndConflicts(t *testing.T) {
	r := NewRouter()
	got := ""
	h := func(name string) Handler { return func(c *Context) error { got = name; return nil } }
	for _, p := range []string{"/:kind/:id", "/jobs/:id", "/jobs/new"} {
		if err := r.Component(p, h(p)); err != nil {
			t.Fatal(err)
		}
	}
	for id, want := range map[string]string{"/jobs/new": "/jobs/new", "/jobs/123": "/jobs/:id", "/users/123": "/:kind/:id"} {
		if err := r.Dispatch(component(id)); err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("got %s want %s", got, want)
		}
	}
	if err := r.Component("/jobs/:other", h("bad")); !errors.Is(err, ErrRouteConflict) {
		t.Fatal(err)
	}
	if err := r.Component("/jobs/new", h("bad")); !errors.Is(err, ErrRouteConflict) {
		t.Fatal(err)
	}
	if err := r.Dispatch(component("/not/found/extra")); !errors.Is(err, ErrRouteNotFound) {
		t.Fatal(err)
	}
	if err := r.Dispatch(nil); !errors.Is(err, ErrInvalidInteraction) {
		t.Fatal(err)
	}
	if err := r.Component("/nil", nil); err == nil {
		t.Fatal("nil handler")
	}
}
func TestRouterAllKindsAndMiddleware(t *testing.T) {
	var order []string
	mw := func(s string) Middleware {
		return func(next Handler) Handler {
			return func(c *Context) error {
				order = append(order, s+"+")
				err := next(c)
				order = append(order, s+"-")
				return err
			}
		}
	}
	r := NewRouter(mw("global"))
	h := func(c *Context) error { order = append(order, "handler"); return nil }
	if e := r.Command("jobs search", h, mw("local")); e != nil {
		t.Fatal(e)
	}
	if e := r.Component("/go", h, mw("local")); e != nil {
		t.Fatal(e)
	}
	if e := r.Modal("/form", h, mw("local")); e != nil {
		t.Fatal(e)
	}
	if e := r.Autocomplete("jobs search", "query", h, mw("local")); e != nil {
		t.Fatal(e)
	}
	data := discordgo.ApplicationCommandInteractionData{Name: "jobs", Options: []*discordgo.ApplicationCommandInteractionDataOption{{Name: "search", Type: 1, Options: []*discordgo.ApplicationCommandInteractionDataOption{{Name: "query", Type: 3, Focused: true}}}}}
	cases := []*Context{interaction(2, data), component("/go"), interaction(5, discordgo.ModalSubmitInteractionData{CustomID: "/form"}), interaction(4, data)}
	for _, c := range cases {
		order = nil
		if err := r.Dispatch(c); err != nil {
			t.Fatal(err)
		}
		if fmt.Sprint(order) != "[global+ local+ handler local- global-]" {
			t.Fatal(order)
		}
	}
}
func TestRecoveryAndHook(t *testing.T) {
	r := NewRouter()
	if e := r.Component("/panic", func(c *Context) error { panic("secret") }); e != nil {
		t.Fatal(e)
	}
	err := r.Dispatch(component("/panic"))
	var p *PanicError
	if !errors.As(err, &p) || len(p.Stack) == 0 || strings.Contains(err.Error(), "secret") {
		t.Fatal(err)
	}
	called := false
	r.OnError(func(c *Context, err error) { called = true; panic("hook panic") })
	r.Handle(nil, component("/panic").Interaction)
	if !called {
		t.Fatal("hook not called")
	}
}
func TestRouterConcurrent(t *testing.T) {
	r := NewRouter()
	if e := r.Component("/go", func(c *Context) error { return nil }); e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	for n := 0; n < 20; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			if e := r.Component(fmt.Sprintf("/x%d", n), func(c *Context) error { return nil }); e != nil {
				t.Error(e)
			}
			if e := r.Dispatch(component("/go")); e != nil {
				t.Error(e)
			}
		}(n)
	}
	wg.Wait()
}
