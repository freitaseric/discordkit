---
title: "Crie um bot de atendimento"
description: "Um bot completo com painel de ajuda, respostas privadas e verificação de permissões."
---

Esta receita cria uma central de ajuda para seu servidor: `/help` abre um painel privado, `/panel` publica o painel para todos e os botões respondem perguntas frequentes de forma privada. Apenas membros com Gerenciar Mensagens ou Administrador podem publicar o painel público.

## 1. Prepare a aplicação

Conclua os passos 1 e 2 de [Seu primeiro bot](/pt-br/getting-started/first-bot/): crie a aplicação, instale com `bot` e `applications.commands` e crie o módulo Go. Dê ao bot acesso ao canal de destino e permissão para enviar mensagens. Configure `DISCORD_TOKEN` e `DISCORD_GUILD_ID`.

No Developer Portal, deixe Interactions Endpoint URL vazio: este bot recebe interações pelo Gateway. Não precisa de intents privilegiadas.

## 2. Crie main.go

Substitua o `main.go` inteiro pelo programa abaixo. Ele contém todos os handlers; não precisa de outros arquivos Go nem de banco de dados. Há uma cópia em [examples/support-bot/main.go](https://github.com/freitaseric/discordkit/blob/main/examples/support-bot/main.go).

```go title="main.go"
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
```

## 3. Execute e confira

Carregue as variáveis de ambiente como mostrado no tutorial do primeiro bot e execute:

```bash
go mod tidy
go run .
```

1. Aguarde `ready` no terminal.
2. Execute `/ping`: somente você deve ver `Pong!`.
3. Execute `/help`: o painel privado exibe três botões.
4. Clique em Server rules e Get support: cada um retorna uma resposta privada.
5. Com Gerenciar Mensagens, execute `/panel` em um canal de ajuda. Outros membros poderão clicar nele.
6. Teste `/panel` com um membro comum: ele deve receber uma mensagem de permissão.
7. Reinicie o bot e clique no painel existente. Os custom IDs estáveis continuam funcionando após a reinicialização.

O painel público permanece no Discord até ser excluído. Ele não é recriado na inicialização. Não execute `/panel` repetidamente, a menos que queira vários painéis.

## 4. Adapte para seu servidor

Troque as respostas no `switch`, o texto do painel e a URL do link. Para adicionar um assunto, acrescente um botão e um case correspondente. A rota `/help/:topic` atende todos os assuntos. Cada row aceita no máximo cinco botões.

Este bot oferece orientações estáticas; não cria tickets nem persiste conversas. Se adicionar armazenamento de tickets, mantenha verificações de autorização no handler e faça defer antes de acessar um banco de dados demorado.

Continue em [Hospedando o bot](/pt-br/guides/deployment/) para mantê-lo online. Pressione Ctrl+C para encerrá-lo localmente.
