---
title: "01 · Prepare and connect the bot"
description: "From the Developer Portal to the first runnable checkpoint."
---

**Goal:** execute `/ping` in a test server. No panel or ticket writes yet.

## Install the Discord application

Create an app in the [Developer Portal](https://discord.com/developers/applications) and obtain its bot token. Configure a guild installation with `bot` and `applications.commands`. Grant the bot **View Channels**, **Send Messages**, **Embed Links**, and **Attach Files** in your test channel, not Administrator. Install the app and copy your server ID using Discord developer mode.

Leave **Interactions Endpoint URL** empty. This application receives events through the discordgo Gateway connection. It only requests `IntentsGuilds`; Message Content, Server Members and Presence privileged intents are unnecessary. The user's Manage Messages permission to publish panels is separate from the bot's channel permissions.

## Clone and configure

Install Go **1.26 or newer**, Git and an editor:

```bash
git clone https://github.com/freitaseric/discordkit.git
cd discordkit/examples/community-bot
go version
go test ./...
```

Do not run `go mod init` here. Go finds the repository's parent `go.mod`; the example packages already belong to that module.

Bash/Zsh:

```bash
read -rs -p 'Bot token: ' DISCORD_TOKEN; echo
export DISCORD_TOKEN
export DISCORD_GUILD_ID='YOUR_SERVER_ID'
export BOT_DATA_FILE='data/community.json'
export COOKBOOK_STAGE=1
go run ./cmd/bot
```

PowerShell 7:

```powershell
$env:DISCORD_TOKEN = Read-Host 'Bot token' -MaskInput
$env:DISCORD_GUILD_ID = 'YOUR_SERVER_ID'
$env:BOT_DATA_FILE = 'data/community.json'
$env:COOKBOOK_STAGE = '1'
go run ./cmd/bot
```

The application **does not automatically load `.env`**. `.env.example` is a variable reference. A relative data path is relative to the process working directory; use an absolute path in production.

## Verify and change something

Wait for `ready` with `stage=1`, then run `/ping` in the configured server. Only you should see the reply. Stop with Ctrl+C. The JSON file appears on the first ticket write, not on startup.

Missing command? Check server ID, installation scopes and sync errors. No response? Check whether the process is running. Do not start two processes sharing the file.

**Exercise:** change the ping response in `internal/bot/bot.go`, restart and check the result. The response text is independent from the registered command definition.

Next: [application architecture](/cookbook/architecture/).

---

[← Previous: The project: a support desk](/cookbook/support-bot/) · [Next →: 02 · Organize the Go application](/cookbook/architecture/)
