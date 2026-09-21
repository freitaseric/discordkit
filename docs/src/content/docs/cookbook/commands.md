---
title: "03 · Commands and options"
description: "Separate Discord command definitions from response handlers."
---

**Enable:** set `COOKBOOK_STAGE=2` and run `go run ./cmd/bot`. `/about` joins `/ping`.

A command builder describes the command displayed by Discord. `router.Command` registers its executable handler. You need both, with matching canonical names. Nested handler paths use spaces: `ticket show`, not `/ticket/show`.

Read only the `ping` and `about` definitions for now; later stages enable the remaining commands. The optional boolean `private` demonstrates why `c.Bool` returns both a value and a presence flag. Missing is different from false, so the handler explicitly defaults to a private response.

<!-- cookbook-source: internal/bot/commands.go -->
```go title="internal/bot/commands.go"
package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	dk "github.com/freitaseric/discordkit"
)

func (b *Bot) Commands() ([]*discordgo.ApplicationCommand, error) {
	commands := []*dk.CommandBuilder{dk.Command("ping", "Verifique a conexão")}
	if b.Stage >= 2 {
		commands = append(commands, dk.Command("about", "Conheça o bot", dk.BooleanOption("private", "Responder apenas para você")))
	}
	if b.Stage >= 3 {
		commands = append(commands, dk.Command("help", "Abra a central de ajuda"), dk.Command("panel", "Publique a central de ajuda").DefaultMemberPermissions(discordgo.PermissionManageMessages))
	}
	if b.Stage >= 4 {
		subs := []dk.Option{
			dk.SubCommand("new", "Abra um chamado"),
			dk.SubCommand("show", "Consulte um chamado", dk.StringOption("id", "ID do chamado").Required().Autocomplete()),
		}
		if b.Stage >= 5 {
			subs = append(subs, dk.SubCommand("list", "Consulte a fila", dk.StringOption("status", "Filtro").Choices(dk.Choice("Abertos", "open"), dk.Choice("Encerrados", "closed"), dk.Choice("Todos", "all"))))
		}
		if b.Stage >= 6 {
			subs = append(subs, dk.SubCommand("rate", "Avalie um chamado encerrado", dk.StringOption("id", "ID do chamado").Required().Autocomplete(), dk.IntegerOption("score", "Nota de 1 a 5").Required().Min(1).Max(5)), dk.SubCommand("export", "Exporte os chamados que você pode acessar"))
		}
		commands = append(commands, dk.Command("ticket", "Gerencie chamados", subs...))
	}
	if b.Stage >= 6 {
		commands = append(commands, labCommands()...)
	}
	return dk.BuildCommands(commands...)
}
func (b *Bot) about(c *dk.Context) error {
	private, ok := c.Bool("private")
	if !ok {
		private = true
	}
	spec := message(dk.Text(fmt.Sprintf("## Central DiscordKit\nEtapa %d/6 · Go + discordgo + DiscordKit\nIdioma da interação: %s", b.Stage, c.Locale())))
	if private {
		return c.Ephemeral(spec)
	}
	return c.Reply(spec)
}
func (b *Bot) topic(c *dk.Context) error {
	data := c.Interaction.MessageComponentData()
	if len(data.Values) != 1 {
		return c.EphemeralText("Escolha um assunto.")
	}
	switch data.Values[0] {
	case "rules":
		return c.EphemeralText("Respeite os membros. Não envie spam nem credenciais.")
	case "support":
		return c.EphemeralText("Descreva o problema e como reproduzi-lo. Consulte /ticket list para acompanhar.")
	default:
		return c.EphemeralText("Assunto desconhecido.")
	}
}
```
<!-- /cookbook-source -->

`BuildCommands` lets bootstrap report validation errors. `MustBuild` is suitable for known valid constants but panics on invalid input. The `about` handler shows `c.Locale()` and chooses between `Ephemeral` and `Reply`. Visibility is decided at the initial response; a later edit cannot make a public reply private.

## Nested commands and input constraints

Later, `/ticket` groups new, show, list, rate and export. Each subcommand has both a builder definition and a full router path. Fixed choices fit status filters; autocomplete fits changing ticket IDs. Choices and autocomplete are mutually exclusive on one option.

`IntegerOption` constrains the score to 1–5 in Discord, but the store validates it again. UI constraints are not a substitute for domain validation.

## Synchronization

The application syncs once during startup, scoped to the test guild. Registering a local handler alone does not create a Discord command. Sync preserves unrelated commands by default.

Going backwards between stages can leave visible commands without active handlers. Advance again or remove only your experimental commands explicitly. Do not enable `Delete: true` on a shared application without reviewing a dry-run plan.

**Verify:** `/about` is private; `/about private:false` is public. Change the builder description, restart and check Discord's command picker.

**Exercise:** add an optional boolean `verbose` and read it with `c.Bool` to show additional information.

Next: [panels and components](/cookbook/components/).

---

[← Previous: 02 · Organize the Go application](/cookbook/architecture/) · [Next →: 04 · Panel, components and permissions](/cookbook/components/)
