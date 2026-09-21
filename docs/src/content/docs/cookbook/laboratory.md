---
title: "09 · Explore the remaining API"
description: "Runnable experiments separated from ticket business rules."
---

Stage 6 enables `/lab` and two context menus. Portuguese clients may display the localized `/laboratorio`; routing still uses canonical `lab`.

Not every API belongs in the core ticket workflow. Premium purchases, polls and every possible field type would distract from support. Use the lab for real experiments and the API map for variants and boundaries.

## Resolved options and groups

`/lab options inspect` uses a subcommand group and typed user, role, channel, mentionable, attachment and number options. Getters read resolved payload objects without hidden REST calls. Users and guild members are different entities; member data can be absent. Mentionables distinguish users and roles.

Attachment metadata does not mean the file was downloaded. Signed URLs may expire; storing a URL is not persisting its contents. This example never downloads user-provided files.

## Context menus and layouts

Right-click a user or message, then choose Apps → Inspect user/Inspect message. The command target is `TargetID`, not necessarily the invoking user. Resolved targets live in discordgo's native maps.

`/lab layout` combines a Section with a Thumbnail accessory, Gallery and Raw-wrapped discordgo button. `/lab embed` sends a separate legacy embed. Do not mix V2 layouts and embeds. Raw remains subject to validation and upstream support.

## Non-text form fields

`/lab form` demonstrates string, user, role and channel selects plus an optional file upload. Five top-level fields fill the modal limit. `Form().Strings` reads values/IDs; `Files` resolves attachment metadata. Require variants reject missing input. This lab returns counts without storing or downloading data.

MentionableSelect accepts users or roles in one control. DefaultValues, bounds, placeholders, disabled states and emojis customize controls without changing routing.

<!-- cookbook-source: internal/bot/lab.go -->
```go title="internal/bot/lab.go"
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
```
<!-- /cookbook-source -->

## Review sync without applying changes

With an open session, temporarily replace the sync call in main:

```go
remote, err := session.ApplicationCommands(session.State.User.ID, cfg.GuildID)
if err != nil { return err }
plan := dk.DiffCommands(remote, commands, false)
slog.Info("sync preview", "create", len(plan.Create),
    "update", len(plan.Update), "delete", len(plan.Delete))
```

`DiffCommands` compares two existing lists offline; `SyncPlan.Empty` reports no changes. Enable `Delete: true` only after reviewing the affected commands in an application you fully control.

Use `c.Session` for operations outside interaction responses, such as editing an old public message by channel/ID. Convert with `MessageSpec.MessageEdit`, then send using `ChannelMessageEditComplex` with appropriate permissions and context. Avoid raw InteractionRespond calls for a Context-managed interaction: they bypass its response tracking. `ModalOf` accepts existing native modal components.

**Verify:** run all lab commands and both context menus; submit the form with and without a file. The ticket database must remain unchanged.

**Exercise:** replace the role selector with MentionableSelect, inspect resolved types, and explain why selecting a role does not grant staff authorization.

---

[← Previous: 08 · Test and operate the bot](/cookbook/operations/)
