---
title: Router
description: Learn how DiscordKit resolves commands, components, modals, autocomplete interactions, middleware, groups, and route parameters.
---

The `Router` is the central interaction dispatcher in DiscordKit.

It maps incoming Discord interactions to application handlers.

A single Router can handle application commands, message components, modal submissions, and autocomplete interactions.

## Create a Router

```go
router := discordkit.NewRouter()
```

With global middleware:

```go
router := discordkit.NewRouter(
    discordkit.Logging(nil),
    discordkit.RequireGuild(),
)
```

## Attach it to discordgo

```go
session.AddHandler(router.Handle)
```

## Command routes

```go
router.Command("ping", pingHandler)
router.Command("admin ban", banHandler)
router.Command("config channel set", setChannelHandler)
```

Command paths use spaces.

## Component routes

Literal:

```go
router.Component("settings:save", saveSettings)
```

Parameterized:

```go
router.Component(
    "/jobs/:jobID/save",
    saveJob,
)
```

For `/jobs/42/save`:

```go
jobID := c.MustParam("jobID")
```

## Modal routes

```go
router.Modal(
    "/jobs/:jobID/edit",
    updateJob,
)
```

## Autocomplete routes

```go
router.Autocomplete(
    "jobs search",
    "query",
    autocompleteJobs,
)
```

## Route-specific middleware

```go
router.Command(
    "admin ban",
    banUser,
    discordkit.RequireGuild(),
    discordkit.RequirePermissions(
        discordgo.PermissionBanMembers,
    ),
)
```

## Add global middleware later

```go
router.Use(
    discordkit.Logging(nil),
)
```

Middleware added with `Use` applies to routes registered before and after it.

## Groups

```go
admin := router.Group(
    "admin",
    discordkit.RequireGuild(),
)

admin.Command(
    "ban",
    banHandler,
)
```

Groups share prefixes and middleware but register routes into the original Router.

## Route parameters

```text
/jobs/:jobID/save
```

Access with:

```go
jobID, ok := c.Param("jobID")
jobID, err := c.RequireParam("jobID")
jobID := c.MustParam("jobID")
```

## Route conflicts

DiscordKit returns `ErrRouteConflict` when routes conflict.

These patterns are structurally equivalent and conflict:

```text
/jobs/:jobID/save
/jobs/:id/save
```

Always check registration errors during startup.

## Missing routes

Unmatched valid interactions return `ErrRouteNotFound`.

## Error handling

```go
router.OnError(func(
    c *discordkit.Context,
    err error,
) {
    slog.Error("interaction failed", "error", err)
})
```

If no hook is configured, DiscordKit logs using `slog`.

## Recovery

Router dispatch automatically applies `Recovery()`.

A panic becomes `*discordkit.PanicError`.

## Direct dispatch

For tests or low-level integrations:

```go
err := router.Dispatch(ctx)
```

`Dispatch` does not invoke `OnError`; it returns the error directly.

## Next step

Continue with Context to learn how handlers access interaction data and route parameters.
