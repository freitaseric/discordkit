---
title: "Options and subcommands"
description: "Use typed inputs instead of parsing message text."
---

```go
command := discordkit.Command("greet", "Greet someone",
    discordkit.StringOption("name", "Name").Required().MaxLen(80),
)
commands, err := discordkit.BuildCommands(command)
if err != nil { return err }
_ = commands // Pass to SyncCommands during startup.

if err := router.Command("greet", func(c *discordkit.Context) error {
    name, err := c.RequireString("name")
    if err != nil { return err }
    return c.EphemeralText("Hello, " + name + "!")
}); err != nil { return err }
```
These snippets belong inside your startup function, where `router` already exists. Declare required options before optional ones. `BuildCommands` checks the declaration; `RequireString` validates the interaction input.

For `Command("config", ..., SubCommand("show", ...))`, register the route as `router.Command("config show", handler)`. Route paths use spaces, without the leading slash. Declaring a subcommand does not register its handler.

User and message context menus use `UserCommand` and `MessageCommand` builders. Read their resolved target from the original `c.Interaction` and discordgo types when needed.
