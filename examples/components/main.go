// Command components shows a Components V2 message, a component route with a
// typed custom ID, and a modal form round-trip.
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

	r := discordkit.NewRouter(discordkit.Recovery())
	must(r.Command("panel", panel))
	must(r.Component("/jobs/:jobID/feedback", openFeedback))
	must(r.Modal("feedback:new", saveFeedback))
	s.AddHandler(r.Handle)

	if err := s.Open(); err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	cmds, _ := discordkit.BuildCommands(discordkit.Command("panel", "Show the demo panel"))
	if _, err := discordkit.SyncCommands(s, s.State.User.ID, cmds, discordkit.SyncOptions{}); err != nil {
		log.Fatal(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}

func panel(c *discordkit.Context) error {
	id, err := discordkit.Route("/jobs/:jobID/feedback").Param("jobID", "42").Build()
	if err != nil {
		return err
	}
	return c.Reply(discordkit.MessageSpec{
		Components: discordkit.Components(
			discordkit.Container(
				discordkit.Text("## Job #42"),
				discordkit.Separator().Divider(true),
				discordkit.Section(discordkit.Text("Tell us how it went.")).
					Accessory(discordkit.Button("Give feedback", string(id))),
			),
		),
	})
}

func openFeedback(c *discordkit.Context) error {
	jobID, _ := c.Param("jobID")
	modal, err := discordkit.Form("feedback:new", "Feedback for job "+jobID,
		discordkit.Field("Summary", discordkit.TextInput("summary")),
		discordkit.Field("Details", discordkit.TextInput("details").Paragraph()),
	).Build()
	if err != nil {
		return err
	}
	return c.ShowModal(modal)
}

func saveFeedback(c *discordkit.Context) error {
	f := c.Form()
	summary, _ := f.String("summary")
	return c.EphemeralText("Thanks! We recorded: " + summary)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
