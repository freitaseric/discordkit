package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	dk "github.com/freitaseric/discordkit"
)

// The optional lab keeps API experiments out of the ticket domain.
func labCommands() []*dk.CommandBuilder {
	return []*dk.CommandBuilder{
		dk.Command("lab", "Explore outras APIs",
			dk.SubCommandGroup("options", "Opções tipadas", dk.SubCommand("inspect", "Inspecione valores resolvidos",
				dk.UserOption("user", "Usuário"), dk.RoleOption("role", "Cargo"), dk.ChannelOption("channel", "Canal").ChannelTypes(discordgo.ChannelTypeGuildText), dk.MentionableOption("mentionable", "Usuário ou cargo"), dk.AttachmentOption("file", "Arquivo"), dk.NumberOption("weight", "Peso").Min(0).Max(10))),
			dk.SubCommand("layout", "Veja seções e mídia"),
			dk.SubCommand("form", "Experimente seleções e upload"),
			dk.SubCommand("embed", "Veja um embed legado"),
		).NameLocalizations(map[discordgo.Locale]string{discordgo.PortugueseBR: "laboratorio"}),
		dk.UserCommand("Inspect user"), dk.MessageCommand("Inspect message"),
	}
}
func (b *Bot) registerLab(r *dk.Router) error {
	handlers := map[string]dk.Handler{
		"lab options inspect": func(c *dk.Context) error {
			lines := []string{"## Valores resolvidos"}
			if u, ok := c.UserOption("user"); ok {
				lines = append(lines, "Usuário: "+u.ID)
			}
			if m, ok := c.MemberOption("user"); ok {
				lines = append(lines, fmt.Sprintf("Cargos do membro: %d", len(m.Roles)))
			}
			if role, ok := c.Role("role"); ok {
				lines = append(lines, "Cargo: "+role.ID)
			}
			if ch, ok := c.Channel("channel"); ok {
				lines = append(lines, "Canal: "+ch.ID)
			}
			if m, ok := c.Mentionable("mentionable"); ok {
				if m.User != nil {
					lines = append(lines, "Mentionable usuário: "+m.User.ID)
				}
				if m.Role != nil {
					lines = append(lines, "Mentionable cargo: "+m.Role.ID)
				}
			}
			if a, ok := c.Attachment("file"); ok {
				lines = append(lines, fmt.Sprintf("Arquivo: %s (%d bytes)", safe(a.Filename), a.Size))
			}
			if n, ok := c.Float("weight"); ok {
				lines = append(lines, fmt.Sprintf("Peso: %.2f", n))
			}
			return c.Ephemeral(message(dk.Text(strings.Join(lines, "\n"))))
		},
		"lab layout": func(c *dk.Context) error {
			avatar := c.User().AvatarURL("128")
			return c.Ephemeral(message(dk.Container(
				dk.Section(dk.Text("## Seu perfil"), dk.Text("Uma seção combina texto com um acessório.")).Accessory(dk.Thumbnail(avatar).Description("Seu avatar")),
				dk.Separator().Large().Divider(true),
				dk.Gallery(dk.MediaItem(avatar).Description("Avatar do solicitante")),
				dk.Row(dk.Raw(discordgo.Button{Label: "Documentação", Style: discordgo.LinkButton, URL: "https://discordkit.freitaseric.com"})),
			).AccentColor(0x38D9B0)))
		},
		"lab embed": func(c *dk.Context) error {
			return c.Ephemeral(dk.MessageSpec{Embeds: []*discordgo.MessageEmbed{{Title: "Embed legado", Description: "Não combine embeds com layouts V2.", Color: 0x38D9B0}}})
		},
		"lab form": func(c *dk.Context) error {
			form, err := dk.Form("/lab/form", "Formulário de laboratório",
				dk.Field("Categoria", dk.StringSelect("category").Options(dk.SelectOption("Dúvida", "question"), dk.SelectOption("Problema", "issue")).Required()),
				dk.Field("Pessoas", dk.UserSelect("people").MinValues(0).MaxValues(2)),
				dk.Field("Cargos", dk.RoleSelect("roles").MinValues(0).MaxValues(2)),
				dk.Field("Canal", dk.ChannelSelect("channel").ChannelTypes(discordgo.ChannelTypeGuildText).MinValues(0).MaxValues(1)),
				dk.Field("Arquivo opcional", dk.FileUpload("evidence").Required(false).MaxValues(1)),
			).Build()
			if err != nil {
				return err
			}
			return c.ShowModal(form)
		},
		"Inspect user": func(c *dk.Context) error {
			data := c.Interaction.ApplicationCommandData()
			u := data.Resolved.Users[data.TargetID]
			if u == nil {
				return c.EphemeralText("Usuário não resolvido.")
			}
			return c.EphemeralText("ID do usuário: " + u.ID)
		},
		"Inspect message": func(c *dk.Context) error {
			data := c.Interaction.ApplicationCommandData()
			m := data.Resolved.Messages[data.TargetID]
			if m == nil {
				return c.EphemeralText("Mensagem não resolvida.")
			}
			return c.EphemeralText("ID da mensagem: " + m.ID)
		},
	}
	for path, h := range handlers {
		if err := r.Command(path, h); err != nil {
			return err
		}
	}
	return r.Modal("/lab/form", func(c *dk.Context) error {
		category, err := c.Form().RequireStrings("category")
		if err != nil {
			return err
		}
		people, _ := c.Form().Strings("people")
		roles, _ := c.Form().Strings("roles")
		channels, _ := c.Form().Strings("channel")
		files, _ := c.Form().Files("evidence")
		return c.EphemeralText(fmt.Sprintf("Categoria: %s · pessoas: %d · cargos: %d · canais: %d · arquivos: %d. Nada foi salvo ou baixado.", strings.Join(category, ","), len(people), len(roles), len(channels), len(files)))
	})
}
