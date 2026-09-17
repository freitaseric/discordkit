package discordkit

import (
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestMessageSpecTraditional(t *testing.T) {
	spec := MessageSpec{
		Content:    "hello",
		Embeds:     []*discordgo.MessageEmbed{{Title: "t"}},
		Ephemeral:  true,
		Components: Components(Row(Button("Save", "save"))),
	}
	data, err := spec.InteractionResponseData()
	if err != nil {
		t.Fatal(err)
	}
	if data.Content != "hello" {
		t.Fatalf("content = %q", data.Content)
	}
	if data.Flags&discordgo.MessageFlagsEphemeral == 0 {
		t.Fatal("expected ephemeral flag")
	}
	if data.Flags&discordgo.MessageFlagsIsComponentsV2 != 0 {
		t.Fatal("traditional message must not set V2 flag")
	}
	if len(data.Embeds) != 1 {
		t.Fatal("embeds dropped")
	}
}

func TestMessageSpecV2Normalization(t *testing.T) {
	spec := MessageSpec{
		Content:    "title text",
		Components: Components(Text("body"), Separator()),
	}
	data, err := spec.InteractionResponseData()
	if err != nil {
		t.Fatal(err)
	}
	if data.Content != "" {
		t.Fatalf("V2 content must be cleared, got %q", data.Content)
	}
	if data.Flags&discordgo.MessageFlagsIsComponentsV2 == 0 {
		t.Fatal("expected V2 flag")
	}
	// The original content must become a leading TextDisplay.
	if len(data.Components) != 3 {
		t.Fatalf("expected 3 components, got %d", len(data.Components))
	}
	td, ok := data.Components[0].(discordgo.TextDisplay)
	if !ok || td.Content != "title text" {
		t.Fatalf("first component should be TextDisplay with original content, got %#v", data.Components[0])
	}
}

func TestMessageSpecV2Incompatibilities(t *testing.T) {
	base := Components(Text("hi"))
	if _, err := (MessageSpec{Components: base, Embeds: []*discordgo.MessageEmbed{{Title: "x"}}}).InteractionResponseData(); !errors.Is(err, ErrInvalidComponent) {
		t.Fatalf("V2 + embeds should fail, got %v", err)
	}
	if _, err := (MessageSpec{Components: base, Poll: &discordgo.Poll{}}).InteractionResponseData(); !errors.Is(err, ErrInvalidComponent) {
		t.Fatalf("V2 + poll should fail, got %v", err)
	}
}

func TestMessageSpecWebhookRejectsPoll(t *testing.T) {
	spec := MessageSpec{Content: "hi", Poll: &discordgo.Poll{}}
	if _, err := spec.WebhookParams(); !errors.Is(err, ErrInvalidComponent) {
		t.Fatalf("followup with poll should fail, got %v", err)
	}
}

func TestMessageSpecConversions(t *testing.T) {
	spec := MessageSpec{Content: "edit me", Components: Components(Row(Button("Go", "go")))}
	edit, err := spec.WebhookEdit()
	if err != nil {
		t.Fatal(err)
	}
	if edit.Content == nil || *edit.Content != "edit me" {
		t.Fatal("webhook edit content")
	}
	me, err := spec.MessageEdit("chan", "msg")
	if err != nil {
		t.Fatal(err)
	}
	if me.ID != "msg" || me.Channel != "chan" || me.Content == nil {
		t.Fatal("message edit fields")
	}
	// V2 edit must not populate a content pointer.
	v2, err := (MessageSpec{Content: "x", Components: Components(Text("y"))}).WebhookEdit()
	if err != nil {
		t.Fatal(err)
	}
	if v2.Content != nil {
		t.Fatal("V2 webhook edit must not set content")
	}
}
