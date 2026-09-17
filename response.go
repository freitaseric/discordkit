package discordkit

import (
	"github.com/bwmarrin/discordgo"
)

// The response lifecycle tracks interaction acknowledgement so that a second
// initial response fails loudly instead of silently producing an API error.
//
// Allowed transitions:
//
//      pending  --Reply/Ephemeral/Update--> sent
//      pending  --Defer/DeferUpdate-------> deferred
//      sent     --Followup/Edit/Delete----> sent
//      deferred --Followup/Edit/Delete----> deferred
//
// ShowModal and Autocomplete are terminal initial responses and require a
// pending interaction of the correct type.

// transition atomically moves the response state, returning an error when the
// interaction was already acknowledged.
func (c *Context) transition(to responseState) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state != responsePending {
		return ErrAlreadyResponded
	}
	c.state = to
	return nil
}

func (c *Context) requireAcknowledged() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state == responsePending {
		return ErrNotResponded
	}
	return nil
}

// Reply sends an initial response as a new message. It fails with
// ErrAlreadyResponded if the interaction was already acknowledged.
func (c *Context) Reply(spec MessageSpec) error {
	data, err := spec.InteractionResponseData()
	if err != nil {
		return err
	}
	if err := c.transition(responseSent); err != nil {
		return err
	}
	return c.Session.InteractionRespond(c.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: data,
	}, discordgo.WithContext(c.Context()))
}

// ReplyText is a convenience for a plain text initial response.
func (c *Context) ReplyText(content string) error {
	return c.Reply(MessageSpec{Content: content})
}

// Ephemeral sends an initial response visible only to the invoking user.
func (c *Context) Ephemeral(spec MessageSpec) error {
	spec.Ephemeral = true
	return c.Reply(spec)
}

// EphemeralText is a convenience for a plain text ephemeral response.
func (c *Context) EphemeralText(content string) error {
	return c.Ephemeral(MessageSpec{Content: content})
}

// Defer acknowledges a command interaction and shows a loading state. Edit the
// original response later with Edit. Pass ephemeral to hide the loading state.
func (c *Context) Defer(ephemeral bool) error {
	if err := c.transition(responseDeferred); err != nil {
		return err
	}
	var flags discordgo.MessageFlags
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}
	return c.Session.InteractionRespond(c.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: flags},
	}, discordgo.WithContext(c.Context()))
}

// DeferUpdate acknowledges a component interaction without changing the message.
// Edit the message later with Edit. Only valid for component interactions.
func (c *Context) DeferUpdate() error {
	if err := c.transition(responseDeferred); err != nil {
		return err
	}
	return c.Session.InteractionRespond(c.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	}, discordgo.WithContext(c.Context()))
}

// Update replaces the message a component belongs to. Only valid as the initial
// response to a component interaction.
func (c *Context) Update(spec MessageSpec) error {
	data, err := spec.InteractionResponseData()
	if err != nil {
		return err
	}
	if err := c.transition(responseSent); err != nil {
		return err
	}
	return c.Session.InteractionRespond(c.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: data,
	}, discordgo.WithContext(c.Context()))
}

// Edit edits the original interaction response after Reply or Defer.
func (c *Context) Edit(spec MessageSpec) (*discordgo.Message, error) {
	if err := c.requireAcknowledged(); err != nil {
		return nil, err
	}
	edit, err := spec.WebhookEdit()
	if err != nil {
		return nil, err
	}
	return c.Session.InteractionResponseEdit(c.Interaction.Interaction, edit, discordgo.WithContext(c.Context()))
}

// Followup sends an additional message after the interaction is acknowledged.
func (c *Context) Followup(spec MessageSpec) (*discordgo.Message, error) {
	if err := c.requireAcknowledged(); err != nil {
		return nil, err
	}
	params, err := spec.WebhookParams()
	if err != nil {
		return nil, err
	}
	return c.Session.FollowupMessageCreate(c.Interaction.Interaction, true, params, discordgo.WithContext(c.Context()))
}

// FollowupEdit edits a previously sent followup message by ID.
func (c *Context) FollowupEdit(messageID string, spec MessageSpec) (*discordgo.Message, error) {
	if err := c.requireAcknowledged(); err != nil {
		return nil, err
	}
	edit, err := spec.WebhookEdit()
	if err != nil {
		return nil, err
	}
	return c.Session.FollowupMessageEdit(c.Interaction.Interaction, messageID, edit, discordgo.WithContext(c.Context()))
}

// DeleteResponse deletes the original interaction response.
func (c *Context) DeleteResponse() error {
	if err := c.requireAcknowledged(); err != nil {
		return err
	}
	return c.Session.InteractionResponseDelete(c.Interaction.Interaction, discordgo.WithContext(c.Context()))
}

// Autocomplete responds to an autocomplete interaction with up to 25 choices.
func (c *Context) Autocomplete(choices ...ChoiceValue) error {
	if c.interaction() == nil || c.Interaction.Type != discordgo.InteractionApplicationCommandAutocomplete {
		return ErrInvalidInteraction
	}
	built, err := buildChoices(choices)
	if err != nil {
		return err
	}
	if err := c.transition(responseSent); err != nil {
		return err
	}
	return c.Session.InteractionRespond(c.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{Choices: built},
	}, discordgo.WithContext(c.Context()))
}

// ShowModal opens a modal as the initial response. Not valid for modal submit
// or autocomplete interactions.
func (c *Context) ShowModal(m Modal) error {
	i := c.interaction()
	if i == nil || i.Type == discordgo.InteractionModalSubmit || i.Type == discordgo.InteractionApplicationCommandAutocomplete {
		return ErrInvalidInteraction
	}
	data, err := m.responseData()
	if err != nil {
		return err
	}
	if err := c.transition(responseSent); err != nil {
		return err
	}
	return c.Session.InteractionRespond(c.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: data,
	}, discordgo.WithContext(c.Context()))
}
