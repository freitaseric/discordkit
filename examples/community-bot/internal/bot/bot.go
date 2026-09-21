package bot

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
	dk "github.com/freitaseric/discordkit"
	"github.com/freitaseric/discordkit/examples/community-bot/internal/store"
)

type Bot struct {
	Store   *store.Store
	GuildID string
	Stage   int
}

func actor(c *dk.Context) store.Actor {
	a := store.Actor{GuildID: c.GuildID()}
	if u := c.User(); u != nil {
		a.UserID = u.ID
	}
	if m := c.Member(); m != nil {
		a.Staff = m.Permissions&(discordgo.PermissionManageMessages|discordgo.PermissionAdministrator) != 0
	}
	return a
}
func (b *Bot) scope(next dk.Handler) dk.Handler {
	return func(c *dk.Context) error {
		if c.GuildID() != b.GuildID {
			return c.EphemeralText("Este bot atende apenas o servidor configurado.")
		}
		ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
		defer cancel()
		c.SetContext(ctx)
		return next(c)
	}
}
func friendly(err error) string {
	switch {
	case errors.Is(err, store.ErrNotFound):
		return "Chamado não encontrado ou sem acesso."
	case errors.Is(err, store.ErrForbidden):
		return "Você não pode executar esta ação."
	case errors.Is(err, store.ErrClosed):
		return "Este chamado já foi encerrado."
	case errors.Is(err, store.ErrInvalid):
		return "Confira os campos e tente novamente."
	default:
		return "Não foi possível concluir. Consulte o chamado antes de repetir a ação."
	}
}

// finish handles failures after Defer/DeferUpdate without a second initial reply.
func finish(c *dk.Context, spec dk.MessageSpec, err error) error {
	if err != nil {
		slog.Error("ticket operation failed", "error", err)
		spec = message(dk.Text(friendly(err)))
	}
	_, sendErr := c.Edit(spec)
	return sendErr
}
func (b *Bot) Router() (*dk.Router, error) {
	r := dk.NewRouter(dk.Logging(nil), b.scope, dk.RequireGuild())
	r.OnError(func(c *dk.Context, err error) {
		slog.Error("interaction failed", "error", err)
		if c.Interaction.Type == discordgo.InteractionApplicationCommandAutocomplete {
			_ = c.Autocomplete()
			return
		}
		// ErrAlreadyResponded means an initial response was attempted, not necessarily delivered.
		// Do not blindly create a followup after a modal or an uncertain network failure.
		if e := c.EphemeralText(friendly(err)); e != nil && !errors.Is(e, dk.ErrAlreadyResponded) {
			slog.Error("error response failed", "error", e)
		}
	})
	registrations := []func() error{
		func() error {
			return r.Command("ping", func(c *dk.Context) error { return c.EphemeralText("Pong! DiscordKit está conectado.") })
		},
	}
	if b.Stage >= 2 {
		registrations = append(registrations, func() error { return r.Command("about", b.about) })
	}
	if b.Stage >= 3 {
		registrations = append(registrations,
			func() error {
				return r.Command("help", func(c *dk.Context) error { return c.Ephemeral(panel(b.Stage)) })
			},
			func() error {
				return r.Command("panel", func(c *dk.Context) error { return c.Reply(panel(b.Stage)) }, dk.RequirePermissions(discordgo.PermissionManageMessages))
			},
			func() error { return r.Component("/help/topic", b.topic) },
		)
	}
	if b.Stage >= 4 {
		registrations = append(registrations, b.registerTickets(r))
	}
	if b.Stage >= 5 {
		registrations = append(registrations, b.registerQueue(r))
	}
	if b.Stage >= 6 {
		registrations = append(registrations, b.registerOperations(r), func() error { return b.registerLab(r) })
	}
	for _, register := range registrations {
		if err := register(); err != nil {
			return nil, err
		}
	}
	return r, nil
}
