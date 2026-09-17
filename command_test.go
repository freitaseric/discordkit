package discordkit

import (
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestCommandBuildsOptions(t *testing.T) {
	cmd, err := Command("greet", "greet someone",
		StringOption("name", "the name").Required(),
		BooleanOption("loud", "shout it"),
	).Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if cmd.Name != "greet" || len(cmd.Options) != 2 {
		t.Fatalf("unexpected command: %+v", cmd)
	}
	if !cmd.Options[0].Required || cmd.Options[1].Required {
		t.Fatalf("required flags wrong: %+v", cmd.Options)
	}
}

func TestRequiredMustPrecedeOptional(t *testing.T) {
	_, err := Command("c", "d",
		StringOption("a", "a"),
		StringOption("b", "b").Required(),
	).Build()
	if !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("expected ErrInvalidCommand, got %v", err)
	}
}

func TestDuplicateOptionNameRejected(t *testing.T) {
	_, err := Command("c", "d",
		StringOption("dup", "a"),
		StringOption("dup", "b"),
	).Build()
	if !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("expected ErrInvalidCommand, got %v", err)
	}
}

func TestIntegerMaxZeroRejected(t *testing.T) {
	_, err := Command("c", "d", IntegerOption("n", "num").Max(0)).Build()
	if !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("expected ErrInvalidCommand for Max(0), got %v", err)
	}
	if _, err := Command("c", "d", IntegerOption("n", "num").Max(10)).Build(); err != nil {
		t.Fatalf("Max(10) should be valid: %v", err)
	}
}

func TestChoicesAndAutocompleteMutuallyExclusive(t *testing.T) {
	_, err := Command("c", "d",
		StringOption("s", "s").Choices(Choice("A", "a")).Autocomplete(),
	).Build()
	if !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("expected ErrInvalidCommand, got %v", err)
	}
}

func TestContextMenuRejectsOptions(t *testing.T) {
	b := UserCommand("Profile")
	b.options = append(b.options, StringOption("x", "y"))
	if _, err := b.Build(); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("expected ErrInvalidCommand, got %v", err)
	}
	if _, err := UserCommand("Profile").Build(); err != nil {
		t.Fatalf("valid user command: %v", err)
	}
}

func TestSubCommandsBuild(t *testing.T) {
	cmd, err := Command("config", "configure",
		SubCommandGroup("user", "user config",
			SubCommand("set", "set value", StringOption("key", "key").Required()),
		),
		SubCommand("reset", "reset all"),
	).Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if cmd.Options[0].Type != discordgo.ApplicationCommandOptionSubCommandGroup {
		t.Fatalf("expected group, got %v", cmd.Options[0].Type)
	}
	if cmd.Options[0].Options[0].Type != discordgo.ApplicationCommandOptionSubCommand {
		t.Fatalf("expected subcommand under group")
	}
}

func TestOptionLimit(t *testing.T) {
	var opts []Option
	for i := 0; i < 26; i++ {
		opts = append(opts, StringOption(string(rune('a'+i%26))+string(rune('0'+i)), "d"))
	}
	if _, err := Command("c", "d", opts...).Build(); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("expected ErrInvalidCommand for >25 options, got %v", err)
	}
}
