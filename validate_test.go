package discordkit

import (
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func mustInvalid(t *testing.T, err error, name string) {
	t.Helper()
	if !errors.Is(err, ErrInvalidComponent) && !errors.Is(err, ErrInvalidCustomID) {
		t.Fatalf("%s: expected invalid component error, got %v", name, err)
	}
}

func TestValidateButtons(t *testing.T) {
	if err := ValidateMessageComponents(Components(Row(Button("Save", "save")))); err != nil {
		t.Fatalf("valid button: %v", err)
	}
	if err := ValidateMessageComponents(Components(Row(LinkButton("Open", "https://x")))); err != nil {
		t.Fatalf("valid link button: %v", err)
	}
	if err := ValidateMessageComponents(Components(Row(PremiumButton("sku")))); err != nil {
		t.Fatalf("valid premium button: %v", err)
	}
	// Button at message root (not in a row) is invalid.
	mustInvalid(t, ValidateMessageComponents(Components(Button("x", "y"))), "root button")
	// Link button with custom_id is invalid.
	bad := discordgo.Button{Style: discordgo.LinkButton, URL: "https://x", CustomID: "y"}
	mustInvalid(t, ValidateMessageComponents([]discordgo.MessageComponent{discordgo.ActionsRow{Components: []discordgo.MessageComponent{bad}}}), "link+customid")
	// Premium button with a label is invalid.
	badPrem := discordgo.Button{Style: discordgo.PremiumButton, SKUID: "s", Label: "no"}
	mustInvalid(t, ValidateMessageComponents([]discordgo.MessageComponent{discordgo.ActionsRow{Components: []discordgo.MessageComponent{badPrem}}}), "premium+label")
	// Interactive button without custom_id is invalid.
	mustInvalid(t, ValidateMessageComponents(Components(Row(Button("x", "")))), "no customid")
}

func TestValidateSelects(t *testing.T) {
	sel := StringSelect("s").Options(SelectOption("A", "a"), SelectOption("B", "b"))
	if err := ValidateMessageComponents(Components(Row(sel))); err != nil {
		t.Fatalf("valid string select: %v", err)
	}
	if err := ValidateMessageComponents(Components(Row(UserSelect("u")))); err != nil {
		t.Fatalf("valid user select: %v", err)
	}
	// A select must be alone in its row.
	mustInvalid(t, ValidateMessageComponents(Components(Row(UserSelect("u"), Button("x", "y")))), "select+button")
	// String select with no options is invalid.
	mustInvalid(t, ValidateMessageComponents(Components(Row(StringSelect("s")))), "empty string select")
}

func TestValidateSectionAndContainer(t *testing.T) {
	sec := Section(Text("one"), Text("two")).Accessory(Thumbnail("https://example.com/x.png"))
	if err := ValidateMessageComponents(Components(sec)); err != nil {
		t.Fatalf("valid section: %v", err)
	}
	secBtn := Section(Text("one")).Accessory(Button("Go", "go"))
	if err := ValidateMessageComponents(Components(secBtn)); err != nil {
		t.Fatalf("valid section button accessory: %v", err)
	}
	// Section without accessory is invalid.
	mustInvalid(t, ValidateMessageComponents(Components(Section(Text("one")))), "no accessory")
	// Section with 4 texts is invalid.
	mustInvalid(t, ValidateMessageComponents(Components(Section(Text("1"), Text("2"), Text("3"), Text("4")).Accessory(Thumbnail("u")))), "too many texts")
	// Container of layout components is valid.
	cont := Container(Text("hi"), Separator())
	if err := ValidateMessageComponents(Components(cont)); err != nil {
		t.Fatalf("valid container: %v", err)
	}
	// Nested containers are invalid.
	mustInvalid(t, ValidateMessageComponents(Components(Container(Container(Text("x"))))), "nested container")
	// Thumbnail only allowed as section accessory.
	mustInvalid(t, ValidateMessageComponents(Components(Thumbnail("u"))), "root thumbnail")
}

func TestValidateComponentLimit(t *testing.T) {
	var nodes []Component
	for i := 0; i < 41; i++ {
		nodes = append(nodes, Text("x"))
	}
	mustInvalid(t, ValidateMessageComponents(Components(nodes...)), "over 40 components")
}

func TestValidateModal(t *testing.T) {
	field := Field("Name", TextInput("name"))
	if err := ValidateModalComponents(Components(field)); err != nil {
		t.Fatalf("valid modal: %v", err)
	}
	fileField := Field("Upload", FileUpload("file").FileTypes("png"))
	if err := ValidateModalComponents(Components(fileField)); err != nil {
		t.Fatalf("valid file modal: %v", err)
	}
	// Text input outside a label is invalid at modal root.
	mustInvalid(t, ValidateModalComponents(Components(Text("x"))), "text display in modal")
	// Empty modal is invalid.
	mustInvalid(t, ValidateModalComponents(nil), "empty modal")
}

func TestUsesComponentsV2(t *testing.T) {
	if usesComponentsV2(Components(Row(Button("x", "y")))) {
		t.Fatal("action row of buttons is not V2")
	}
	if !usesComponentsV2(Components(Text("x"))) {
		t.Fatal("text display is V2")
	}
	if !usesComponentsV2(Components(Container(Text("x")))) {
		t.Fatal("container is V2")
	}
}
