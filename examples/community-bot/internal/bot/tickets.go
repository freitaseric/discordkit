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
