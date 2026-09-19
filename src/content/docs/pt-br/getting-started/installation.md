---
title: Instalação
description: Instale o DiscordKit e prepare um projeto Go para desenvolver aplicações para Discord.
---

DiscordKit é distribuído como um módulo Go normal.

## Requisitos

Atualmente, DiscordKit requer:

- Go 1.26 ou superior;
- uma aplicação Discord e um token de bot;
- discordgo através da versão utilizada pelo próprio DiscordKit.

Você não precisa instalar discordgo separadamente ao iniciar um projeto novo.
O sistema de módulos do Go resolverá automaticamente as dependências do DiscordKit.

## Crie um projeto Go

```bash
mkdir my-discord-bot
cd my-discord-bot
go mod init example.com/my-discord-bot
```

Substitua `example.com/my-discord-bot` pelo caminho real do módulo que pretende utilizar.

## Instale o DiscordKit

```bash
go get github.com/freitaseric/discordkit
```

Go adicionará DiscordKit e suas dependências necessárias ao `go.mod`.

## Verifique a instalação

Crie `main.go`:

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

Execute:

```bash
go run .
```

Se o projeto compilar normalmente, DiscordKit está pronto para ser utilizado.

## Usando DiscordKit em um projeto discordgo existente

DiscordKit não cria nem substitui sua `discordgo.Session`.

```go
session, err := discordgo.New("Bot " + token)
if err != nil {
    return err
}

router := discordkit.NewRouter()
session.AddHandler(router.Handle)
```

Handlers existentes do discordgo podem coexistir com handlers gerenciados pelo DiscordKit enquanto a aplicação é migrada aos poucos.

## Versões durante o desenvolvimento

Antes de DiscordKit atingir uma API estável `v1`, novas versões podem introduzir mudanças incompatíveis.

Para manter builds reproduzíveis, mantenha `go.mod` e `go.sum` versionados e, quando necessário, utilize uma versão explícita:

```bash
go get github.com/freitaseric/discordkit@VERSION
```

## Próximo passo

Agora vamos estabelecer uma conexão completa com o Discord em [Seu primeiro bot](./first-bot/).
