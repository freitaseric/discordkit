---

title: Router
description: Learn how DiscordKit resolves commands, components, modals, autocomplete interactions, middleware, groups, and route parameters.
---
The `Router` is the central interaction dispatcher in DiscordKit.

It maps incoming Discord interactions to application handlers.

A single Router can handle:

* application commands;
* message components;
* modal submissions;
* autocomplete interactions.

## Create a Router

The simplest Router is:

```go
router := discordkit.NewRouter()
```

Global middleware can be provided when the Router is created:

```go
router := discordkit.NewRouter(
    discordkit.Logging(nil),
    discordkit.RequireGuild(),
)
```

Middleware runs in the order it is supplied.

## Attach it to discordgo

The Router is connected to a normal discordgo session:

```go
session.AddHandler(router.Handle)
```

`router.Handle` receives interactions from discordgo, creates a DiscordKit `Context`, resolves a route, applies middleware, and invokes the corresponding handler.

## Command routes

Register a command handler with:

```go
err := router.Command(
    "ping",
    pingHandler,
)
```

Command paths use spaces.

For example, a subcommand:

```text
/admin ban
```

is registered as:

```go
router.Command(
    "admin ban",
    banHandler,
)
```

A subcommand inside a group follows the same pattern:

```text
/config channel set
```

becomes:

```go
router.Command(
    "config channel set",
    setChannelHandler,
)
```

The route path must be canonical. Extra or repeated whitespace is rejected.

## Component routes

Components are routed using their `custom_id`.

A literal route:

```go
router.Component(
    "settings:save",
    saveSettings,
)
```

matches exactly that custom ID.

DiscordKit also supports path-style parameters:

```go
router.Component(
    "/jobs/:jobID/save",
    saveJob,
)
```

A custom ID such as:

```text
/jobs/42/save
```

matches the route and exposes:

```go
jobID, ok := c.Param("jobID")
```

or:

```go
jobID := c.MustParam("jobID")
```

## Modal routes

Modal submissions use the same parameterized routing model:

```go
router.Modal(
    "/jobs/:jobID/edit",
    updateJob,
)
```

This allows the same resource identifier to travel through a component and its modal submission.

For example:

```text
/jobs/42/edit
```

can identify both the button that opens a form and the form submission that updates job `42`.

## Autocomplete routes

Autocomplete routes use the command path plus the name of the focused option.

For a command such as:

```text
/jobs search
```

with an autocomplete-enabled `query` option:

```go
router.Autocomplete(
    "jobs search",
    "query",
    autocompleteJobs,
)
```

The Router only invokes that handler when `query` is the focused autocomplete option.

## Route-specific middleware

Middleware can be attached directly to a route:

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

This middleware runs in addition to the Router's global middleware.

## Add global middleware later

Global middleware can also be added after Router creation:

```go
router.Use(
    discordkit.Logging(nil),
)
```

Middleware added through `Use` applies to previously registered routes as well as routes registered later.

## Groups

Groups combine a shared route prefix with shared middleware.

For example:

```go
admin := router.Group(
    "admin",
    discordkit.RequireGuild(),
)
```

Commands registered through this group inherit the prefix:

```go
admin.Command(
    "ban",
    banHandler,
)
```

registers:

```text
admin ban
```

The same Group can also register components and modals.

For component and modal paths, prefixes use `/` instead of spaces.

Example:

```go
jobs := router.Group(
    "/jobs",
)

jobs.Component(
    ":jobID/save",
    saveJob,
)
```

Conceptually, this produces a route under the `/jobs` prefix.

Groups do not create independent Router instances.

They register routes into the original Router.

## Route parameters

Parameters are identified by a leading colon:

```text
/jobs/:jobID/save
```

When the route matches:

```text
/jobs/123/save
```

the Context contains:

```go
jobID, ok := c.Param("jobID")
```

For required parameters:

```go
jobID, err := c.RequireParam("jobID")
```

or:

```go
jobID := c.MustParam("jobID")
```

Use `MustParam` when the route definition itself guarantees that the parameter must exist and failure represents a programmer error.

## Literal segments are more specific

Suppose an application has conceptually similar routes:

```text
/jobs/:jobID
/jobs/new
```

Literal segments are more specific than parameter segments during route selection.

This helps avoid generic parameterized routes capturing interactions intended for more specific paths.

You should still keep route structures explicit and easy to reason about rather than relying heavily on overlapping patterns.

## Route conflicts

DiscordKit detects conflicting routes during registration.

For example, registering the same command route twice returns:

```go
discordkit.ErrRouteConflict
```

Component and modal patterns that are structurally indistinguishable also conflict.

For example, these represent the same routing shape:

```text
/jobs/:jobID/save
/jobs/:id/save
```

The parameter name is different, but both match the same set of custom IDs.

DiscordKit rejects this ambiguity.

Always check registration errors during application startup:

```go
if err := router.Command(
    "ping",
    pingHandler,
); err != nil {
    log.Fatal(err)
}
```

A route conflict is normally a configuration or programming error and should fail early.

## Missing routes

If DiscordKit receives an interaction that it understands but cannot match to a registered route, dispatch returns:

```go
discordkit.ErrRouteNotFound
```

When using `router.Handle`, the error is forwarded to the Router's error handling mechanism.

## Error handling

Configure a central error hook with:

```go
router.OnError(func(
    c *discordkit.Context,
    err error,
) {
    slog.Error(
        "interaction failed",
        "error",
        err,
    )
})
```

If no hook is configured, DiscordKit logs the failure using `slog`.

The error hook is intended for central application error handling.

It does not automatically turn every error into a Discord response.

This distinction is important because the correct user-facing behavior depends on whether the interaction has already been acknowledged.

## Recovery

The Router applies `Recovery()` automatically around dispatch.

This means a panic in a handler is converted into a `*discordkit.PanicError`.

You may also explicitly use `Recovery()` when composing middleware elsewhere, but adding it to `NewRouter` is generally unnecessary because Router dispatch already guarantees recovery.

## Dispatch directly

Most applications should use:

```go
router.Handle
```

through discordgo.

For testing or lower-level integrations, the Router also exposes:

```go
err := router.Dispatch(ctx)
```

`Dispatch` resolves and executes the matching route but does not invoke the Router's `OnError` hook.

The returned error remains under the caller's control.

This makes it useful in tests.

## Thread safety

Route registration, middleware changes, error hook configuration, and dispatch are protected for concurrent use.

In practice, applications should still prefer registering routes during startup and treating the Router configuration as mostly immutable while the bot is running.

## A larger example

```go
router := discordkit.NewRouter(
    discordkit.Logging(nil),
)

router.OnError(func(
    c *discordkit.Context,
    err error,
) {
    slog.Error("interaction failed", "error", err)
})

jobs := router.Group(
    "jobs",
    discordkit.RequireGuild(),
)

if err := jobs.Command(
    "search",
    searchJobs,
); err != nil {
    log.Fatal(err)
}

if err := router.Component(
    "/jobs/:jobID/save",
    saveJob,
); err != nil {
    log.Fatal(err)
}

if err := router.Modal(
    "/jobs/:jobID/edit",
    editJob,
); err != nil {
    log.Fatal(err)
}

if err := router.Autocomplete(
    "jobs search",
    "query",
    autocompleteJobs,
); err != nil {
    log.Fatal(err)
}
```

The resulting application uses one routing model for several Discord interaction types.

## Next step

Continue with [Context](/reference/api-overview/) to learn how handlers access interaction data, route parameters, users, channels, roles, attachments, and request-scoped values.
