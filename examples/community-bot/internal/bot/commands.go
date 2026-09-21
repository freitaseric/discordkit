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
