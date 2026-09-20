---
title: Seu primeiro bot
description: Crie um bot Discord completo do zero com Go, discordgo e DiscordKit, do Developer Portal ao comando /ping funcionando.
---

Neste cookbook, você vai criar e executar um bot de verdade. Ao final, ele estará instalado em um servidor de teste e responderá ao comando `/ping`.

O bot usa comandos de aplicação, então não precisa ler mensagens nem habilitar intents privilegiadas.

## O que você precisa

- Go 1.26 ou mais recente;
- uma conta Discord;
- um servidor de teste onde você possa instalar aplicativos;
- um terminal.

Confira a versão do Go:

```bash
go version
```

## 1. Crie o aplicativo Discord

1. Abra o [Discord Developer Portal](https://discord.com/developers/applications) e selecione **New Application**.
2. Dê um nome ao aplicativo e crie-o.
3. Na página **Bot**, gere ou copie o token do bot. Guarde-o em um gerenciador de senhas. Esse token funciona como uma senha: não o compartilhe e nunca o coloque no código.
4. Na página **Installation**, configure **Guild Install** com os escopos `applications.commands` e `bot`.
5. Para este tutorial, conceda apenas a permissão **Send Messages** ao bot.
6. Use o **Install Link** para adicionar o aplicativo ao seu servidor de teste.

Não habilite **Message Content Intent** nem outros intents privilegiados. Este exemplo recebe interações de slash commands pelo Gateway e não lê mensagens comuns.

## 2. Crie o projeto Go

Crie uma pasta e inicialize o módulo:

```bash
mkdir meu-bot
cd meu-bot
go mod init example.com/meu-bot
go get github.com/freitaseric/discordkit
```

DiscordKit instala o discordgo necessário como dependência. Requer Go 1.26 ou mais recente.

Crie o arquivo `.gitignore`:

```gitignore
.env
```

Crie `.env` para manter seu token fora do código:

```dotenv
DISCORD_TOKEN=cole-seu-token-aqui
DISCORD_GUILD_ID=id-do-seu-servidor-de-teste
```

No Discord, ative **Developer Mode** em **User Settings → Advanced**, clique com o botão direito no ícone do servidor e escolha **Copy Server ID**. Substitua os valores de exemplo no arquivo.

> Nunca publique o arquivo `.env`. Se o token vazar, redefina-o imediatamente na página **Bot** do Developer Portal.

## 3. Escreva o bot

Crie `main.go`:

```go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/freitaseric/discordkit"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN não foi definido")
	}

	guildID := os.Getenv("DISCORD_GUILD_ID")
	if guildID == "" {
		log.Fatal("DISCORD_GUILD_ID não foi definido")
	}

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("criar sessão Discord: %v", err)
	}

	router := discordkit.NewRouter(
		discordkit.Recovery(),
		discordkit.Logging(nil),
	)

	if err := router.Command("ping", func(c *discordkit.Context) error {
		return c.ReplyText("Pong! 🏓")
	}); err != nil {
		log.Fatalf("registrar rota /ping: %v", err)
	}

	session.AddHandler(router.Handle)

	if err := session.Open(); err != nil {
		log.Fatalf("conectar ao Discord: %v", err)
	}
	defer session.Close()

	commands, err := discordkit.BuildCommands(
		discordkit.Command("ping", "Verifica se o bot está respondendo"),
	)
	if err != nil {
		log.Fatalf("validar comandos: %v", err)
	}

	plan, err := discordkit.SyncCommands(
		session,
		session.State.User.ID,
		commands,
		discordkit.SyncOptions{GuildID: guildID},
	)
	if err != nil {
		log.Fatalf("sincronizar comandos: %v", err)
	}
	log.Printf("comandos sincronizados: %d criados, %d atualizados, %d sem alterações",
		len(plan.Create), len(plan.Update), len(plan.Unchanged))

	log.Printf("bot conectado como %s; pressione Ctrl+C para encerrar", session.State.User.Username)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}
```

O comando é registrado somente no servidor de teste usando `GuildID`. Isso deixa o comando disponível rapidamente durante o desenvolvimento. `SyncCommands` cria ou atualiza o comando declarado e mantém comandos remotos obsoletos sem removê-los.

## 4. Execute

No macOS ou Linux, carregue as variáveis do arquivo e inicie o bot:

```bash
set -a
source .env
set +a
go run .
```

No PowerShell:

```powershell
Get-Content .env | ForEach-Object {
    if ($_ -match '^([^#][^=]*)=(.*)$') {
        [Environment]::SetEnvironmentVariable($matches[1], $matches[2])
    }
}
go run .
```

Mantenha o processo em execução. No servidor de teste, digite `/ping` e execute o comando. O bot deve responder `Pong! 🏓`.

## O que acabou de acontecer

- `discordgo.Session` abriu a conexão com o Gateway do Discord;
- `router.Handle` encaminhou as interações recebidas ao DiscordKit;
- `router.Command` associou a rota `ping` ao handler;
- `BuildCommands` validou a declaração do slash command;
- `SyncCommands` registrou a declaração no servidor de teste;
- `ReplyText` enviou a resposta inicial da interação.

Registrar a rota e publicar o comando são etapas separadas: a rota trata a interação; a sincronização faz o comando aparecer no Discord.

## Problemas comuns

**O comando `/ping` não aparece:** confirme que o bot foi instalado no servidor correto com o escopo `applications.commands`, que `DISCORD_GUILD_ID` é o ID desse servidor e que a saída do programa informa a sincronização.

**O bot aparece offline:** confirme que o processo continua rodando e que o token está correto. Gere um novo token no Developer Portal se o anterior foi exposto.

**`401: Unauthorized`:** o token é inválido ou foi copiado com espaços. Gere outro token e atualize `DISCORD_TOKEN`.

**`403: Missing Access` ao sincronizar:** confirme o ID do servidor e instale o aplicativo nele com o escopo `applications.commands`.

## Próximos passos

- Veja como o Router funciona em [Router](/pt-br/concepts/router/).
- Entenda as respostas e o ciclo de vida das interações em [Modelo de interações](/pt-br/concepts/interaction-model/).
- Continue em [Seu primeiro comando](./first-command/) para entender a declaração, sincronização e roteamento de comandos em mais detalhes.

Para produção, mantenha o token em um gerenciador de segredos, execute o processo sob um supervisor e sincronize comandos globais deliberadamente.
