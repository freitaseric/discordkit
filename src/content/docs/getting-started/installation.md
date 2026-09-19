---
title: Installation
description: Install DiscordKit and prepare a Go project to build Discord applications.
---

DiscordKit is distributed as a standard Go module.

## Requirements

DiscordKit currently requires:

- Go 1.26 or newer;
- a Discord application and bot token;
- discordgo through the version pinned by DiscordKit.

You do not need to install discordgo separately when starting a new project.
Go's module system will resolve DiscordKit's dependencies automatically.

## Create a Go project

```bash
mkdir my-discord-bot
cd my-discord-bot
go mod init example.com/my-discord-bot
```

Replace `example.com/my-discord-bot` with the actual module path you intend to use.

## Install DiscordKit

```bash
go get github.com/freitaseric/discordkit
```

Go will add DiscordKit and its required dependencies to `go.mod`.

## Verify the installation

Create `main.go`:

```go
package main

import (
    "fmt"

    "github.com/freitaseric/discordkit"
)

func main() {
    router := discordkit.NewRouter()
    fmt.Printf("%T\n", router)
}
```

Run:

```bash
go run .
```

If the project compiles successfully, DiscordKit is ready to use.

## Using DiscordKit with an existing discordgo project

DiscordKit does not create or replace your `discordgo.Session`.

```go
session, err := discordgo.New("Bot " + token)
if err != nil {
    return err
}

router := discordkit.NewRouter()
session.AddHandler(router.Handle)
```

Existing discordgo handlers and DiscordKit handlers can coexist while you migrate application functionality gradually.

## Development versions

Before DiscordKit reaches a stable `v1` API, new releases may contain breaking changes.

For reproducible builds, commit your `go.mod` and `go.sum` files and use an explicit DiscordKit version when appropriate:

```bash
go get github.com/freitaseric/discordkit@VERSION
```

## Next step

Now build a complete connection to Discord in [Your first bot](./first-bot/).
