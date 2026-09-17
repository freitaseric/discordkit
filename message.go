package discordkit

import (
	"github.com/bwmarrin/discordgo"
)

// MessageSpec is the single source of truth for an outgoing message or
// interaction response. It converts, without duplicating rules, into every
// discordgo payload shape used across the response lifecycle.
//
// When Components contains Components V2 layout components, MessageSpec applies
// the IS_COMPONENTS_V2 flag, converts Content into a leading TextDisplay, and
// rejects payloads that Discord forbids (legacy embeds and polls).
type MessageSpec struct {
	Content         string
	Components      []discordgo.MessageComponent
	Embeds          []*discordgo.MessageEmbed
	Files           []*discordgo.File
	Attachments     []*discordgo.MessageAttachment
	AllowedMentions *discordgo.MessageAllowedMentions
	Poll            *discordgo.Poll
	Ephemeral       bool
}

// resolvedMessage is a MessageSpec after validation and Components V2
// normalization. Its fields map directly onto discordgo payload structs.
type resolvedMessage struct {
	content         string
	components      []discordgo.MessageComponent
	embeds          []*discordgo.MessageEmbed
	files           []*discordgo.File
	attachments     []*discordgo.MessageAttachment
	allowedMentions *discordgo.MessageAllowedMentions
	poll            *discordgo.Poll
	flags           discordgo.MessageFlags
	v2              bool
}

// resolve validates the component tree and applies Components V2 normalization.
// It never mutates the receiver's slices.
func (m MessageSpec) resolve() (*resolvedMessage, error) {
	if err := ValidateMessageComponents(m.Components); err != nil {
		return nil, err
	}
	r := &resolvedMessage{
		content:         m.Content,
		components:      m.Components,
		embeds:          m.Embeds,
		files:           m.Files,
		attachments:     m.Attachments,
		allowedMentions: m.AllowedMentions,
		poll:            m.Poll,
		v2:              usesComponentsV2(m.Components),
	}
	if r.v2 {
		if len(m.Embeds) > 0 {
			return nil, componentErr("embeds cannot be combined with Components V2")
		}
		if m.Poll != nil {
			return nil, componentErr("polls cannot be combined with Components V2")
		}
		if m.Content != "" {
			text := discordgo.TextDisplay{Content: m.Content}
			r.components = append([]discordgo.MessageComponent{text}, m.Components...)
			r.content = ""
		}
		r.flags |= discordgo.MessageFlagsIsComponentsV2
	}
	if m.Ephemeral {
		r.flags |= discordgo.MessageFlagsEphemeral
	}
	return r, nil
}

// InteractionResponseData converts the spec into interaction response data,
// applying Components V2 normalization and flags.
func (m MessageSpec) InteractionResponseData() (*discordgo.InteractionResponseData, error) {
	r, err := m.resolve()
	if err != nil {
		return nil, err
	}
	data := &discordgo.InteractionResponseData{
		Content:         r.content,
		Components:      r.components,
		Embeds:          r.embeds,
		Files:           r.files,
		AllowedMentions: r.allowedMentions,
		Poll:            r.poll,
		Flags:           r.flags,
	}
	if r.attachments != nil {
		attachments := r.attachments
		data.Attachments = &attachments
	}
	return data, nil
}

// WebhookParams converts the spec into followup webhook params. Discord's
// webhook create endpoint has no poll field, so a poll is rejected here.
func (m MessageSpec) WebhookParams() (*discordgo.WebhookParams, error) {
	r, err := m.resolve()
	if err != nil {
		return nil, err
	}
	if r.poll != nil {
		return nil, componentErr("polls are not supported in followup messages")
	}
	return &discordgo.WebhookParams{
		Content:         r.content,
		Components:      r.components,
		Embeds:          r.embeds,
		Files:           r.files,
		Attachments:     r.attachments,
		AllowedMentions: r.allowedMentions,
		Flags:           r.flags,
	}, nil
}

// WebhookEdit converts the spec into a webhook edit payload used to edit the
// original interaction response or a followup message.
func (m MessageSpec) WebhookEdit() (*discordgo.WebhookEdit, error) {
	r, err := m.resolve()
	if err != nil {
		return nil, err
	}
	if r.poll != nil {
		return nil, componentErr("polls cannot be edited into a message")
	}
	edit := &discordgo.WebhookEdit{
		Components:      &r.components,
		Files:           r.files,
		AllowedMentions: r.allowedMentions,
		Flags:           r.flags,
	}
	// A V2 message has no separate content field; content lives in a TextDisplay.
	if !r.v2 {
		edit.Content = &r.content
		edit.Embeds = &r.embeds
	}
	if r.attachments != nil {
		attachments := r.attachments
		edit.Attachments = &attachments
	}
	return edit, nil
}

// MessageEdit converts the spec into a channel message edit for the given
// channel and message IDs.
func (m MessageSpec) MessageEdit(channelID, messageID string) (*discordgo.MessageEdit, error) {
	r, err := m.resolve()
	if err != nil {
		return nil, err
	}
	if r.poll != nil {
		return nil, componentErr("polls cannot be edited into a message")
	}
	edit := &discordgo.MessageEdit{
		ID:              messageID,
		Channel:         channelID,
		Components:      &r.components,
		Files:           r.files,
		AllowedMentions: r.allowedMentions,
		Flags:           r.flags,
	}
	if !r.v2 {
		edit.Content = &r.content
		edit.Embeds = &r.embeds
	}
	if r.attachments != nil {
		attachments := r.attachments
		edit.Attachments = &attachments
	}
	return edit, nil
}
