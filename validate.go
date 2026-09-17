package discordkit

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// componentContext identifies where a component sits in a Discord component tree.
// Discord permits different component types in different structural positions.
type componentContext uint8

const (
	ctxMessageRoot componentContext = iota
	ctxActionRow
	ctxContainer
	ctxSection
	ctxSectionAccessory
	ctxModalRoot
	ctxLabel
)

// maxMessageComponents is Discord's total component budget for a single message,
// counted recursively across every nested layout component.
const maxMessageComponents = 40

func componentErr(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidComponent, fmt.Sprintf(format, args...))
}

// deref unwraps a pointer component to its concrete value form so that callers
// may register builders as either values or pointers without changing behavior.
func deref(c discordgo.MessageComponent) discordgo.MessageComponent {
	switch v := c.(type) {
	case *discordgo.ActionsRow:
		if v != nil {
			return *v
		}
	case *discordgo.Button:
		if v != nil {
			return *v
		}
	case *discordgo.SelectMenu:
		if v != nil {
			return *v
		}
	case *discordgo.TextInput:
		if v != nil {
			return *v
		}
	case *discordgo.Section:
		if v != nil {
			return *v
		}
	case *discordgo.TextDisplay:
		if v != nil {
			return *v
		}
	case *discordgo.Thumbnail:
		if v != nil {
			return *v
		}
	case *discordgo.MediaGallery:
		if v != nil {
			return *v
		}
	case *discordgo.FileComponent:
		if v != nil {
			return *v
		}
	case *discordgo.Separator:
		if v != nil {
			return *v
		}
	case *discordgo.Container:
		if v != nil {
			return *v
		}
	case *discordgo.Label:
		if v != nil {
			return *v
		}
	case *discordgo.FileUpload:
		if v != nil {
			return *v
		}
	}
	return c
}

// isV2Component reports whether a component type belongs to Components V2.
// Their presence forces MessageFlagsIsComponentsV2 and disables legacy content.
func isV2Component(t discordgo.ComponentType) bool {
	switch t {
	case discordgo.SectionComponent, discordgo.TextDisplayComponent, discordgo.ThumbnailComponent,
		discordgo.MediaGalleryComponent, discordgo.FileComponentType, discordgo.SeparatorComponent,
		discordgo.ContainerComponent:
		return true
	}
	return false
}

// usesComponentsV2 reports whether any component in the tree is a V2 component.
func usesComponentsV2(components []discordgo.MessageComponent) bool {
	for _, c := range components {
		if c == nil {
			continue
		}
		switch v := deref(c).(type) {
		case discordgo.ActionsRow:
			if usesComponentsV2(v.Components) {
				return true
			}
		case discordgo.Section:
			return true
		case discordgo.Container:
			return true
		default:
			if isV2Component(v.Type()) {
				return true
			}
		}
	}
	return false
}

// countComponents totals every component in the tree, matching how Discord
// enforces the per-message limit across nested layout components.
func countComponents(components []discordgo.MessageComponent) int {
	total := 0
	for _, c := range components {
		if c == nil {
			continue
		}
		total++
		switch v := deref(c).(type) {
		case discordgo.ActionsRow:
			total += countComponents(v.Components)
		case discordgo.Section:
			total += countComponents(v.Components)
			if v.Accessory != nil {
				total += countComponents([]discordgo.MessageComponent{v.Accessory})
			}
		case discordgo.Container:
			total += countComponents(v.Components)
		case discordgo.Label:
			if v.Component != nil {
				total += countComponents([]discordgo.MessageComponent{v.Component})
			}
		}
	}
	return total
}

func isSelect(t discordgo.ComponentType) bool {
	switch t {
	case discordgo.SelectMenuComponent, discordgo.UserSelectMenuComponent, discordgo.RoleSelectMenuComponent,
		discordgo.MentionableSelectMenuComponent, discordgo.ChannelSelectMenuComponent:
		return true
	}
	return false
}

// ValidateMessageComponents checks a message component tree against Discord's
// structural rules and the 40-component limit. It returns ErrInvalidComponent.
func ValidateMessageComponents(components []discordgo.MessageComponent) error {
	if n := countComponents(components); n > maxMessageComponents {
		return componentErr("message has %d components, limit is %d", n, maxMessageComponents)
	}
	return validateTree(components, ctxMessageRoot)
}

// ValidateModalComponents checks a modal component tree. Modern modals use Label
// wrappers around a single text input, string select or file upload.
func ValidateModalComponents(components []discordgo.MessageComponent) error {
	if len(components) == 0 {
		return componentErr("modal requires at least one component")
	}
	return validateTree(components, ctxModalRoot)
}

func validateTree(components []discordgo.MessageComponent, ctx componentContext) error {
	for _, c := range components {
		if c == nil {
			return componentErr("nil component in %s", contextName(ctx))
		}
		if err := validateComponent(c, ctx); err != nil {
			return err
		}
	}
	return nil
}

func contextName(ctx componentContext) string {
	switch ctx {
	case ctxMessageRoot:
		return "message"
	case ctxActionRow:
		return "action row"
	case ctxContainer:
		return "container"
	case ctxSection:
		return "section"
	case ctxSectionAccessory:
		return "section accessory"
	case ctxModalRoot:
		return "modal"
	case ctxLabel:
		return "label"
	}
	return "component tree"
}

func validateComponent(c discordgo.MessageComponent, ctx componentContext) error {
	switch v := deref(c).(type) {
	case discordgo.ActionsRow:
		return validateActionRow(v, ctx)
	case discordgo.Button:
		return validateButton(v, ctx)
	case discordgo.SelectMenu:
		return validateSelect(v, ctx)
	case discordgo.TextInput:
		if ctx != ctxLabel && ctx != ctxActionRow {
			return componentErr("text input is only valid inside a label or action row")
		}
		return nil
	case discordgo.TextDisplay:
		if ctx != ctxMessageRoot && ctx != ctxContainer && ctx != ctxSection {
			return componentErr("text display is not valid in a %s", contextName(ctx))
		}
		return nil
	case discordgo.Section:
		return validateSection(v, ctx)
	case discordgo.Thumbnail:
		if ctx != ctxSectionAccessory {
			return componentErr("thumbnail is only valid as a section accessory")
		}
		return nil
	case discordgo.MediaGallery:
		if ctx != ctxMessageRoot && ctx != ctxContainer {
			return componentErr("media gallery is not valid in a %s", contextName(ctx))
		}
		if len(v.Items) < 1 || len(v.Items) > 10 {
			return componentErr("media gallery requires 1-10 items, got %d", len(v.Items))
		}
		return nil
	case discordgo.FileComponent:
		if ctx != ctxMessageRoot && ctx != ctxContainer {
			return componentErr("file is not valid in a %s", contextName(ctx))
		}
		return nil
	case discordgo.Separator:
		if ctx != ctxMessageRoot && ctx != ctxContainer {
			return componentErr("separator is not valid in a %s", contextName(ctx))
		}
		return nil
	case discordgo.Container:
		return validateContainer(v, ctx)
	case discordgo.Label:
		return validateLabel(v, ctx)
	case discordgo.FileUpload:
		if ctx != ctxLabel {
			return componentErr("file upload is only valid inside a label")
		}
		return validateFileUpload(v)
	case fileUploadComponent:
		if ctx != ctxLabel {
			return componentErr("file upload is only valid inside a label")
		}
		return validateFileUpload(v.upload)
	}
	return componentErr("unsupported component type %d", c.Type())
}

func validateActionRow(r discordgo.ActionsRow, ctx componentContext) error {
	if ctx != ctxMessageRoot && ctx != ctxContainer && ctx != ctxModalRoot {
		return componentErr("action row is not valid in a %s", contextName(ctx))
	}
	if len(r.Components) == 0 {
		return componentErr("action row must contain at least one component")
	}
	hasSelect := false
	for _, child := range r.Components {
		if child == nil {
			return componentErr("nil component in action row")
		}
		if isSelect(child.Type()) {
			hasSelect = true
		}
	}
	if hasSelect && len(r.Components) != 1 {
		return componentErr("an action row with a select menu must contain only that select")
	}
	if !hasSelect && len(r.Components) > 5 {
		return componentErr("action row allows at most 5 buttons, got %d", len(r.Components))
	}
	return validateTree(r.Components, ctxActionRow)
}

func validateButton(b discordgo.Button, ctx componentContext) error {
	if ctx != ctxActionRow && ctx != ctxSectionAccessory {
		return componentErr("button must be in an action row or a section accessory")
	}
	switch b.Style {
	case discordgo.LinkButton:
		if b.URL == "" {
			return componentErr("link button requires a url")
		}
		if b.CustomID != "" || b.SKUID != "" {
			return componentErr("link button must not set custom_id or sku_id")
		}
	case discordgo.PremiumButton:
		if b.SKUID == "" {
			return componentErr("premium button requires a sku_id")
		}
		if b.CustomID != "" || b.URL != "" || b.Label != "" || b.Emoji != nil {
			return componentErr("premium button must not set custom_id, url, label or emoji")
		}
	default:
		if b.CustomID == "" {
			return componentErr("interactive button requires a custom_id")
		}
		if b.URL != "" || b.SKUID != "" {
			return componentErr("interactive button must not set url or sku_id")
		}
		if err := CustomID(b.CustomID).Validate(); err != nil {
			return err
		}
		if b.Label == "" && b.Emoji == nil {
			return componentErr("interactive button requires a label or emoji")
		}
	}
	return nil
}

func validateSelect(s discordgo.SelectMenu, ctx componentContext) error {
	if ctx != ctxActionRow && ctx != ctxLabel {
		return componentErr("select menu must be in an action row or a label")
	}
	if s.CustomID == "" {
		return componentErr("select menu requires a custom_id")
	}
	if err := CustomID(s.CustomID).Validate(); err != nil {
		return err
	}
	if s.MinValues != nil && (*s.MinValues < 0 || *s.MinValues > 25) {
		return componentErr("select menu min_values must be within 0-25")
	}
	if s.MaxValues < 0 || s.MaxValues > 25 {
		return componentErr("select menu max_values must be within 0-25")
	}
	if s.MinValues != nil && s.MaxValues != 0 && *s.MinValues > s.MaxValues {
		return componentErr("select menu min_values exceeds max_values")
	}
	stringMenu := s.MenuType == discordgo.StringSelectMenu || s.MenuType == 0
	if stringMenu {
		if len(s.Options) < 1 || len(s.Options) > 25 {
			return componentErr("string select requires 1-25 options, got %d", len(s.Options))
		}
	} else if len(s.Options) != 0 {
		return componentErr("only string selects may declare options")
	}
	if s.MenuType != discordgo.ChannelSelectMenu && len(s.ChannelTypes) != 0 {
		return componentErr("only channel selects may declare channel_types")
	}
	return nil
}

func validateSection(s discordgo.Section, ctx componentContext) error {
	if ctx != ctxMessageRoot && ctx != ctxContainer {
		return componentErr("section is not valid in a %s", contextName(ctx))
	}
	if len(s.Components) < 1 || len(s.Components) > 3 {
		return componentErr("section requires 1-3 text displays, got %d", len(s.Components))
	}
	for _, child := range s.Components {
		if child == nil || child.Type() != discordgo.TextDisplayComponent {
			return componentErr("section components must all be text displays")
		}
	}
	if s.Accessory == nil {
		return componentErr("section requires an accessory")
	}
	switch s.Accessory.Type() {
	case discordgo.ButtonComponent, discordgo.ThumbnailComponent:
		return validateComponent(s.Accessory, ctxSectionAccessory)
	}
	return componentErr("section accessory must be a button or thumbnail")
}

func validateContainer(c discordgo.Container, ctx componentContext) error {
	if ctx != ctxMessageRoot {
		return componentErr("container may only appear at the message root")
	}
	if len(c.Components) == 0 {
		return componentErr("container must contain at least one component")
	}
	return validateTree(c.Components, ctxContainer)
}

func validateLabel(l discordgo.Label, ctx componentContext) error {
	if ctx != ctxModalRoot {
		return componentErr("label is only valid at the modal root")
	}
	if l.Label == "" {
		return componentErr("label requires text")
	}
	if l.Component == nil {
		return componentErr("label requires a component")
	}
	switch l.Component.Type() {
	case discordgo.TextInputComponent, discordgo.FileUploadComponent, discordgo.SelectMenuComponent,
		discordgo.UserSelectMenuComponent, discordgo.RoleSelectMenuComponent,
		discordgo.MentionableSelectMenuComponent, discordgo.ChannelSelectMenuComponent:
		return validateComponent(l.Component, ctxLabel)
	}
	return componentErr("label component must be a text input, select menu or file upload")
}

func validateFileUpload(f discordgo.FileUpload) error {
	if f.CustomID == "" {
		return componentErr("file upload requires a custom_id")
	}
	if err := CustomID(f.CustomID).Validate(); err != nil {
		return err
	}
	if f.MinValues != nil && (*f.MinValues < 0 || *f.MinValues > 10) {
		return componentErr("file upload min_values must be within 0-10")
	}
	if f.MaxValues < 0 || f.MaxValues > 10 {
		return componentErr("file upload max_values must be within 0-10")
	}
	if f.MinValues != nil && f.MaxValues != 0 && *f.MinValues > f.MaxValues {
		return componentErr("file upload min_values exceeds max_values")
	}
	return nil
}
