---
title: Your first bot
description: Create a complete Discord bot from scratch with Go, discordgo, and DiscordKit, from the Developer Portal to a working /ping command.
---

In this cookbook, you will create and run a real bot. At the end, it will be installed in a test server and respond to the `/ping` command.

The bot uses application commands, so it does not need to read messages or enable privileged intents.

## What you need

- Go 1.26 or newer;
- a Discord account;
- a test server where you can install applications;
- a terminal.

Check your Go version:

```bash
go version
```

## 1. Create a Discord application

1. Open the [Discord Developer Portal](https://discord.com/developers/applications) and select **New Application**.
2. Name your application and create it.
3. On the **Bot** page, generate or copy the bot token. Store it in a password manager. This token works like a password: do not share it or put it in your code.
4. On the **Installation** page, configure **Guild Install** with the `applications.commands` and `bot` scopes.
5. For this tutorial, grant only the **Send Messages** permission to the bot.
6. Use the **Install Link** to add the application to your test server.

Do not enable **Message Content Intent** or other privileged intents. This example receives slash command interactions through the Gateway and does not read regular messages.

## 2. Create the Go project

Create a folder and initialize the module:

```bash
mkdir my-bot
cd my-bot
go mod init example.com/my-bot
go get github.com/freitaseric/discordkit
```

DiscordKit installs the required discordgo dependency. It requires Go 1.26 or newer.

Create a `.gitignore` file:

```gitignore
.env
```

Create `.env` to keep your token out of your code:

```dotenv
DISCORD_TOKEN=paste-your-token-here
DISCORD_GUILD_ID=your-test-server-id
```

In Discord, enable **Developer Mode** in **User Settings → Advanced**, right-click your test server icon, and select **Copy Server ID**. Replace the example values in the file.

> Never publish the `.env` file. If your token leaks, reset it immediately on the **Bot** page of the Developer Portal.

## 3. Write the bot

Create `main.go`:

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
		log.Fatal("DISCORD_TOKEN is not set")
	}

	guildID := os.Getenv("DISCORD_GUILD_ID")
	if guildID == "" {
		log.Fatal("DISCORD_GUILD_ID is not set")
	}

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("create Discord session: %v", err)
	}

	router := discordkit.NewRouter(
		discordkit.Recovery(),
		discordkit.Logging(nil),
	)

	if err := router.Command("ping", func(c *discordkit.Context) error {
		return c.ReplyText("Pong! 🏓")
	}); err != nil {
		log.Fatalf("register /ping route: %v", err)
	}

	session.AddHandler(router.Handle)

	if err := session.Open(); err != nil {
		log.Fatalf("connect to Discord: %v", err)
	}
	defer session.Close()

	commands, err := discordkit.BuildCommands(
		discordkit.Command("ping", "Check whether the bot is responding"),
	)
	if err != nil {
		log.Fatalf("validate commands: %v", err)
	}

	plan, err := discordkit.SyncCommands(
		session,
		session.State.User.ID,
		commands,
		discordkit.SyncOptions{GuildID: guildID},
	)
	if err != nil {
		log.Fatalf("synchronize commands: %v", err)
	}
	log.Printf("commands synchronized: %d created, %d updated, %d unchanged",
		len(plan.Create), len(plan.Update), len(plan.Unchanged))

	log.Printf("bot connected as %s; press Ctrl+C to stop", session.State.User.Username)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}
```

The command is registered only in the test server by using `GuildID`. This makes it available quickly during development. `SyncCommands` creates or updates the declared command and leaves obsolete remote commands in place.

## 4. Run it

On macOS or Linux, load the variables from the file and start the bot:

```bash
set -a
source .env
set +a
go run .
```

In PowerShell:

```powershell
Get-Content .env | ForEach-Object {
    if ($_ -match '^([^#][^=]*)=(.*)$') {
        [Environment]::SetEnvironmentVariable($matches[1], $matches[2])
    }
}
go run .
```

Keep the process running. In your test server, type `/ping` and run the command. The bot should respond with `Pong! 🏓`.

## What just happened

- `discordgo.Session` opened a connection to the Discord Gateway;
- `router.Handle` forwarded incoming interactions to DiscordKit;
- `router.Command` connected the `ping` route to its handler;
- `BuildCommands` validated the slash command declaration;
- `SyncCommands` registered the declaration in your test server;
- `ReplyText` sent the initial interaction response.

Registering a route and publishing a command are separate steps: the route handles the interaction; synchronization makes the command appear in Discord.

## Troubleshooting

**The `/ping` command does not appear:** make sure the bot is installed in the correct server with the `applications.commands` scope, `DISCORD_GUILD_ID` is that server's ID, and the program output confirms synchronization.

**The bot appears offline:** make sure the process is still running and the token is valid. Generate a new token in the Developer Portal if the old one was exposed.

**`401: Unauthorized`:** the token is invalid or was copied with spaces. Generate a new token and update `DISCORD_TOKEN`.

**`403: Missing Access` while synchronizing:** confirm the server ID and install the application there with the `applications.commands` scope.

## Next steps

- Learn how the Router works in [Router](/concepts/router/).
- Understand responses and the interaction lifecycle in [The interaction model](/concepts/interaction-model/).
- Continue with [Your first command](./first-command/) to learn more about command declaration, synchronization, and routing.

For production, keep the token in a secrets manager, run the process under a supervisor, and synchronize global commands deliberately.
