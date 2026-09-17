// Command basic is a minimal DiscordKit bot: one slash command, synced on start.
package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/freitaseric/discordkit"
)

func main() {
	s, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	r := discordkit.NewRouter(discordkit.Recovery(), discordkit.Logging(nil))
	if err := r.Command("greet", greet); err != nil {
		log.Fatal(err)
	}
	s.AddHandler(r.Handle)

	if err := s.Open(); err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	cmds, err := discordkit.BuildCommands(
		discordkit.Command("greet", "Greet a member",
			discordkit.UserOption("who", "Who to greet").Required(),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := discordkit.SyncCommands(s, s.State.User.ID, cmds, discordkit.SyncOptions{}); err != nil {
		log.Fatal(err)
	}

	log.Println("running; press Ctrl+C to stop")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}

func greet(c *discordkit.Context) error {
	who, err := c.RequireUserOption("who")
	if err != nil {
		return c.EphemeralText("Please choose a member to greet.")
	}
	return c.ReplyText("Hello, " + who.Mention() + "!")
}
