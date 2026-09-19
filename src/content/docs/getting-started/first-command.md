---
title: Your first command
description: Define, synchronize, route, and handle your first Discord application command with DiscordKit.
---

In this guide, you will create a `/ping` command and connect its declaration to a DiscordKit handler.

An important DiscordKit concept:

**declaring a Discord command and routing its interaction are two separate operations.**

## Define the command

```go
pingCommand := discordkit.Command(
    "ping",
    "Check whether the bot is responding",
)
```

Nothing has been sent to Discord yet.

## Build the command

```go
commands, err := discordkit.BuildCommands(
    pingCommand,
)
if err != nil {
    log.Fatal(err)
}
```

`BuildCommands` validates declarations and returns standard discordgo commands.

## Synchronize with Discord

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

`SyncCommands` can create missing commands, update changed commands, leave unchanged commands untouched, and optionally delete obsolete commands.

To delete obsolete commands:

```go
discordkit.SyncOptions{
    Delete: true,
}
```

## Register the handler

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

## Declaration and routing are separate

```go
discordkit.Command("ping", "Check whether the bot is responding")
```

declares the command that should exist on Discord.

```go
router.Command("ping", pingHandler)
```

registers the application handler.

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

## Guild commands during development

```go
discordkit.SyncOptions{
    GuildID: guildID,
}
```

Guild-scoped commands are useful during development because they generally become available faster than global commands.

## Synchronization plans

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

## Next step

Continue with [The interaction model](/concepts/interaction-model/).
