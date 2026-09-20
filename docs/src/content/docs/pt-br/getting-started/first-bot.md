---
title: Seu primeiro bot
description: Conecte um bot ao Discord com discordgo e prepare-o para tratar interações com DiscordKit.
---

Neste guia, você criará a menor aplicação DiscordKit realmente útil:
conectar um bot ao Discord e anexar um Router do DiscordKit.

Os comandos serão adicionados no próximo guia.

## Antes de começar

Você precisa de:

- uma aplicação Discord;
- um bot criado para essa aplicação;
- o token do bot;
- DiscordKit instalado no projeto Go.

Mantenha o token privado. Nunca faça commit dele no Git.

Neste guia, armazene-o em uma variável de ambiente chamada `DISCORD_TOKEN`.

## Crie a sessão

Crie `main.go`:

```go
package main

import (
    "log"
    "os"

    "github.com/bwmarrin/discordgo"
    "github.com/freitaseric/discordkit"
)

func main() {
    token := os.Getenv("DISCORD_TOKEN")
    if token == "" {
        log.Fatal("DISCORD_TOKEN is not set")
    }

    session, err := discordgo.New("Bot " + token)
    if err != nil {
        log.Fatal(err)
    }

    router := discordkit.NewRouter()
    session.AddHandler(router.Handle)
}
```

Nesse momento existem três objetos importantes.

### `discordgo.Session`

A sessão controla a conexão com o Discord e expõe Gateway e API REST do discordgo.

### `discordkit.Router`

O Router decide qual handler da aplicação deverá receber uma interação.

### `router.Handle`

`Router.Handle` é um handler normal do discordgo.

Conceitualmente:

```text
Discord
   ↓
discordgo.Session
   ↓
Router.Handle
   ↓
DiscordKit Router
   ↓
Middleware
   ↓
Handler
```

## Abra a conexão

```go
if err := session.Open(); err != nil {
    log.Fatal(err)
}
defer session.Close()
```

## Mantenha o processo em execução

```go
stop := make(chan os.Signal, 1)
signal.Notify(stop, os.Interrupt)
<-stop
```

Adicione `os/signal` aos imports.

O programa completo:

```go
package main

import (
    "log"
    "os"
    "os/signal"

    "github.com/bwmarrin/discordgo"
    "github.com/freitaseric/discordkit"
)

func main() {
    token := os.Getenv("DISCORD_TOKEN")
    if token == "" {
        log.Fatal("DISCORD_TOKEN is not set")
    }

    session, err := discordgo.New("Bot " + token)
    if err != nil {
        log.Fatal(err)
    }

    router := discordkit.NewRouter()
    session.AddHandler(router.Handle)

    if err := session.Open(); err != nil {
        log.Fatal(err)
    }
    defer session.Close()

    log.Println("bot connected")

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt)
    <-stop
}
```

Linux e macOS:

```bash
export DISCORD_TOKEN="your-token"
go run .
```

PowerShell:

```powershell
$env:DISCORD_TOKEN="your-token"
go run .
```

## Ainda não responde — e isso é esperado

O Router está conectado, mas nenhuma rota foi registrada.

Continue em [Seu primeiro comando](./first-command/).
