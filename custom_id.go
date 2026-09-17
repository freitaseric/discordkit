package discordkit

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"
)

// CustomID identifies an interactive component or modal. Validate before sending.
type CustomID string

// Validate checks Discord's 1–100 character limit and excludes control characters.
func (id CustomID) Validate() error {
	if !utf8.ValidString(string(id)) || utf8.RuneCountInString(string(id)) < 1 || utf8.RuneCountInString(string(id)) > 100 {
		return fmt.Errorf("%w: expected 1–100 UTF-8 characters", ErrInvalidCustomID)
	}
	for _, r := range id {
		if unicode.IsControl(r) {
			return fmt.Errorf("%w: control character", ErrInvalidCustomID)
		}
	}
	return nil
}

// RouteBuilder binds named path parameters without treating parameter values as paths.
type RouteBuilder struct {
	pattern string
	params  map[string]string
}

// Route starts a custom ID from a pattern such as /jobs/:jobID/save.
func Route(pattern string) *RouteBuilder {
	return &RouteBuilder{pattern: pattern, params: map[string]string{}}
}

// Param sets a raw, unescaped parameter. Build escapes each value exactly once.
func (b *RouteBuilder) Param(name, value string) *RouteBuilder { b.params[name] = value; return b }

// Build validates the pattern and all bindings, and enforces the final ID limit.
func (b *RouteBuilder) Build() (CustomID, error) {
	segments, err := parsePattern(b.pattern)
	if err != nil {
		return "", err
	}
	used := map[string]bool{}
	out := make([]string, len(segments))
	for i, s := range segments {
		v := s.literal
		if s.param != "" {
			var ok bool
			v, ok = b.params[s.param]
			if !ok || v == "" {
				return "", fmt.Errorf("%w: missing parameter %q", ErrInvalidCustomID, s.param)
			}
			used[s.param] = true
		}
		if !validSegment(v) {
			return "", fmt.Errorf("%w: invalid parameter", ErrInvalidCustomID)
		}
		out[i] = url.PathEscape(v)
	}
	if len(used) != len(b.params) {
		return "", fmt.Errorf("%w: unused parameter", ErrInvalidCustomID)
	}
	prefix := ""
	if strings.HasPrefix(b.pattern, "/") {
		prefix = "/"
	}
	id := CustomID(prefix + strings.Join(out, "/"))
	return id, id.Validate()
}

// MustBuild is Build with panic on error, for static application configuration.
func (b *RouteBuilder) MustBuild() CustomID {
	id, err := b.Build()
	if err != nil {
		panic(err)
	}
	return id
}

type pathSegment struct{ literal, param string }

func validSegment(s string) bool {
	if s == "" || !utf8.ValidString(s) {
		return false
	}
	for _, r := range s {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}
func splitID(s string) ([]string, error) {
	if err := CustomID(s).Validate(); err != nil {
		return nil, err
	}
	parts := strings.Split(strings.TrimPrefix(s, "/"), "/")
	for i, p := range parts {
		v, e := url.PathUnescape(p)
		if e != nil || !validSegment(v) {
			return nil, fmt.Errorf("%w: malformed path segment", ErrInvalidCustomID)
		}
		parts[i] = v
	}
	return parts, nil
}
func parsePattern(pattern string) ([]pathSegment, error) {
	if err := CustomID(pattern).Validate(); err != nil {
		return nil, err
	}
	parts := strings.Split(strings.TrimPrefix(pattern, "/"), "/")
	out := make([]pathSegment, len(parts))
	seen := map[string]bool{}
	for i, p := range parts {
		if strings.HasPrefix(p, ":") {
			name := strings.TrimPrefix(p, ":")
			if !validSegment(name) || strings.ContainsAny(name, ":% ") || seen[name] {
				return nil, fmt.Errorf("%w: invalid or repeated parameter", ErrInvalidCustomID)
			}
			seen[name] = true
			out[i].param = name
		} else {
			v, e := url.PathUnescape(p)
			if e != nil || !validSegment(v) {
				return nil, fmt.Errorf("%w: invalid pattern", ErrInvalidCustomID)
			}
			out[i].literal = v
		}
	}
	return out, nil
}
