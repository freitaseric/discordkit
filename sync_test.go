package discordkit

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestDiffCommandsCreateUpdateDeleteUnchanged(t *testing.T) {
	remote := []*discordgo.ApplicationCommand{
		{ID: "1", Name: "same", Description: "d", Type: discordgo.ChatApplicationCommand},
		{ID: "2", Name: "change", Description: "old", Type: discordgo.ChatApplicationCommand},
		{ID: "3", Name: "gone", Description: "d", Type: discordgo.ChatApplicationCommand},
	}
	desired := []*discordgo.ApplicationCommand{
		{Name: "same", Description: "d", Type: discordgo.ChatApplicationCommand},
		{Name: "change", Description: "new", Type: discordgo.ChatApplicationCommand},
		{Name: "fresh", Description: "d", Type: discordgo.ChatApplicationCommand},
	}

	plan := DiffCommands(remote, desired, true)
	if len(plan.Create) != 1 || plan.Create[0].Name != "fresh" {
		t.Fatalf("create wrong: %+v", plan.Create)
	}
	if len(plan.Update) != 1 || plan.Update[0].Name != "change" || plan.Update[0].ID != "2" {
		t.Fatalf("update wrong: %+v", plan.Update)
	}
	if len(plan.Unchanged) != 1 || plan.Unchanged[0].Name != "same" {
		t.Fatalf("unchanged wrong: %+v", plan.Unchanged)
	}
	if len(plan.Delete) != 1 || plan.Delete[0].Name != "gone" {
		t.Fatalf("delete wrong: %+v", plan.Delete)
	}
	if plan.Empty() {
		t.Fatal("plan should not be empty")
	}
}

func TestDiffCommandsNoDeleteWhenDisabled(t *testing.T) {
	remote := []*discordgo.ApplicationCommand{
		{ID: "3", Name: "gone", Description: "d", Type: discordgo.ChatApplicationCommand},
	}
	plan := DiffCommands(remote, nil, false)
	if len(plan.Delete) != 0 {
		t.Fatalf("expected no deletes, got %+v", plan.Delete)
	}
}

func TestDiffCommandsSameNameDifferentType(t *testing.T) {
	remote := []*discordgo.ApplicationCommand{
		{ID: "1", Name: "thing", Type: discordgo.UserApplicationCommand},
	}
	desired := []*discordgo.ApplicationCommand{
		{Name: "thing", Description: "d", Type: discordgo.ChatApplicationCommand},
	}
	plan := DiffCommands(remote, desired, true)
	if len(plan.Create) != 1 || len(plan.Delete) != 1 {
		t.Fatalf("expected create+delete for type mismatch: %+v", plan)
	}
}

func TestDiffCommandsAllUnchangedIsEmpty(t *testing.T) {
	cmds := []*discordgo.ApplicationCommand{
		{ID: "1", Name: "a", Description: "d", Type: discordgo.ChatApplicationCommand},
	}
	desired := []*discordgo.ApplicationCommand{
		{Name: "a", Description: "d", Type: discordgo.ChatApplicationCommand},
	}
	if !DiffCommands(cmds, desired, true).Empty() {
		t.Fatal("expected empty plan")
	}
}
