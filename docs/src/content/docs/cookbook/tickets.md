---
title: "06 · Open and track tickets"
description: "Forms, parameterized routes, defer and state changes."
---

**Enable stage 4.** `/ticket new` opens a form and `/ticket show` reads a ticket. Publish a fresh panel to include the new-ticket button.

## Two interactions

The command or button receives `ShowModal` as its initial response. Do not defer first. Submitting the form later creates a separate interaction routed to `/tickets/create`.

`Field` labels wrap text inputs; paragraph and length options guide entry. The store still trims and validates values. `RequireString` detects absent fields. `createTicket` sends `Defer(true)` before disk access and completes it with `Edit`, never a second initial reply.

Discord requires an initial response within 3 seconds and interaction tokens remain valid for 15 minutes. Modal opening and autocomplete do not use the normal message-defer flow. See the [official response lifecycle](https://docs.discord.com/developers/interactions/receiving-and-responding).

## Custom IDs are references

Route builders encode and validate parameters; `RequireParam` reads decoded values. Keep custom IDs within 100 characters and never store secrets or ticket descriptions in them.

The store rechecks guild, actor and current state on every click. `DeferUpdate` acknowledges a private ticket button, and `Edit` refreshes that private message after persistence. The shared public panel is never replaced with ticket content.

<!-- cookbook-source: internal/bot/tickets.go -->
```go title="internal/bot/tickets.go"
package bot

import (
	"strings"

	dk "github.com/freitaseric/discordkit"
)

func (b *Bot) registerTickets(r *dk.Router) func() error {
	return func() error {
		group := r.Group("ticket")
		registrations := []func() error{
			func() error { return group.Command("new", b.newTicket) },
			func() error { return group.Command("show", b.showTicket) },
			func() error { return group.Autocomplete("show", "id", b.completeTicket) },
			func() error { return r.Component("/tickets/new", b.newTicket) },
			func() error { return r.Modal("/tickets/create", b.createTicket) },
			func() error { return r.Component("/tickets/:id/close", b.changeTicket("close")) },
			func() error { return r.Component("/tickets/:id/claim", b.changeTicket("claim")) },
		}
		for _, register := range registrations {
			if err := register(); err != nil {
				return err
			}
		}
		return nil
	}
}
func (b *Bot) newTicket(c *dk.Context) error {
	form, err := dk.Form("/tickets/create", "Novo chamado",
		dk.Field("Assunto", dk.TextInput("subject").Required(true).MinLen(3).MaxLen(80)),
		dk.Field("Descrição", dk.TextInput("description").Paragraph().Required(true).MinLen(10).MaxLen(1000)).Description("Explique como reproduzir. Não inclua senhas."),
	).Build()
	if err != nil {
		return err
	}
	return c.ShowModal(form)
}
func (b *Bot) createTicket(c *dk.Context) error {
	subject, err := c.Form().RequireString("subject")
	if err != nil {
		return err
	}
	description, err := c.Form().RequireString("description")
	if err != nil {
		return err
	}
	if err = c.Defer(true); err != nil {
		return err
	}
	t, err := b.Store.Create(actor(c), c.Interaction.ID, subject, description)
	if err != nil {
		return finish(c, dk.MessageSpec{}, err)
	}
	return finish(c, ticketView(t), nil)
}
func (b *Bot) showTicket(c *dk.Context) error {
	id, err := c.RequireString("id")
	if err != nil {
		return err
	}
	if err = c.Defer(true); err != nil {
		return err
	}
	t, err := b.Store.Get(actor(c), id)
	if err != nil {
		return finish(c, dk.MessageSpec{}, err)
	}
	return finish(c, ticketView(t), nil)
}
func (b *Bot) changeTicket(action string) dk.Handler {
	return func(c *dk.Context) error {
		id, err := c.RequireParam("id")
		if err != nil {
			return err
		}
		// These controls only occur in private ticket responses, never on the public panel.
		if err = c.DeferUpdate(); err != nil {
			return err
		}
		t, err := b.Store.Change(actor(c), id, action, 0)
		if err != nil {
			return finish(c, dk.MessageSpec{}, err)
		}
		return finish(c, ticketView(t), nil)
	}
}
func (b *Bot) completeTicket(c *dk.Context) error {
	focused, ok := c.FocusedOption()
	if !ok {
		return c.Autocomplete()
	}
	query, _ := focused.Value.(string)
	query = strings.ToLower(query)
	choices := []dk.ChoiceValue{}
	for _, t := range b.Store.List(actor(c), "") {
		if strings.Contains(strings.ToLower(t.Subject), query) || strings.Contains(t.ID, query) {
			// IDs are short Discord snowflakes; truncate the title to stay below 100 characters.
			title := []rune(t.Subject)
			for len(string(title)) > 60 {
				title = title[:len(title)-1]
			}
			choices = append(choices, dk.Choice(string(title)+" · "+t.ID, t.ID))
			if len(choices) == 25 {
				break
			}
		}
	}
	return c.Autocomplete(choices...)
}
```
<!-- /cookbook-source -->

A disk failure is reported through `finish` while the underlying error is logged. If persistence succeeds but the Discord response fails, the ticket already exists: inspect it before submitting another form. There is no distributed transaction between the file and Discord.

Old messages are snapshots. A state change does not refresh every previously sent copy, but every new action checks the current store state.

**Verify with three accounts:** create as owner; manually query as an unrelated member and confirm rejection; claim as staff; reject a second staff claim; close as owner or staff; restart and verify the persisted state. Autocomplete must not reveal unauthorized records.

**Exercise:** add a refresh button that reloads and redraws the ticket, retaining authorization on the read.

Next: [queue navigation](/cookbook/queue/).

---

[← Previous: 05 · Build the JSON store](/cookbook/persistence/) · [Next →: 07 · Filters, pages and autocomplete](/cookbook/queue/)
