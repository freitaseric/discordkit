package discordkit

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/bwmarrin/discordgo"
)

// fakeRT is an http.RoundTripper that records requests and returns a canned
// 200 response, letting response lifecycle tests run without network access.
type fakeRT struct{ calls int }

func (f *fakeRT) RoundTrip(*http.Request) (*http.Response, error) {
	f.calls++
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte(`{}`))),
		Header:     make(http.Header),
	}, nil
}

func newTestContext(t *testing.T, typ discordgo.InteractionType) (*Context, *fakeRT) {
	t.Helper()
	s, err := discordgo.New("Bot test")
	if err != nil {
		t.Fatalf("session: %v", err)
	}
	rt := &fakeRT{}
	s.Client = &http.Client{Transport: rt}
	i := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID:    "1",
		Token: "tok",
		Type:  typ,
	}}
	return NewContext(s, i), rt
}

func TestReplyThenSecondResponseFails(t *testing.T) {
	c, _ := newTestContext(t, discordgo.InteractionApplicationCommand)
	if err := c.ReplyText("hello"); err != nil {
		t.Fatalf("first reply: %v", err)
	}
	if err := c.ReplyText("again"); !errors.Is(err, ErrAlreadyResponded) {
		t.Fatalf("expected ErrAlreadyResponded, got %v", err)
	}
}

func TestEditBeforeAckFails(t *testing.T) {
	c, _ := newTestContext(t, discordgo.InteractionApplicationCommand)
	if _, err := c.Edit(MessageSpec{Content: "x"}); !errors.Is(err, ErrNotResponded) {
		t.Fatalf("expected ErrNotResponded, got %v", err)
	}
	if _, err := c.Followup(MessageSpec{Content: "x"}); !errors.Is(err, ErrNotResponded) {
		t.Fatalf("expected ErrNotResponded for followup, got %v", err)
	}
}

func TestDeferThenEditSucceeds(t *testing.T) {
	c, _ := newTestContext(t, discordgo.InteractionApplicationCommand)
	if err := c.Defer(false); err != nil {
		t.Fatalf("defer: %v", err)
	}
	if _, err := c.Edit(MessageSpec{Content: "done"}); err != nil {
		t.Fatalf("edit after defer: %v", err)
	}
	if err := c.Defer(false); !errors.Is(err, ErrAlreadyResponded) {
		t.Fatalf("expected ErrAlreadyResponded on second defer, got %v", err)
	}
}

func TestAutocompleteOnlyOnAutocompleteInteraction(t *testing.T) {
	c, _ := newTestContext(t, discordgo.InteractionApplicationCommand)
	if err := c.Autocomplete(Choice("a", "a")); !errors.Is(err, ErrInvalidInteraction) {
		t.Fatalf("expected ErrInvalidInteraction, got %v", err)
	}
	ac, _ := newTestContext(t, discordgo.InteractionApplicationCommandAutocomplete)
	if err := ac.Autocomplete(Choice("a", "a")); err != nil {
		t.Fatalf("autocomplete: %v", err)
	}
}

func TestShowModalRejectsModalSubmit(t *testing.T) {
	c, _ := newTestContext(t, discordgo.InteractionModalSubmit)
	m := ModalOf("id", "Title", discordgo.Label{Label: "L", Component: discordgo.TextInput{CustomID: "x"}})
	if err := c.ShowModal(m); !errors.Is(err, ErrInvalidInteraction) {
		t.Fatalf("expected ErrInvalidInteraction, got %v", err)
	}
	ok, _ := newTestContext(t, discordgo.InteractionApplicationCommand)
	if err := ok.ShowModal(m); err != nil {
		t.Fatalf("show modal: %v", err)
	}
}
