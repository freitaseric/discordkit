---
title: "02 · Organize the Go application"
description: "Packages, dependencies, configuration, Gateway and shutdown."
---

**Checkpoint:** keep `COOKBOOK_STAGE=1`. Understand the skeleton before enabling features.

| Package | Responsibility |
| --- | --- |
| `cmd/bot` | Compose dependencies and control process lifetime |
| `internal/config` | Read and validate environment variables |
| `internal/store` | Authorize, validate and persist tickets |
| `internal/bot` | Translate Discord interactions into operations and views |

Go enforces the parent-tree import boundary for `internal`; it is not a data security boundary. `cmd/bot` is a convention. Files in `internal/bot` share one package, not separate services. `Bot` receives its store explicitly. Introduce a consumer-owned interface when a second storage implementation makes it useful.

## Configuration

`Load` returns `(Config, error)` and rejects missing credentials or invalid stages before connecting. Never log the full config because it includes the token.

<!-- cookbook-source: internal/config/config.go -->
```go title="internal/config/config.go"
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Token, GuildID, DataFile string
	Stage                    int
}

func Load() (Config, error) {
	c := Config{Token: strings.TrimSpace(os.Getenv("DISCORD_TOKEN")), GuildID: strings.TrimSpace(os.Getenv("DISCORD_GUILD_ID")), DataFile: os.Getenv("BOT_DATA_FILE"), Stage: 6}
	if c.Token == "" || c.GuildID == "" {
		return c, fmt.Errorf("set DISCORD_TOKEN and DISCORD_GUILD_ID")
	}
	if c.DataFile == "" {
		c.DataFile = "data/community.json"
	}
	if v := os.Getenv("COOKBOOK_STAGE"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 6 {
			return c, fmt.Errorf("COOKBOOK_STAGE must be 1..6")
		}
		c.Stage = n
	}
	return c, nil
}
```
<!-- /cookbook-source -->

## Process lifecycle

Read `run` in order: config, store, bot, command definitions, router, session, handlers, connection, sync, shutdown signal. Local builder and registration errors fail before opening the Gateway.

`defer session.Close()` runs when `run` returns. `signal.NotifyContext` turns Ctrl+C/SIGTERM into cancellation. The example closes the session but does not implement full draining of in-flight handlers. Stop during an idle period before a consistent backup.

`Ready` may fire after reconnects. It only sets presence; publishing panels or syncing commands there would repeat side effects. `UpdateGameStatus` illustrates using discordgo directly where no interaction abstraction is needed.

<!-- cookbook-source: cmd/bot/main.go -->
```go title="cmd/bot/main.go"
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	dk "github.com/freitaseric/discordkit"
	"github.com/freitaseric/discordkit/examples/community-bot/internal/bot"
	"github.com/freitaseric/discordkit/examples/community-bot/internal/config"
	"github.com/freitaseric/discordkit/examples/community-bot/internal/store"
)

func main() {
	if err := run(); err != nil {
		slog.Error("bot stopped", "error", err)
		os.Exit(1)
	}
}
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	db, err := store.Open(cfg.DataFile)
	if err != nil {
		return err
	}
	app := &bot.Bot{Store: db, GuildID: cfg.GuildID, Stage: cfg.Stage}
	commands, err := app.Commands()
	if err != nil {
		return err
	}
	router, err := app.Router()
	if err != nil {
		return err
	}
	session, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return err
	}
	session.Identify.Intents = discordgo.IntentsGuilds
	session.AddHandler(router.Handle)
	// Ready can run again after reconnects. Do not create panels or sync here.
	session.AddHandler(func(s *discordgo.Session, _ *discordgo.Ready) {
		if err := s.UpdateGameStatus(0, "/help · Central de atendimento"); err != nil {
			slog.Warn("presence failed", "error", err)
		}
	})
	if err = session.Open(); err != nil {
		return err
	}
	defer session.Close()
	// Sync is non-destructive: unrelated guild commands are not deleted.
	plan, err := dk.SyncCommands(session, session.State.User.ID, commands, dk.SyncOptions{GuildID: cfg.GuildID})
	if err != nil {
		return err
	}
	slog.Info("ready", "stage", cfg.Stage, "created", len(plan.Create), "updated", len(plan.Update), "unchanged", len(plan.Unchanged))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	return nil
}
```
<!-- /cookbook-source -->

## Middleware and errors

Global middleware runs logging → configured guild and context → RequireGuild → handler. Router automatically wraps execution in Recovery. Recovery converts panics into errors; it does not reply.

The scope middleware adds a timeout to Context REST requests. It does not cancel filesystem calls or replace Discord's initial response deadline. Failures after defer go through `finish`, which edits the acknowledged response. The error hook does not blindly create followups after uncertain network errors or modal responses.

<!-- cookbook-source: internal/bot/bot.go -->
```go title="internal/bot/bot.go"
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
```
<!-- /cookbook-source -->

**Verify:** run `go test ./...`, then start with `COOKBOOK_STAGE=0` and confirm a clear startup failure. Restore `1`. Restarting must not publish channel messages.

**Exercise:** add `BOT_LOG_LEVEL` to config and initialize `slog` in main. Avoid reading environment variables inside handlers.

Next: [command definitions and handlers](/cookbook/commands/).

---

[← Previous: 01 · Prepare and connect the bot](/cookbook/setup/) · [Next →: 03 · Commands and options](/cookbook/commands/)
