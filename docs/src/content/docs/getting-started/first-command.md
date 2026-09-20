---

title: Your first command
description: Define, synchronize, route, and handle your first Discord application command with DiscordKit.
---
In this guide, you will create a `/ping` command and connect its declaration to a DiscordKit handler.

This introduces an important DiscordKit concept:

**declaring a Discord command and routing its interaction are two separate operations.**

## Define the command

DiscordKit provides typed builders for Discord application commands.

Create a command:

```go
pingCommand := discordkit.Command(
    "ping",
    "Check whether the bot is responding",
)
```

This creates a `CommandBuilder`.

At this point, nothing has been sent to Discord yet.

## Build the command

A builder is converted into a standard discordgo application command:

```go
commands, err := discordkit.BuildCommands(
    pingCommand,
)
if err != nil {
    log.Fatal(err)
}
```

`BuildCommands` validates the declarations and returns:

```go
[]*discordgo.ApplicationCommand
```

DiscordKit therefore does not introduce its own command representation at runtime. The final values are normal discordgo types.

## Synchronize the command with Discord

Once the session is connected, synchronize the desired commands:

```go
_, err = discordkit.SyncCommands(
    session,
    session.State.User.ID,
    commands,
    discordkit.SyncOptions{},
)
if err != nil {
    log.Fatal(err)
}
```

`SyncCommands` compares the commands declared by your application with the commands currently registered on Discord.

It can:

* create missing commands;
* update changed commands;
* leave unchanged commands untouched;
* optionally delete obsolete commands.

By default, obsolete commands are not deleted.

To remove commands that are no longer declared:

```go
discordkit.SyncOptions{
    Delete: true,
}
```

## Register the handler

Command synchronization only tells Discord that `/ping` exists.

Your application still needs to decide what should happen when someone uses it.

Register a route:

```go
err = router.Command(
    "ping",
    func(c *discordkit.Context) error {
        return c.ReplyText("Pong!")
    },
)
if err != nil {
    log.Fatal(err)
}
```

When Discord sends a `/ping` interaction, the router matches the command path and invokes the registered handler.

## The complete application

Your program can now look like this:

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

    if err := router.Command(
        "ping",
        func(c *discordkit.Context) error {
            return c.ReplyText("Pong!")
        },
    ); err != nil {
        log.Fatal(err)
    }

    session.AddHandler(router.Handle)

    if err := session.Open(); err != nil {
        log.Fatal(err)
    }
    defer session.Close()

    commands, err := discordkit.BuildCommands(
        discordkit.Command(
            "ping",
            "Check whether the bot is responding",
        ),
    )
    if err != nil {
        log.Fatal(err)
    }

    if _, err := discordkit.SyncCommands(
        session,
        session.State.User.ID,
        commands,
        discordkit.SyncOptions{},
    ); err != nil {
        log.Fatal(err)
    }

    log.Println("bot connected")

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt)
    <-stop
}
```

## Declaration and routing are intentionally separate

These two calls look similar:

```go
discordkit.Command("ping", "Check whether the bot is responding")
```

and:

```go
router.Command("ping", pingHandler)
```

but they solve different problems.

The first declares the command that should exist on Discord.

The second registers the application code that handles its interaction.

Conceptually:

```text
Command declaration
        ↓
BuildCommands
        ↓
SyncCommands
        ↓
Discord knows /ping exists

User executes /ping
        ↓
Discord interaction
        ↓
Router
        ↓
"ping" route
        ↓
Handler
```

Keeping these responsibilities separate makes command declarations testable and allows synchronization to be handled independently from interaction dispatch.

## Guild commands during development

DiscordKit can synchronize commands to a specific guild:

```go
discordkit.SyncOptions{
    GuildID: guildID,
}
```

This is useful during development because guild-scoped application commands generally become available faster than global commands.

The same declaration can later be synchronized globally by removing `GuildID`.

## Synchronization plans

Command synchronization is built around a pure diff operation.

You can inspect what would change without making Discord API requests:

```go
plan := discordkit.DiffCommands(
    remoteCommands,
    desiredCommands,
    true,
)
```

A `SyncPlan` contains:

```go
plan.Create
plan.Update
plan.Unchanged
plan.Delete
```

This is especially useful for tests, tooling, and deployment diagnostics.

## What happens when `/ping` runs?

When a user executes the command:

1. Discord sends an interaction to your application.
2. discordgo receives the interaction.
3. `router.Handle` creates a DiscordKit `Context`.
4. the router resolves the command path as `ping`;
5. matching middleware is applied;
6. your handler runs;
7. `ReplyText` sends the initial interaction response.

You will learn more about this flow in [The interaction model](/concepts/interaction-model/).

## Next step

Continue with [The interaction model](/concepts/interaction-model/) to understand how commands, components, modals, autocomplete, the router, and handlers fit together.
