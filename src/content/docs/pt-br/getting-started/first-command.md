---
title: Seu primeiro comando
description: Declare, sincronize, roteie e trate seu primeiro Application Command do Discord com DiscordKit.
---

Neste guia, você criará um comando `/ping` e conectará sua declaração a um handler do DiscordKit.

Um conceito importante:

**declarar um comando no Discord e rotear sua interação são duas operações diferentes.**

## Declare o comando

```go
pingCommand := discordkit.Command(
    "ping",
    "Check whether the bot is responding",
)
```

Nesse momento nada foi enviado ao Discord.

## Construa o comando

```go
commands, err := discordkit.BuildCommands(
    pingCommand,
)
if err != nil {
    log.Fatal(err)
}
```

`BuildCommands` valida as declarações e retorna comandos normais do discordgo.

## Sincronize com o Discord

```go
_, err = discordkit.SyncCommands(
    session,
    session.State.User.ID,
    commands,
    discordkit.SyncOptions{},
)
if err != nil {
    log.Fatal(err)
}
```

`SyncCommands` pode criar comandos ausentes, atualizar comandos modificados, manter os inalterados e opcionalmente remover comandos obsoletos.

Para remover comandos obsoletos:

```go
discordkit.SyncOptions{
    Delete: true,
}
```

## Registre o handler

```go
err = router.Command(
    "ping",
    func(c *discordkit.Context) error {
        return c.ReplyText("Pong!")
    },
)
if err != nil {
    log.Fatal(err)
}
```

## Declaração e roteamento são separados

```go
discordkit.Command("ping", "Check whether the bot is responding")
```

declara o comando que deve existir no Discord.

```go
router.Command("ping", pingHandler)
```

registra o handler da aplicação.

Conceitualmente:

```text
Declaração do comando
        ↓
BuildCommands
        ↓
SyncCommands
        ↓
Discord sabe que /ping existe

Usuário executa /ping
        ↓
Interação do Discord
        ↓
Router
        ↓
rota "ping"
        ↓
Handler
```

## Comandos de guild durante o desenvolvimento

```go
discordkit.SyncOptions{
    GuildID: guildID,
}
```

Comandos específicos de guild são úteis durante o desenvolvimento porque geralmente ficam disponíveis mais rapidamente que comandos globais.

## Planos de sincronização

```go
plan := discordkit.DiffCommands(
    remoteCommands,
    desiredCommands,
    true,
)
```

Um `SyncPlan` contém:

```go
plan.Create
plan.Update
plan.Unchanged
plan.Delete
```

## Próximo passo

Continue em [O modelo de interações](/pt-br/concepts/interaction-model/).
