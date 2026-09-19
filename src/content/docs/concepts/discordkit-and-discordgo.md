---
title: DiscordKit and discordgo
description: Understand what DiscordKit abstracts, what discordgo continues to provide, and when to use each layer.
---

DiscordKit and discordgo solve different parts of the same problem.

DiscordKit is built **on top of** discordgo and intentionally keeps the underlying library visible.

## The layers

```text
┌───────────────────────────┐
│       Your application    │
├───────────────────────────┤
│         DiscordKit        │
│ Router · Context          │
│ Commands · Components     │
│ Forms · Middleware        │
│ Response lifecycle        │
├───────────────────────────┤
│          discordgo        │
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

`RequireUserOption` returns `*discordgo.User`, channels remain `*discordgo.Channel`, guild members remain `*discordgo.Member`, and your application owns the original `*discordgo.Session`.

This is intentional.

Wrapping these types would create another object model that developers would constantly need to convert back into discordgo values.

## What DiscordKit adds

DiscordKit focuses on patterns above discordgo:

- routing;
- Context helpers;
- typed builders;
- interaction lifecycle management;
- middleware;
- Components V2;
- forms and modal parsing.

## You can always drop down to discordgo

DiscordKit is intentionally not a closed abstraction.

Inside a handler, the underlying session and interaction remain available.

This escape hatch is useful when Discord introduces a new API feature, when discordgo supports something before DiscordKit does, or when you need lower-level control.

## Existing discordgo applications

You do not need to rewrite an existing application.

```go
session.AddHandler(existingMessageHandler)
session.AddHandler(existingReadyHandler)
session.AddHandler(router.Handle)
```

Incremental adoption is supported.

## Rule of thumb

Use DiscordKit when the problem is primarily about **application interaction structure**.

Use discordgo directly when the problem is primarily about **Discord itself**.

In practice, well-designed DiscordKit applications use both.
