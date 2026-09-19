---
title: The interaction model
description: Understand how Discord interactions travel through discordgo, the DiscordKit Router, middleware, Context, and application handlers.
---

Most of DiscordKit is built around one idea:

**Discord sends interactions, and your application routes them to handlers.**

Commands are only one kind of interaction.

DiscordKit uses the same application model for commands, components, modal submissions, and autocomplete.

## The complete flow

```text
Discord
   │
   │ Interaction
   ▼
discordgo.Session
   │
   ▼
Router.Handle
   │
   ▼
Context
   │
   ▼
Route resolution
   │
   ▼
Global middleware
   │
   ▼
Route middleware
   │
   ▼
Handler
   │
   ▼
Response
   │
   ▼
Discord
```

## Route categories

### Commands

```go
router.Command("ping", pingHandler)
router.Command("admin ban", banHandler)
```

### Components

```go
router.Component(
    "/jobs/:jobID/save",
    saveJobHandler,
)
```

### Modal submissions

```go
router.Modal(
    "/jobs/:jobID/edit",
    editJobHandler,
)
```

### Autocomplete

```go
router.Autocomplete(
    "search",
    "query",
    searchAutocomplete,
)
```

## Middleware wraps handlers

```go
type Middleware func(Handler) Handler
```

Global middleware:

```go
router := discordkit.NewRouter(
    discordkit.Logging(nil),
    discordkit.RequireGuild(),
)
```

Route middleware:

```go
router.Command(
    "admin",
    adminHandler,
    discordkit.RequirePermissions(
        discordgo.PermissionAdministrator,
    ),
)
```

## Recovery is always applied

Router dispatch automatically applies DiscordKit recovery.

A handler panic becomes `*discordkit.PanicError`.

Recovery does not automatically send an error response to the user.

## Handlers return errors

```go
type Handler func(*Context) error
```

Errors can propagate through middleware and reach the Router error hook:

```go
router.OnError(func(c *discordkit.Context, err error) {
    slog.Error("interaction failed", "error", err)
})
```

## Responses acknowledge interactions

```go
return c.ReplyText("Done")
```

or:

```go
if err := c.Defer(false); err != nil {
    return err
}

_, err := c.Edit(discordkit.MessageSpec{
    Content: "Done",
})
return err
```

DiscordKit tracks the lifecycle to prevent invalid response sequences.

## Next

Continue with [Router](/concepts/router/).
