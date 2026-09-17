package discordkit

import (
	"github.com/bwmarrin/discordgo"
)

// Component is a DSL node that finalizes into a concrete discordgo component.
// Every builder in this package implements it, and Raw wraps an existing
// discordgo component as an escape hatch.
type Component interface {
	component() discordgo.MessageComponent
}

// Components finalizes DSL nodes into a discordgo component slice, ready for a
// MessageSpec. Validation happens when the MessageSpec is resolved.
func Components(nodes ...Component) []discordgo.MessageComponent {
	out := make([]discordgo.MessageComponent, 0, len(nodes))
	for _, n := range nodes {
		if n == nil {
			continue
		}
		out = append(out, n.component())
	}
	return out
}

// rawComponent is the escape hatch wrapper produced by Raw.
type rawComponent struct{ c discordgo.MessageComponent }

func (r rawComponent) component() discordgo.MessageComponent { return r.c }

// Raw lets callers drop a discordgo component directly into the DSL, so an
// unsupported Discord feature never blocks use of DiscordKit.
func Raw(c discordgo.MessageComponent) Component { return rawComponent{c} }

// TextDisplayBuilder builds a Components V2 text display.
type TextDisplayBuilder struct{ t discordgo.TextDisplay }

// Text creates a text display component.
func Text(content string) *TextDisplayBuilder {
	return &TextDisplayBuilder{discordgo.TextDisplay{Content: content}}
}
func (b *TextDisplayBuilder) component() discordgo.MessageComponent { return b.t }

// RowBuilder builds an action row.
type RowBuilder struct{ children []Component }

// Row groups interactive components. A row with a select must contain only it.
func Row(children ...Component) *RowBuilder { return &RowBuilder{children} }
func (b *RowBuilder) component() discordgo.MessageComponent {
	return discordgo.ActionsRow{Components: Components(b.children...)}
}

// ButtonBuilder builds an interactive, link or premium button.
type ButtonBuilder struct{ b discordgo.Button }

// Button creates an interactive button with the given custom ID.
func Button(label, customID string) *ButtonBuilder {
	return &ButtonBuilder{discordgo.Button{Label: label, CustomID: customID, Style: discordgo.PrimaryButton}}
}

// LinkButton creates a link button that opens a URL.
func LinkButton(label, url string) *ButtonBuilder {
	return &ButtonBuilder{discordgo.Button{Label: label, URL: url, Style: discordgo.LinkButton}}
}

// PremiumButton creates a premium upsell button for a SKU.
func PremiumButton(skuID string) *ButtonBuilder {
	return &ButtonBuilder{discordgo.Button{SKUID: skuID, Style: discordgo.PremiumButton}}
}

// Primary, Secondary, Success and Danger set an interactive button's style.
func (b *ButtonBuilder) Primary() *ButtonBuilder   { b.b.Style = discordgo.PrimaryButton; return b }
func (b *ButtonBuilder) Secondary() *ButtonBuilder { b.b.Style = discordgo.SecondaryButton; return b }
func (b *ButtonBuilder) Success() *ButtonBuilder   { b.b.Style = discordgo.SuccessButton; return b }
func (b *ButtonBuilder) Danger() *ButtonBuilder    { b.b.Style = discordgo.DangerButton; return b }

// Emoji sets a button emoji.
func (b *ButtonBuilder) Emoji(e discordgo.ComponentEmoji) *ButtonBuilder { b.b.Emoji = &e; return b }

// Disabled disables the button.
func (b *ButtonBuilder) Disabled() *ButtonBuilder { b.b.Disabled = true; return b }

func (b *ButtonBuilder) component() discordgo.MessageComponent { return b.b }

// SelectOptionBuilder builds a string select option.
type SelectOptionBuilder struct{ o discordgo.SelectMenuOption }

// SelectOption creates a string select option.
func SelectOption(label, value string) *SelectOptionBuilder {
	return &SelectOptionBuilder{discordgo.SelectMenuOption{Label: label, Value: value}}
}

// Description sets the option description.
func (b *SelectOptionBuilder) Description(d string) *SelectOptionBuilder {
	b.o.Description = d
	return b
}

// Emoji sets the option emoji.
func (b *SelectOptionBuilder) Emoji(e discordgo.ComponentEmoji) *SelectOptionBuilder {
	b.o.Emoji = &e
	return b
}

// Default marks the option selected by default.
func (b *SelectOptionBuilder) Default() *SelectOptionBuilder { b.o.Default = true; return b }

// StringSelectBuilder builds a string select menu with manual options.
type StringSelectBuilder struct{ m discordgo.SelectMenu }

// StringSelect creates a string select menu.
func StringSelect(customID string) *StringSelectBuilder {
	return &StringSelectBuilder{discordgo.SelectMenu{MenuType: discordgo.StringSelectMenu, CustomID: customID}}
}

// Options sets the selectable options. Only string selects accept manual options.
func (b *StringSelectBuilder) Options(options ...*SelectOptionBuilder) *StringSelectBuilder {
	b.m.Options = make([]discordgo.SelectMenuOption, len(options))
	for i, o := range options {
		b.m.Options[i] = o.o
	}
	return b
}

// Placeholder sets placeholder text.
func (b *StringSelectBuilder) Placeholder(p string) *StringSelectBuilder {
	b.m.Placeholder = p
	return b
}

// MinValues sets the minimum number of selections.
func (b *StringSelectBuilder) MinValues(n int) *StringSelectBuilder { b.m.MinValues = &n; return b }

// MaxValues sets the maximum number of selections.
func (b *StringSelectBuilder) MaxValues(n int) *StringSelectBuilder { b.m.MaxValues = n; return b }

// Disabled disables the select menu.
func (b *StringSelectBuilder) Disabled() *StringSelectBuilder { b.m.Disabled = true; return b }

// Required marks the select as required within a modal label.
func (b *StringSelectBuilder) Required() *StringSelectBuilder {
	req := true
	b.m.Required = &req
	return b
}

func (b *StringSelectBuilder) component() discordgo.MessageComponent { return b.m }

// EntitySelectBuilder builds a user, role or mentionable auto-populated select.
type EntitySelectBuilder struct{ m discordgo.SelectMenu }

// UserSelect creates a user select menu.
func UserSelect(customID string) *EntitySelectBuilder {
	return &EntitySelectBuilder{discordgo.SelectMenu{MenuType: discordgo.UserSelectMenu, CustomID: customID}}
}

// RoleSelect creates a role select menu.
func RoleSelect(customID string) *EntitySelectBuilder {
	return &EntitySelectBuilder{discordgo.SelectMenu{MenuType: discordgo.RoleSelectMenu, CustomID: customID}}
}

// MentionableSelect creates a mentionable (user or role) select menu.
func MentionableSelect(customID string) *EntitySelectBuilder {
	return &EntitySelectBuilder{discordgo.SelectMenu{MenuType: discordgo.MentionableSelectMenu, CustomID: customID}}
}

// Placeholder sets placeholder text.
func (b *EntitySelectBuilder) Placeholder(p string) *EntitySelectBuilder {
	b.m.Placeholder = p
	return b
}

// MinValues sets the minimum number of selections.
func (b *EntitySelectBuilder) MinValues(n int) *EntitySelectBuilder { b.m.MinValues = &n; return b }

// MaxValues sets the maximum number of selections.
func (b *EntitySelectBuilder) MaxValues(n int) *EntitySelectBuilder { b.m.MaxValues = n; return b }

// Disabled disables the select menu.
func (b *EntitySelectBuilder) Disabled() *EntitySelectBuilder { b.m.Disabled = true; return b }

// DefaultValues sets auto-populated default values.
func (b *EntitySelectBuilder) DefaultValues(values ...discordgo.SelectMenuDefaultValue) *EntitySelectBuilder {
	b.m.DefaultValues = values
	return b
}

func (b *EntitySelectBuilder) component() discordgo.MessageComponent { return b.m }

// ChannelSelectBuilder builds a channel select menu.
type ChannelSelectBuilder struct{ m discordgo.SelectMenu }

// ChannelSelect creates a channel select menu.
func ChannelSelect(customID string) *ChannelSelectBuilder {
	return &ChannelSelectBuilder{discordgo.SelectMenu{MenuType: discordgo.ChannelSelectMenu, CustomID: customID}}
}

// Placeholder sets placeholder text.
func (b *ChannelSelectBuilder) Placeholder(p string) *ChannelSelectBuilder {
	b.m.Placeholder = p
	return b
}

// MinValues sets the minimum number of selections.
func (b *ChannelSelectBuilder) MinValues(n int) *ChannelSelectBuilder { b.m.MinValues = &n; return b }

// MaxValues sets the maximum number of selections.
func (b *ChannelSelectBuilder) MaxValues(n int) *ChannelSelectBuilder { b.m.MaxValues = n; return b }

// Disabled disables the select menu.
func (b *ChannelSelectBuilder) Disabled() *ChannelSelectBuilder { b.m.Disabled = true; return b }

// ChannelTypes restricts selectable channel types.
func (b *ChannelSelectBuilder) ChannelTypes(types ...discordgo.ChannelType) *ChannelSelectBuilder {
	b.m.ChannelTypes = types
	return b
}

// DefaultValues sets auto-populated default channels.
func (b *ChannelSelectBuilder) DefaultValues(values ...discordgo.SelectMenuDefaultValue) *ChannelSelectBuilder {
	b.m.DefaultValues = values
	return b
}

func (b *ChannelSelectBuilder) component() discordgo.MessageComponent { return b.m }

// SectionBuilder builds a section: 1-3 text displays plus an accessory.
type SectionBuilder struct {
	texts     []Component
	accessory Component
}

// Section groups text displays with a button or thumbnail accessory.
func Section(texts ...Component) *SectionBuilder { return &SectionBuilder{texts: texts} }

// Accessory sets the section accessory (a button or thumbnail).
func (b *SectionBuilder) Accessory(a Component) *SectionBuilder { b.accessory = a; return b }

func (b *SectionBuilder) component() discordgo.MessageComponent {
	s := discordgo.Section{Components: Components(b.texts...)}
	if b.accessory != nil {
		s.Accessory = b.accessory.component()
	}
	return s
}

// ThumbnailBuilder builds a thumbnail accessory.
type ThumbnailBuilder struct{ t discordgo.Thumbnail }

// Thumbnail creates a thumbnail from a media URL.
func Thumbnail(url string) *ThumbnailBuilder {
	return &ThumbnailBuilder{discordgo.Thumbnail{Media: discordgo.UnfurledMediaItem{URL: url}}}
}

// Description sets alt text.
func (b *ThumbnailBuilder) Description(d string) *ThumbnailBuilder { b.t.Description = &d; return b }

// Spoiler marks the thumbnail as a spoiler.
func (b *ThumbnailBuilder) Spoiler() *ThumbnailBuilder { b.t.Spoiler = true; return b }

func (b *ThumbnailBuilder) component() discordgo.MessageComponent { return b.t }

// GalleryItem is a single media gallery entry.
type GalleryItem struct{ item discordgo.MediaGalleryItem }

// MediaItem creates a gallery item from a media URL.
func MediaItem(url string) GalleryItem {
	return GalleryItem{discordgo.MediaGalleryItem{Media: discordgo.UnfurledMediaItem{URL: url}}}
}

// Description sets alt text on a gallery item.
func (g GalleryItem) Description(d string) GalleryItem { g.item.Description = &d; return g }

// Spoiler marks a gallery item as a spoiler.
func (g GalleryItem) Spoiler() GalleryItem { g.item.Spoiler = true; return g }

// GalleryBuilder builds a media gallery.
type GalleryBuilder struct{ items []GalleryItem }

// Gallery creates a media gallery of 1-10 items.
func Gallery(items ...GalleryItem) *GalleryBuilder { return &GalleryBuilder{items} }

func (b *GalleryBuilder) component() discordgo.MessageComponent {
	items := make([]discordgo.MediaGalleryItem, len(b.items))
	for i, it := range b.items {
		items[i] = it.item
	}
	return discordgo.MediaGallery{Items: items}
}

// FileBuilder builds a file display component.
type FileBuilder struct{ f discordgo.FileComponent }

// File displays an uploaded file referenced by attachment://name.
func File(url string) *FileBuilder {
	return &FileBuilder{discordgo.FileComponent{File: discordgo.UnfurledMediaItem{URL: url}}}
}

// Spoiler marks the file as a spoiler.
func (b *FileBuilder) Spoiler() *FileBuilder { b.f.Spoiler = true; return b }

func (b *FileBuilder) component() discordgo.MessageComponent { return b.f }

// SeparatorBuilder builds a separator.
type SeparatorBuilder struct{ s discordgo.Separator }

// Separator creates a visual separator.
func Separator() *SeparatorBuilder { return &SeparatorBuilder{} }

// Divider toggles the visible divider line.
func (b *SeparatorBuilder) Divider(on bool) *SeparatorBuilder { b.s.Divider = &on; return b }

// Large sets large spacing; the default spacing is small.
func (b *SeparatorBuilder) Large() *SeparatorBuilder {
	sz := discordgo.SeparatorSpacingSizeLarge
	b.s.Spacing = &sz
	return b
}

func (b *SeparatorBuilder) component() discordgo.MessageComponent { return b.s }

// ContainerBuilder builds a container with an optional accent color.
type ContainerBuilder struct {
	children []Component
	accent   *int
	spoiler  bool
}

// Container groups components with a colored side bar.
func Container(children ...Component) *ContainerBuilder { return &ContainerBuilder{children: children} }

// AccentColor sets the container accent color.
func (b *ContainerBuilder) AccentColor(color int) *ContainerBuilder { b.accent = &color; return b }

// Spoiler marks the container as a spoiler.
func (b *ContainerBuilder) Spoiler() *ContainerBuilder { b.spoiler = true; return b }

func (b *ContainerBuilder) component() discordgo.MessageComponent {
	return discordgo.Container{
		Components:  Components(b.children...),
		AccentColor: b.accent,
		Spoiler:     b.spoiler,
	}
}
