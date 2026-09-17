package discordkit

import (
	"encoding/json"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// Modal is a validated modal ready to be shown with Context.ShowModal.
type Modal struct {
	customID   string
	title      string
	components []discordgo.MessageComponent
}

func (m Modal) responseData() (*discordgo.InteractionResponseData, error) {
	if err := CustomID(m.customID).Validate(); err != nil {
		return nil, err
	}
	if l := len(m.title); l < 1 || l > 45 {
		return nil, fmt.Errorf("%w: modal title must be 1-45 characters", ErrInvalidComponent)
	}
	if err := ValidateModalComponents(m.components); err != nil {
		return nil, err
	}
	return &discordgo.InteractionResponseData{
		CustomID:   m.customID,
		Title:      m.title,
		Components: m.components,
	}, nil
}

// ModalBuilder builds a modal from field components.
type ModalBuilder struct {
	customID string
	title    string
	fields   []Component
}

// Form starts a modal builder using modern Label-wrapped fields.
func Form(customID, title string, fields ...Component) *ModalBuilder {
	return &ModalBuilder{customID: customID, title: title, fields: fields}
}

// Fields appends additional fields to the form.
func (b *ModalBuilder) Fields(fields ...Component) *ModalBuilder {
	b.fields = append(b.fields, fields...)
	return b
}

// Build validates and returns the modal.
func (b *ModalBuilder) Build() (Modal, error) {
	m := Modal{customID: b.customID, title: b.title, components: Components(b.fields...)}
	if _, err := m.responseData(); err != nil {
		return Modal{}, err
	}
	return m, nil
}

// MustBuild is Build with panic on error, for static modal declarations.
func (b *ModalBuilder) MustBuild() Modal {
	m, err := b.Build()
	if err != nil {
		panic(err)
	}
	return m
}

// ModalOf constructs a modal from raw discordgo components as a low-level escape
// hatch when the Form DSL is insufficient.
func ModalOf(customID, title string, components ...discordgo.MessageComponent) Modal {
	return Modal{customID: customID, title: title, components: components}
}

// FieldBuilder wraps a single modal input in a Label.
type FieldBuilder struct {
	label discordgo.Label
	input Component
}

// Field wraps a modal input (text input, string select or file upload) with a
// visible label, producing a discordgo.Label.
func Field(label string, input Component) *FieldBuilder {
	return &FieldBuilder{label: discordgo.Label{Label: label}, input: input}
}

// Description adds helper text below the label.
func (b *FieldBuilder) Description(d string) *FieldBuilder { b.label.Description = d; return b }

func (b *FieldBuilder) component() discordgo.MessageComponent {
	l := b.label
	if b.input != nil {
		l.Component = b.input.component()
	}
	return l
}

// TextInputBuilder builds a text input for a modal field.
type TextInputBuilder struct{ t discordgo.TextInput }

// TextInput creates a short text input. The visible label belongs to the Field.
func TextInput(customID string) *TextInputBuilder {
	return &TextInputBuilder{discordgo.TextInput{CustomID: customID, Style: discordgo.TextInputShort}}
}

// Paragraph switches to a multi-line paragraph style.
func (b *TextInputBuilder) Paragraph() *TextInputBuilder {
	b.t.Style = discordgo.TextInputParagraph
	return b
}

// Placeholder sets placeholder text.
func (b *TextInputBuilder) Placeholder(p string) *TextInputBuilder { b.t.Placeholder = p; return b }

// Value sets a prefilled value.
func (b *TextInputBuilder) Value(v string) *TextInputBuilder { b.t.Value = v; return b }

// Required sets whether the input must be filled.
func (b *TextInputBuilder) Required(required bool) *TextInputBuilder {
	r := required
	b.t.Required = &r
	return b
}

// MinLen sets the minimum length.
func (b *TextInputBuilder) MinLen(n int) *TextInputBuilder { b.t.MinLength = n; return b }

// MaxLen sets the maximum length.
func (b *TextInputBuilder) MaxLen(n int) *TextInputBuilder { b.t.MaxLength = n; return b }

func (b *TextInputBuilder) component() discordgo.MessageComponent { return b.t }

// fileUploadComponent marshals a discordgo.FileUpload together with the
// request-only file_types field that upstream discordgo does not yet model.
// Submissions still decode into discordgo.FileUpload, so no decoder change is
// required. Filtering is by extension, not content validation.
type fileUploadComponent struct {
	upload    discordgo.FileUpload
	fileTypes []string
}

func (f fileUploadComponent) Type() discordgo.ComponentType { return discordgo.FileUploadComponent }

func (f fileUploadComponent) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(f.upload)
	if err != nil {
		return nil, err
	}
	if len(f.fileTypes) == 0 {
		return data, nil
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	types, err := json.Marshal(f.fileTypes)
	if err != nil {
		return nil, err
	}
	m["file_types"] = types
	return json.Marshal(m)
}

// FileUploadBuilder builds a file upload input for a modal field.
type FileUploadBuilder struct{ f fileUploadComponent }

// FileUpload creates a file upload input for a modal field.
func FileUpload(customID string) *FileUploadBuilder {
	return &FileUploadBuilder{fileUploadComponent{upload: discordgo.FileUpload{CustomID: customID}}}
}

// Required sets whether at least one file must be uploaded.
func (b *FileUploadBuilder) Required(required bool) *FileUploadBuilder {
	r := required
	b.f.upload.Required = &r
	return b
}

// MinValues sets the minimum number of files.
func (b *FileUploadBuilder) MinValues(n int) *FileUploadBuilder { b.f.upload.MinValues = &n; return b }

// MaxValues sets the maximum number of files.
func (b *FileUploadBuilder) MaxValues(n int) *FileUploadBuilder { b.f.upload.MaxValues = n; return b }

// FileTypes restricts accepted uploads by extension, e.g. "png", "pdf". This is
// a DiscordKit request-only extension over discordgo's FileUpload.
func (b *FileUploadBuilder) FileTypes(types ...string) *FileUploadBuilder {
	b.f.fileTypes = types
	return b
}

func (b *FileUploadBuilder) component() discordgo.MessageComponent { return b.f }

// FormData provides typed, reflection-free access to a modal submission.
type FormData struct {
	texts    map[string]string
	values   map[string][]string
	resolved *discordgo.ComponentInteractionDataResolved
}

// Form parses the modal submission attached to the context. It returns an empty
// FormData for non-modal interactions.
func (c *Context) Form() *FormData {
	f := &FormData{texts: map[string]string{}, values: map[string][]string{}}
	i := c.interaction()
	if i == nil {
		return f
	}
	var data discordgo.ModalSubmitInteractionData
	switch d := i.Data.(type) {
	case discordgo.ModalSubmitInteractionData:
		data = d
	case *discordgo.ModalSubmitInteractionData:
		if d == nil {
			return f
		}
		data = *d
	default:
		return f
	}
	f.resolved = &data.Resolved
	collectModalValues(data.Components, f)
	return f
}

// collectModalValues walks Label and legacy ActionRow wrappers to index every
// input value by custom ID without reflection.
func collectModalValues(components []discordgo.MessageComponent, f *FormData) {
	for _, c := range components {
		if c == nil {
			continue
		}
		switch v := deref(c).(type) {
		case discordgo.Label:
			if v.Component != nil {
				collectModalValues([]discordgo.MessageComponent{v.Component}, f)
			}
		case discordgo.ActionsRow:
			collectModalValues(v.Components, f)
		case discordgo.TextInput:
			f.texts[v.CustomID] = v.Value
		case discordgo.SelectMenu:
			f.values[v.CustomID] = v.Values
		case discordgo.FileUpload:
			f.values[v.CustomID] = v.Values
		}
	}
}

// String returns a text input value, or the first value of a single-value select.
func (f *FormData) String(customID string) (string, bool) {
	if v, ok := f.texts[customID]; ok {
		return v, true
	}
	if v, ok := f.values[customID]; ok && len(v) > 0 {
		return v[0], true
	}
	return "", false
}

// Strings returns all selected values for a select or file upload input.
func (f *FormData) Strings(customID string) ([]string, bool) {
	if v, ok := f.values[customID]; ok {
		return v, true
	}
	if v, ok := f.texts[customID]; ok {
		return []string{v}, true
	}
	return nil, false
}

// Files returns the resolved attachments uploaded to a file upload input.
func (f *FormData) Files(customID string) ([]*discordgo.MessageAttachment, bool) {
	ids, ok := f.values[customID]
	if !ok {
		return nil, false
	}
	out := make([]*discordgo.MessageAttachment, 0, len(ids))
	if f.resolved != nil {
		for _, id := range ids {
			if a := f.resolved.Attachments[id]; a != nil {
				out = append(out, a)
			}
		}
	}
	return out, true
}

// RequireString returns an error when the field is absent.
func (f *FormData) RequireString(customID string) (string, error) {
	v, ok := f.String(customID)
	if !ok {
		return "", fmt.Errorf("%w: %q", ErrMissingOption, customID)
	}
	return v, nil
}

// MustString panics when the field is absent.
func (f *FormData) MustString(customID string) string {
	v, err := f.RequireString(customID)
	if err != nil {
		panic(err)
	}
	return v
}

// RequireStrings returns an error when the field is absent.
func (f *FormData) RequireStrings(customID string) ([]string, error) {
	v, ok := f.Strings(customID)
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrMissingOption, customID)
	}
	return v, nil
}

// RequireFiles returns an error when the field is absent.
func (f *FormData) RequireFiles(customID string) ([]*discordgo.MessageAttachment, error) {
	v, ok := f.Files(customID)
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrMissingOption, customID)
	}
	return v, nil
}
