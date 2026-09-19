---
title: Introduction
description: Learn what DiscordKit is, what problems it solves, and how it relates to discordgo.
---

DiscordKit is an ergonomic framework layer for building Discord applications
in Go.

It is built **on top of discordgo**, not instead of it.

discordgo remains responsible for connecting to Discord, exposing Discord's
data structures, handling the Gateway and REST API, and representing entities
such as users, members, channels, guilds, interactions, and messages.

DiscordKit focuses on the application layer around those primitives.

## Why DiscordKit exists

A small Discord bot is easy to build directly with discordgo.

As an application grows, however, developers often begin implementing the same
infrastructure repeatedly:

- interaction routing;
- command registration and synchronization;
- typed access to command options;
- custom ID parsing;
- middleware;
- error recovery;
- Components V2 builders;
- modal construction and parsing;
- autocomplete routing;
- response lifecycle tracking.

DiscordKit provides these pieces through one consistent API.

The goal is not to hide Discord. The goal is to make common interaction-driven
application patterns easier to express and harder to misuse.

## What DiscordKit adds

### Unified interaction routing

A single `Router` can dispatch:

- slash commands;
- user and message context commands;
- message components;
- modal submissions;
- autocomplete interactions.

Handlers share the same signature:

```go
func(c *discordkit.Context) error
```

### Typed command builders

Commands and options can be declared through builders:

```go
cmd := discordkit.Command(
    "greet",
    "Greet another user",
    discordkit.UserOption("user", "User to greet").Required(),
)
```

DiscordKit validates common Discord constraints when the command is built.

### Context helpers

Handlers receive a `*discordkit.Context`, which provides typed access to
interaction data:

```go
user, err := c.RequireUserOption("user")
if err != nil {
    return err
}
```

Resolved Discord objects are read from the interaction data without performing
implicit REST requests.

### Components V2

DiscordKit includes a DSL for building Discord Components V2:

```go
discordkit.Container(
    discordkit.Text("## Deployment complete"),
    discordkit.Row(
        discordkit.Button("Open logs", "logs:open"),
    ),
)
```

The resulting values are still standard discordgo component types.

### Forms and modals

Modal definitions use typed builders and submissions can be read without
reflection:

```go
form := c.Form()
subject, ok := form.String("subject")
```

### Middleware

Middleware uses the familiar Go handler-wrapping pattern:

```go
type Middleware func(Handler) Handler
```

DiscordKit includes middleware for recovery, logging, guild requirements, and
permission checks, while allowing applications to define their own.

### Response lifecycle safety

Discord interactions have strict acknowledgement rules.

DiscordKit tracks whether an interaction is still pending, has already received
an initial response, or has been deferred.

Invalid operations return meaningful errors such as:

```go
discordkit.ErrAlreadyResponded
discordkit.ErrNotResponded
```

instead of relying entirely on later Discord API failures.

## DiscordKit does not replace discordgo

DiscordKit intentionally keeps discordgo visible.

A user remains:

```go
*discordgo.User
```

A channel remains:

```go
*discordgo.Channel
```

Your application still creates and owns:

```go
*discordgo.Session
```

And handlers can access the underlying session and interaction whenever they
need functionality that DiscordKit does not abstract.

This makes DiscordKit suitable both for new applications and for existing
discordgo projects that want a higher-level application structure.

## Next step

Continue with [Installation](./installation/) to add DiscordKit to a Go project.
