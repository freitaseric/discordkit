---
title: "The project: a support desk"
description: "What we will build, how to learn, and how to run each stage."
---

Build a support desk: members open requests, staff claim them, the owner or staff close them, and owners rate the outcome. Records survive process restarts.

This is a **ticket registry inside Discord with private interfaces**. It does not create private channels, relay a live conversation, or automatically notify requesters. Staff inspect the queue and coordinate support in the server.

## Learning path

You should be comfortable creating files, running terminal commands, and reading Go functions, structs and errors. The guide explains application structure and each API used.

Use the **guided mode** first: clone the project, start at stage 1, read and modify the indicated modules, then unlock the next stage. The complete skeleton is present so every checkpoint compiles. Later, rebuild it in your own Go module using the complete file listings and the module extraction instructions in Operations.

`COOKBOOK_STAGE` selects command and route registration. It is a teaching switch, not a DiscordKit feature or a database migration mechanism. It defaults to `6`. Stop the process before restarting with another stage. Going backwards does not delete previously registered Discord commands: sync is deliberately non-destructive.

| Chapter | Runnable stage | Outcome |
| --- | --- | --- |
| [01 · Setup](/cookbook/setup/) | 1 | Installed application and working `/ping` |
| [02 · Architecture](/cookbook/architecture/) | 1 | Configuration, lifecycle and packages |
| [03 · Commands](/cookbook/commands/) | 2 | Typed options and response visibility |
| [04 · Panel](/cookbook/components/) | 3 | Public help panel and authorization |
| [05 · JSON store](/cookbook/persistence/) | 3 + tests | Persistent repository and access rules |
| [06 · Tickets](/cookbook/tickets/) | 4 | Forms, routing, claiming and closing |
| [07 · Queue](/cookbook/queue/) | 5 | Filters, pagination and autocomplete |
| [08 · Operations](/cookbook/operations/) | 6 | Ratings, exports, tests and hosting |
| [09 · Laboratory](/cookbook/laboratory/) | 6 | Resolved entities, context menus and media |
| [API map](/cookbook/api-map/) | Reference | API families, examples and boundaries |

Go provides packages, structs, concurrency and files. discordgo connects to the Gateway and exposes Discord types and endpoints. DiscordKit supplies interaction builders, routing, middleware and response handling. The JSON store belongs to this application; it is not a new database abstraction in DiscordKit. Dependencies are passed explicitly by `main`.

You are done when you can create, inspect, claim, close and rate a ticket, reject unauthorized access, restart without losing data, and export only authorized records. You should also be able to explain why `ShowModal` cannot follow `Defer` and why a custom ID is not authorization.

[Complete executable source](https://github.com/freitaseric/discordkit/tree/main/examples/community-bot). Tutorial source blocks are synchronized with the tested example in CI.

---

[Next →: 01 · Prepare and connect the bot](/cookbook/setup/)
