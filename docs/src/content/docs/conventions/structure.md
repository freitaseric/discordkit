---
title: "Project structure"
description: "Organize a bot as it grows beyond main.go."
---

Start with one `main.go` while learning. When handlers grow, use ordinary Go packages:

| Path | Responsibility |
| --- | --- |
| `cmd/bot/main.go` | Read configuration, connect the session and wait for shutdown |
| `internal/discord/commands.go` | Command declarations and route registration |
| `internal/discord/components.go` | Buttons, selects and modal handlers |
| `internal/service/` | Business rules independent of Discord |
| `internal/store/` | Persistence and queries |

DiscordKit does not scan directories or automatically register handlers. Export a registration function from your package and call it explicitly before `session.Open()`.

```go title="internal/discord/commands.go"
package discord

import "github.com/freitaseric/discordkit"

func Register(r *discordkit.Router) error {
    return r.Command("ping", func(c *discordkit.Context) error {
        return c.ReplyText("Pong!")
    })
}
```

Import this package using your own module path, for example `example.com/my-bot/internal/discord`. Keep configuration at startup and inject services into handlers. Avoid mutable package globals: discordgo can invoke handlers concurrently.
