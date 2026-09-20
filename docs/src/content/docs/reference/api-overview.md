---
title: API overview
description: A map of DiscordKit's builders, routing, responses, components, forms, middleware, and command synchronization.
---

This page maps the package's main APIs. Use the focused guides for [routing](/concepts/router/), [interactions](/concepts/interaction-model/), and [installation](/getting-started/installation/).

## Commands

Build Discord application commands with `Command`, `UserCommand`, and `MessageCommand`. Chat commands accept typed option builders such as `StringOption`, `IntegerOption`, `BooleanOption`, `UserOption`, and `ChannelOption`; subcommands and groups use `SubCommand` and `SubCommandGroup`.

Call `Build` to receive a `*discordgo.ApplicationCommand` and an error, or `MustBuild` for static declarations where invalid definitions should panic. `BuildCommands` builds several declarations and returns the first error. Builders validate Discord's name, description, option count, unique sibling names, nesting, and required-before-optional rules. You still choose when and where to register the resulting commands.

```go
ping, err := discordkit.Command("ping", "Check whether the bot is responding").Build()
if err != nil {
    return err
}
```

## Routing and middleware

A `Router` dispatches slash commands, component interactions, modal submissions, and autocomplete. Register handlers with `Command`, `Component`, `Modal`, and `Autocomplete`, then attach `router.Handle` to discordgo's session using `Session.AddHandler`. A handler has the form `func(*discordkit.Context) error`.

Routes can include named segments in command names or custom IDs; see the [router guide](/concepts/router/). `Group` adds a shared prefix and middleware. Middleware wraps handlers in registration order; `Use` adds router-wide middleware and route registrations can add route-specific middleware. The built-in `Recovery` middleware recovers panics and converts them to errors. Configure `OnError` to observe dispatch errors; without a hook the router logs them.

## Context and responses

A handler's `Context` exposes the discordgo session and interaction, typed option accessors, route parameters, and response helpers. Use `ReplyText` or `Reply` for a first response. Use `Defer` when work may exceed Discord's initial response window, then complete it with `EditReply` or `Followup`. Component handlers can acknowledge with `Update` or `DeferUpdate`. `Ephemeral` controls visibility where supported.

The context tracks acknowledgement state and returns an error when an operation is invalid for the current interaction lifecycle. A defer acknowledges the interaction; it does not make the handler run asynchronously. Your application remains responsible for timeouts, external work, and error policy. See [the interaction model](/concepts/interaction-model/).

## Messages and Components V2

Use `MessageSpec` to describe response content, embeds, flags, files, and components, then pass it to the context response methods. The `Components` DSL includes `Row`, `Button`, link and premium buttons, string/entity/channel selects, sections, containers, text displays, galleries, separators, thumbnails, and file displays. `Raw` wraps a discordgo component when you need a type not covered by a builder.

DiscordKit validates component placement, action-row composition, custom IDs, gallery sizes, and the recursive component budget before sending. Components V2 messages set the required V2 flag automatically when their component tree uses V2 components. Discord's V2 message restrictions still apply; don't mix legacy message content or embeds with a V2 layout. Component interactions are routed by custom ID.

## Modals and form values

Build modal components with labels and supported inputs, then create and send a modal through the context. The package supports text inputs, string selects, and file uploads where allowed by the modal API and the pinned discordgo version. Read submitted values through the context's modal value helpers. Validate the modal tree before sending: each label must contain one supported input and the modal must have at least one component. File upload values refer to attachments supplied by Discord; your app decides how to store or process them.

## Custom IDs

DiscordKit's `CustomID` helper builds and parses slash-delimited IDs with named parameters. Use it to keep component and modal IDs structured, then register the matching route with `Router.Component` or `Router.Modal`. IDs must satisfy Discord's length limit; avoid placing secrets or sensitive user data in them because clients can see them.

## Command synchronization

`DiffCommands` compares desired declarations with commands already fetched from Discord and returns a `SyncPlan` without network I/O. `SyncCommands` fetches the remote set and applies creates and updates; obsolete commands are only deleted when `SyncOptions.Delete` is true. Set `GuildID` for guild-scoped commands; an empty value means global commands. Review deletion behavior carefully before enabling it in production.

## Errors and escape hatches

Builders and validators return errors that can be inspected with `errors.Is` against package sentinel errors such as `ErrInvalidCommand`, `ErrInvalidComponent`, `ErrInvalidModal`, `ErrAlreadyAcknowledged`, and `ErrRouteNotFound`. DiscordKit deliberately exposes discordgo types and the underlying session: use the raw APIs when Discord adds a feature before the DSL supports it.
