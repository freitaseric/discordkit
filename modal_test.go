package discordkit

import (
	"encoding/json"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestFormDataCollectsValues(t *testing.T) {
	i := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		Type: discordgo.InteractionModalSubmit,
		Data: discordgo.ModalSubmitInteractionData{
			CustomID: "form",
			Components: []discordgo.MessageComponent{
				discordgo.Label{Component: discordgo.TextInput{CustomID: "name", Value: "Alice"}},
				discordgo.Label{Component: discordgo.SelectMenu{CustomID: "tags", Values: []string{"x", "y"}}},
				discordgo.Label{Component: discordgo.FileUpload{CustomID: "doc", Values: []string{"a1"}}},
			},
			Resolved: discordgo.ComponentInteractionDataResolved{
				Attachments: map[string]*discordgo.MessageAttachment{
					"a1": {ID: "a1", Filename: "report.pdf"},
				},
			},
		},
	}}
	c := NewContext(nil, i)
	f := c.Form()

	if v, ok := f.String("name"); !ok || v != "Alice" {
		t.Fatalf("name: %q %v", v, ok)
	}
	if v, ok := f.Strings("tags"); !ok || len(v) != 2 {
		t.Fatalf("tags: %v %v", v, ok)
	}
	files, ok := f.Files("doc")
	if !ok || len(files) != 1 || files[0].Filename != "report.pdf" {
		t.Fatalf("files: %+v %v", files, ok)
	}
	if _, err := f.RequireString("missing"); err == nil {
		t.Fatal("expected error for missing field")
	}
}

func TestFormDataNonModalIsEmpty(t *testing.T) {
	i := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		Type: discordgo.InteractionApplicationCommand,
	}}
	f := NewContext(nil, i).Form()
	if _, ok := f.String("anything"); ok {
		t.Fatal("expected empty form data")
	}
}

func TestFileUploadMarshalsFileTypes(t *testing.T) {
	fu := FileUpload("doc").FileTypes("png", "pdf").component()
	data, err := json.Marshal(fu)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	raw, ok := m["file_types"]
	if !ok {
		t.Fatalf("file_types missing: %s", data)
	}
	var types []string
	if err := json.Unmarshal(raw, &types); err != nil || len(types) != 2 {
		t.Fatalf("file_types: %v %v", types, err)
	}
}

func TestModalBuildValidates(t *testing.T) {
	_, err := Form("id", "Title", Field("Name", TextInput("name"))).Build()
	if err != nil {
		t.Fatalf("valid modal: %v", err)
	}
	if _, err := Form("id", "", Field("Name", TextInput("name"))).Build(); err == nil {
		t.Fatal("expected error for empty title")
	}
}
