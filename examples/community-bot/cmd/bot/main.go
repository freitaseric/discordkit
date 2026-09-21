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
