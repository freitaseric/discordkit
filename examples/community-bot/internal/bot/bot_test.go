package bot

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	dk "github.com/freitaseric/discordkit"
	"github.com/freitaseric/discordkit/examples/community-bot/internal/store"
)

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestStagesAndViews(t *testing.T) {
	for stage := 1; stage <= 6; stage++ {
		b := &Bot{Stage: stage, GuildID: "guild"}
		if _, err := b.Commands(); err != nil {
			t.Fatalf("stage %d commands: %v", stage, err)
		}
		if _, err := b.Router(); err != nil {
			t.Fatalf("stage %d routes: %v", stage, err)
		}
		if _, err := panel(stage).InteractionResponseData(); err != nil {
			t.Fatal(err)
		}
	}
	for _, spec := range []dk.MessageSpec{ticketView(store.Ticket{ID: "123", Subject: "Title", Description: "Details", Status: "open"}), queueView(nil, "open", 999)} {
		if _, err := spec.InteractionResponseData(); err != nil {
			t.Fatal(err)
		}
	}
}
func TestModalToDiskAndPrivateResponse(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.json")
	db, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	b := &Bot{Stage: 6, GuildID: "guild", Store: db}
	r, err := b.Router()
	if err != nil {
		t.Fatal(err)
	}
	var requests []map[string]any
	session, _ := discordgo.New("Bot test")
	session.Client = &http.Client{Transport: transport(func(req *http.Request) (*http.Response, error) {
		raw, _ := io.ReadAll(req.Body)
		var body map[string]any
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatal(err)
		}
		requests = append(requests, body)
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"id":"response","attachments":[]}`))}, nil
	})}
	interaction := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{ID: "123", AppID: "app", Token: "test", Type: discordgo.InteractionModalSubmit, GuildID: "guild", Member: &discordgo.Member{User: &discordgo.User{ID: "owner"}}, Data: discordgo.ModalSubmitInteractionData{CustomID: "/tickets/create", Components: []discordgo.MessageComponent{discordgo.ActionsRow{Components: []discordgo.MessageComponent{discordgo.TextInput{CustomID: "subject", Value: "Help"}, discordgo.TextInput{CustomID: "description", Value: "A detailed question"}}}}}}}
	if err = r.Dispatch(dk.NewContext(session, interaction)); err != nil {
		t.Fatal(err)
	}
	if len(requests) != 2 {
		t.Fatalf("expected defer then edit, got %d", len(requests))
	}
	if requests[0]["type"] != float64(5) || requests[0]["data"].(map[string]any)["flags"] != float64(64) {
		t.Fatal("not privately deferred", requests[0])
	}
	reopened, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = reopened.Get(store.Actor{GuildID: "guild", UserID: "owner"}, "123"); err != nil {
		t.Fatal(err)
	}
}

func TestLabPayloads(t *testing.T) {
	b := &Bot{Stage: 6, GuildID: "guild"}
	r, err := b.Router()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"layout", "form", "embed"} {
		t.Run(name, func(t *testing.T) {
			var response struct {
				Type discordgo.InteractionResponseType `json:"type"`
				Data json.RawMessage                   `json:"data"`
			}
			session, _ := discordgo.New("Bot test")
			session.Client = &http.Client{Transport: transport(func(req *http.Request) (*http.Response, error) {
				if err := json.NewDecoder(req.Body).Decode(&response); err != nil {
					t.Fatal(err)
				}
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{}`))}, nil
			})}
			event := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{ID: "123", AppID: "app", Token: "test", Type: discordgo.InteractionApplicationCommand, GuildID: "guild", Member: &discordgo.Member{User: &discordgo.User{ID: "123456789012345678"}}, Data: discordgo.ApplicationCommandInteractionData{Name: "lab", Options: []*discordgo.ApplicationCommandInteractionDataOption{{Name: name, Type: discordgo.ApplicationCommandOptionSubCommand}}}}}
			if err := r.Dispatch(dk.NewContext(session, event)); err != nil {
				t.Fatal(err)
			}
			if name == "form" && response.Type != discordgo.InteractionResponseModal {
				t.Fatal("expected modal")
			}
			if name != "form" && response.Type != discordgo.InteractionResponseChannelMessageWithSource {
				t.Fatal("expected message")
			}
		})
	}
}
