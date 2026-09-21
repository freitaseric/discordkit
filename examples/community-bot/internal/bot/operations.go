package bot

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/bwmarrin/discordgo"
	dk "github.com/freitaseric/discordkit"
)

func (b *Bot) registerOperations(r *dk.Router) func() error {
	return func() error {
		if err := r.Autocomplete("ticket rate", "id", b.completeTicket); err != nil {
			return err
		}
		if err := r.Command("ticket rate", func(c *dk.Context) error {
			id, err := c.RequireString("id")
			if err != nil {
				return err
			}
			score, err := c.RequireInt("score")
			if err != nil {
				return err
			}
			if err = c.Defer(true); err != nil {
				return err
			}
			t, err := b.Store.Change(actor(c), id, "rate", int(score))
			if err != nil {
				return finish(c, dk.MessageSpec{}, err)
			}
			return finish(c, ticketView(t), nil)
		}); err != nil {
			return err
		}
		if err := r.Component("/export/dismiss", func(c *dk.Context) error {
			if err := c.DeferUpdate(); err != nil {
				return err
			}
			return c.DeleteResponse()
		}); err != nil {
			return err
		}
		return r.Command("ticket export", b.export)
	}
}
func (b *Bot) export(c *dk.Context) error {
	if err := c.Defer(true); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(b.Store.List(actor(c), ""), "", "  ")
	if err != nil {
		return finish(c, dk.MessageSpec{}, err)
	}
	if len(raw) > 7*1024*1024 {
		return finish(c, message(dk.Text("Exportação muito grande. Solicite um backup ao operador.")), nil)
	}
	// Complete the deferred response before creating an additional private message.
	if err = finish(c, message(dk.Text("Exportação preparada.")), nil); err != nil {
		return err
	}
	spec := message(dk.Text("## Exportação de chamados"), dk.File("attachment://tickets.json"), dk.Row(dk.Button("Ocultar exportação", "/export/dismiss").Secondary()))
	spec.Ephemeral = true
	spec.Files = []*discordgo.File{{Name: "tickets.json", ContentType: "application/json", Reader: bytes.NewReader(raw)}}
	sent, err := c.Followup(spec)
	if err != nil {
		return err
	}
	// Edit the same followup while retaining its attachment and controls.
	spec.Components = dk.Components(dk.Text(fmt.Sprintf("## Exportação pronta\n%d bytes · somente chamados autorizados.", len(raw))), dk.File("attachment://tickets.json"), dk.Row(dk.Button("Ocultar exportação", "/export/dismiss").Secondary()))
	spec.Ephemeral = false // Editing does not change the existing message visibility.
	spec.Files = nil
	spec.Attachments = sent.Attachments
	_, err = c.FollowupEdit(sent.ID, spec)
	return err
}
