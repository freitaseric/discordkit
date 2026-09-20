---
title: "Environment variables"
description: "Read configuration once and fail early when it is missing."
---

The cookbook requires two variables:

| Variable | Value |
| --- | --- |
| `DISCORD_TOKEN` | Bot token from the Developer Portal |
| `DISCORD_GUILD_ID` | Test server ID; not a channel or application ID |

```go
 token := strings.TrimSpace(os.Getenv("DISCORD_TOKEN"))
 if token == "" {
     return fmt.Errorf("DISCORD_TOKEN is required")
 }
```

Import `strings`, `os` and `fmt`. Go's `os.Getenv` reads the process environment; it does **not** load `.env` files. The [first bot tutorial](/getting-started/first-bot/) shows loading them in Bash and PowerShell.

Commit an `.env.example` with empty values and ignore `.env` and `.env.*`, except `.env.example`. In hosting, set secrets through the service manager or platform. Never log tokens. Reset an exposed token in the Developer Portal.
