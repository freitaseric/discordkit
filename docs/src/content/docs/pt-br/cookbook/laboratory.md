---
title: "09 · Explore o restante da API"
description: "Experimentos executáveis sem misturar demonstrações com regras de chamados."
---

A etapa 6 habilita `/lab` e dois menus de contexto. No Discord em português, a localização do nome pode aparecer como `/laboratorio`; o router continua usando o nome canônico `lab`.

Nem toda API melhora o bot quando colocada no fluxo principal. Cobrar um SKU, criar um poll ou pedir cinco tipos de campo em cada chamado tornaria a experiência pior. O laboratório oferece experimentos reais separados, e o mapa da API explica as variantes que não precisam virar uma função artificial do atendimento.

## Opções resolvidas e grupos

Execute `/lab options inspect`. O `SubCommandGroup` adiciona um nível ao caminho: `lab options inspect`. Escolha usuário, cargo, canal, mentionable, anexo e número. Os getters retornam os objetos em `resolved` enviados pelo Discord, não fazem buscas REST escondidas.

`UserOption` e `MemberOption` não são equivalentes: o usuário identifica uma conta; o membro traz dados daquele servidor e pode não estar presente no payload. `Mentionable` distingue usuário de cargo. `Attachment` oferece metadados e URL, mas o exemplo **não baixa arquivos enviados pelo usuário**. URLs assinadas de anexos podem expirar; salvá-las em JSON não é guardar o arquivo.

## Menus de contexto

Clique com o botão direito em usuário ou mensagem e abra Apps → Inspect user/Inspect message. Esses comandos não têm opções slash nem descrição. O alvo está em `ApplicationCommandData().TargetID`; os dados resolvidos estão nos mapas nativos do discordgo. Leia o alvo correto em vez de supor que ele é sempre o autor da interação.

## Layouts e escape hatches

`/lab layout` combina Section, Thumbnail, Gallery e um botão discordgo encapsulado por `Raw`. Uma Section precisa de acessório. A galeria usa o avatar do próprio usuário para evitar depender de um link externo arbitrário.

`/lab embed` responde com um embed legado em uma mensagem separada. Não misture esse payload com Container/Text/Gallery V2. `Raw` não pula a validação nem torna suportado um componente que o Discord ou discordgo não entende.

## Formulários além de texto

`/lab form` demonstra StringSelect, UserSelect, RoleSelect, ChannelSelect e FileUpload dentro de Fields. Modal tem no máximo cinco componentes de topo, por isso não adicionamos mais campos. `Form().Strings` lê os IDs/valores selecionados; `Files` resolve os metadados dos arquivos. `RequireStrings`/`RequireFiles` retornam erro se a entrada necessária faltar. O exemplo responde com contagens e não persiste nem baixa esses dados.

`MentionableSelect` é a alternativa quando um mesmo menu deve aceitar usuário ou cargo. `DefaultValues`, `Placeholder`, `MinValues`, `MaxValues`, `Disabled`, emojis e opções default refinam a interface sem alterar o fluxo de roteamento. Teste uma variante por vez para entender o payload resultante.

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

## Inspecione uma sincronização antes de aplicá-la

Com uma sessão aberta, o trecho abaixo substitui temporariamente a chamada de sync em `main`. A revisão dos nomes e tipos deve acontecer antes de qualquer exclusão:

```go
remote, err := session.ApplicationCommands(session.State.User.ID, cfg.GuildID)
if err != nil { return err }
plan := dk.DiffCommands(remote, commands, false)
slog.Info("sync preview", "create", len(plan.Create),
    "update", len(plan.Update), "delete", len(plan.Delete))
```

`DiffCommands` calcula um plano a partir de duas listas já disponíveis, sem rede. `SyncPlan.Empty` indica ausência de mudanças. `Delete: true` deve ser uma decisão consciente para uma aplicação que você controla inteiramente.

## Quando recorrer ao discordgo

Use `c.Session` para endpoints fora do ciclo de resposta, como editar uma mensagem pública antiga por canal/ID. `MessageSpec.MessageEdit(channelID, messageID)` converte o payload; `Session.ChannelMessageEditComplex` envia. Confira a permissão do bot, valide o destino e passe `discordgo.WithContext(c.Context())` quando aplicável.

Evite responder à mesma interação diretamente via `Session.InteractionRespond`: isso passa por fora do controle de estado do Context. Use `ModalOf` quando já tiver componentes nativos de um modal; prefira Form/Field nos novos exemplos.

**Confira:** execute os quatro comandos de laboratório, os dois menus de contexto e envie o formulário com/sem arquivo opcional. O arquivo de chamados deve permanecer inalterado.

**Exercício:** troque o select de cargos por `MentionableSelect`, leia os valores e explique como distinguir os tipos usando os dados resolvidos. Não transforme uma seleção visual em autorização de equipe.

---

[← Anterior: 08 · Teste e opere o bot](/pt-br/cookbook/operations/)
