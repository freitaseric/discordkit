---

title: Your first bot
description: Connect a Discord bot using discordgo and prepare it to handle interactions with DiscordKit.
---------------------------------------------------------------------------------------------------------

In this guide, you will create the smallest useful DiscordKit application:
connect a bot to Discord and attach a DiscordKit router.

Commands will be added in the next guide.

## Before you begin

You need:

* a Discord application;
* a bot created for that application;
* the bot token;
* DiscordKit installed in your Go project.

Keep your bot token private. Do not commit it to Git.

For this guide, store it in an environment variable named `DISCORD_TOKEN`.

## Create the session

DiscordKit is built on top of discordgo, so your application still creates a
normal `discordgo.Session`.

Create `main.go`:

```go
package main

import (
    "log"
    "os"

    "github.com/bwmarrin/discordgo"
    "github.com/freitaseric/discordkit"
)

func main() {
    token := os.Getenv("DISCORD_TOKEN")
    if token == "" {
        log.Fatal("DISCORD_TOKEN is not set")
    }

    session, err := discordgo.New("Bot " + token)
    if err != nil {
        log.Fatal(err)
    }

    router := discordkit.NewRouter()

    session.AddHandler(router.Handle)
}
```

At this point, three important objects exist.

### `discordgo.Session`

```go
session
```

The session owns the connection to Discord and exposes discordgo's Gateway and
REST functionality.

DiscordKit does not replace it.

### `discordkit.Router`

```go
router
```

The router decides which application handler should receive an incoming
interaction.

Later, the same router will handle commands, buttons, selects, modal
submissions, and autocomplete interactions.

### `router.Handle`

```go
session.AddHandler(router.Handle)
```

`Router.Handle` is a normal discordgo interaction handler.

When discordgo receives an interaction, it passes it to DiscordKit. DiscordKit
creates a `Context`, resolves the matching route, applies middleware, and invokes
your handler.

Conceptually:

```text
Discord
   ↓
discordgo.Session
   ↓
Router.Handle
   ↓
DiscordKit Router
   ↓
Middleware
   ↓
Handler
```

## Open the connection

Add:

```go
if err := session.Open(); err != nil {
    log.Fatal(err)
}
defer session.Close()
```

`Open` connects the session to the Discord Gateway.

## Keep the process running

A bot must remain alive to continue receiving Gateway events.

One simple approach for a command-line application is to wait for an interrupt
signal:

```go
stop := make(chan os.Signal, 1)
signal.Notify(stop, os.Interrupt)

<-stop
```

Add the `os/signal` package to your imports.

The complete program is:

```go
package main

import (
    "log"
    "os"
    "os/signal"

    "github.com/bwmarrin/discordgo"
    "github.com/freitaseric/discordkit"
)

func main() {
    token := os.Getenv("DISCORD_TOKEN")
    if token == "" {
        log.Fatal("DISCORD_TOKEN is not set")
    }

    session, err := discordgo.New("Bot " + token)
    if err != nil {
        log.Fatal(err)
    }

    router := discordkit.NewRouter()

    session.AddHandler(router.Handle)

    if err := session.Open(); err != nil {
        log.Fatal(err)
    }
    defer session.Close()

    log.Println("bot connected")

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt)

    <-stop
}
```

Run it with your token available in the environment.

On Linux and macOS:

```bash
export DISCORD_TOKEN="your-token"
go run .
```

On PowerShell:

```powershell
$env:DISCORD_TOKEN="your-token"
go run .
```

If the connection succeeds, the application will remain running until you stop
it.

## Nothing responds yet — and that is expected

The router is connected, but no routes have been registered.

The next step is to define an application command, synchronize it with Discord,
and attach a handler.

Continue with [Your first command](./first-command/).
