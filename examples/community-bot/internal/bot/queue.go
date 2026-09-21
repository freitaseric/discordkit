package bot

import (
	"strconv"

	dk "github.com/freitaseric/discordkit"
)

func validStatus(s string) bool { return s == "open" || s == "closed" || s == "all" }
func (b *Bot) queue(c *dk.Context, status string, page int) dk.MessageSpec {
	filter := status
	if filter == "all" {
		filter = ""
	}
	return queueView(b.Store.List(actor(c), filter), status, page)
}
func (b *Bot) registerQueue(r *dk.Router) func() error {
	return func() error {
		if err := r.Command("ticket list", func(c *dk.Context) error {
			status, ok := c.String("status")
			if !ok {
				status = "open"
			}
			if !validStatus(status) {
				return c.EphemeralText("Filtro inválido.")
			}
			return c.Ephemeral(b.queue(c, status, 0))
		}); err != nil {
			return err
		}
		if err := r.Component("/queue/filter", func(c *dk.Context) error {
			values := c.Interaction.MessageComponentData().Values
			if len(values) != 1 || !validStatus(values[0]) {
				return c.EphemeralText("Filtro inválido.")
			}
			return c.Update(b.queue(c, values[0], 0))
		}); err != nil {
			return err
		}
		return r.Component("/queue/:status/:page", func(c *dk.Context) error {
			status, err := c.RequireParam("status")
			if err != nil {
				return err
			}
			raw, err := c.RequireParam("page")
			if err != nil {
				return err
			}
			page, err := strconv.Atoi(raw)
			if err != nil || !validStatus(status) {
				return c.EphemeralText("Página inválida.")
			}
			return c.Update(b.queue(c, status, page))
		})
	}
}
