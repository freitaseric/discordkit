---

title: DiscordKit and discordgo
description: Understand what DiscordKit abstracts, what discordgo continues to provide, and when to use each layer.
-------------------------------------------------------------------------------------------------------------------

DiscordKit and discordgo solve different parts of the same problem.

DiscordKit is built **on top of** discordgo and intentionally keeps the
underlying library visible.

Understanding this relationship makes the rest of the framework much easier to
reason about.

## The layers

A simplified DiscordKit application looks like this:

```text
┌───────────────────────────┐
│       Your application    │
├───────────────────────────┤
│         DiscordKit        │
│                           │
│ Router · Context          │
│ Commands · Components     │
│ Forms · Middleware        │
│ Response lifecycle        │
├───────────────────────────┤
│          discordgo        │
│                           │
│ Gateway · REST            │
│ Sessions · Discord types  │
├───────────────────────────┤
│        Discord API        │
└───────────────────────────┘
```

discordgo provides the protocol and data-model layer.

DiscordKit provides application structure around interactions.

## What remains discordgo

DiscordKit does not define alternative models for Discord entities.

For example:

```go
user, err := c.RequireUserOption("user")
```

returns:

```go
*discordgo.User
```

A resolved channel is:

```go
*discordgo.Channel
```

A guild member is:

```go
*discordgo.Member
```

And your application owns the original:

```go
*discordgo.Session
```

This is intentional.

Wrapping all of these types would create another object model that developers
would constantly need to convert back into discordgo values.

DiscordKit avoids that duplication.

## What DiscordKit adds

DiscordKit focuses on patterns that tend to live above discordgo.

### Routing

Instead of manually branching on interaction types and custom IDs, register
handlers:

```go
router.Command("ping", pingHandler)

router.Component(
    "/jobs/:jobID/save",
    saveJobHandler,
)

router.Modal(
    "/jobs/:jobID/edit",
    editJobHandler,
)
```

### Context

Raw Discord interaction data is wrapped in an application-oriented `Context`.

For example:

```go
name, ok := c.String("name")

user, err := c.RequireUserOption("user")

jobID := c.MustParam("jobID")
```

### Builders

DiscordKit provides builders for application commands, options, Components V2,
and modal forms.

The result is still compatible with discordgo.

### Lifecycle management

DiscordKit tracks interaction response state and prevents common invalid
sequences.

For example, sending two initial responses returns:

```go
discordkit.ErrAlreadyResponded
```

### Middleware

Reusable application behavior can be applied consistently around handlers:

```go
router := discordkit.NewRouter(
    discordkit.Recovery(),
    discordkit.Logging(nil),
)
```

## You can always drop down to discordgo

DiscordKit is intentionally not a closed abstraction.

Inside a handler, the underlying session and interaction remain available.

This means a DiscordKit application can use any discordgo functionality even
when DiscordKit does not provide a dedicated abstraction for it.

This design is sometimes called an **escape hatch**.

It is particularly useful when:

* Discord introduces a new API feature;
* discordgo supports something before DiscordKit does;
* you need low-level REST functionality;
* an application has unusual requirements that do not fit a framework helper.

## Existing discordgo applications

You do not need to rewrite an existing discordgo application to adopt
DiscordKit.

The router itself is attached as a regular discordgo handler:

```go
session.AddHandler(router.Handle)
```

Existing handlers can remain registered:

```go
session.AddHandler(existingMessageHandler)
session.AddHandler(existingReadyHandler)
session.AddHandler(router.Handle)
```

This makes incremental adoption possible.

## A useful rule of thumb

Use DiscordKit when the problem is primarily about **application interaction
structure**:

* routes;
* handlers;
* command declarations;
* components;
* forms;
* middleware;
* response state.

Use discordgo directly when the problem is primarily about **Discord itself**:

* sessions;
* Gateway events;
* Discord entities;
* REST endpoints;
* low-level or newly introduced Discord features.

In practice, well-designed DiscordKit applications use both.
