package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/freitaseric/discordkit"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	token := strings.TrimSpace(os.Getenv("DISCORD_TOKEN"))
	guildID := strings.TrimSpace(os.Getenv("DISCORD_GUILD_ID"))
	if token == "" || guildID == "" {
		return fmt.Errorf("set DISCORD_TOKEN and DISCORD_GUILD_ID")
	}
	commands, err := discordkit.BuildCommands(
		discordkit.Command("ping", "Check the bot connection"),
		discordkit.Command("help", "Show private help"),
		discordkit.Command("panel", "Publish a support panel (Manage Messages required)"),
	)
	if err != nil {
		return err
	}

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return err
	}
	session.Identify.Intents = discordgo.IntentsGuilds
	router := discordkit.NewRouter(discordkit.Logging(nil))
	router.OnError(func(c *discordkit.Context, err error) {
		log.Printf("interaction failed: %v", err)
		// A reply may fail if the interaction was already acknowledged.
		if replyErr := c.EphemeralText("Could not complete this action. Please try again."); replyErr != nil {
			log.Printf("error reply failed: %v", replyErr)
		}
	})
	if err := router.Command("ping", func(c *discordkit.Context) error {
		return c.EphemeralText("Pong!")
	}); err != nil {
		return err
	}
	if err := router.Command("help", func(c *discordkit.Context) error {
		return c.Ephemeral(panel())
	}); err != nil {
		return err
	}
	if err := router.Command("panel", func(c *discordkit.Context) error {
		member := c.Member()
		if member == nil || member.Permissions&(discordgo.PermissionManageMessages|discordgo.PermissionAdministrator) == 0 {
			return c.EphemeralText("You need Manage Messages to publish this panel.")
		}
		return c.Reply(panel())
	}); err != nil {
		return err
	}
	if err := router.Component("/help/:topic", func(c *discordkit.Context) error {
		topic, err := c.RequireParam("topic")
		if err != nil {
			return err
		}
		switch topic {
		case "rules":
			return c.EphemeralText("Be respectful. Avoid spam. Read pinned messages before posting.")
		case "support":
			return c.EphemeralText("Ask your question in the support channel with steps to reproduce. Never share tokens.")
		default:
			return c.EphemeralText("Unknown help topic.")
		}
	}); err != nil {
		return err
	}

	session.AddHandler(router.Handle)
	if err := session.Open(); err != nil {
		return err
	}
	defer session.Close()
	plan, err := discordkit.SyncCommands(session, session.State.User.ID, commands,
		discordkit.SyncOptions{GuildID: guildID})
	if err != nil {
		return err
	}
	log.Printf("ready: %d created, %d updated, %d unchanged", len(plan.Create), len(plan.Update), len(plan.Unchanged))
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)
	<-stop
	return nil
}

func panel() discordkit.MessageSpec {
	return discordkit.MessageSpec{
		Components: discordkit.Components(
			discordkit.Text("## Help center\nChoose a topic below. Only you can see the answer."),
			discordkit.Row(
				discordkit.Button("Server rules", "/help/rules").Secondary(),
				discordkit.Button("Get support", "/help/support").Primary(),
				discordkit.LinkButton("DiscordKit docs", "https://discordkit.freitaseric.com"),
			),
		),
	}
}
