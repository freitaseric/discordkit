---

title: Seu primeiro comando
description: Declare, sincronize, roteie e trate seu primeiro Application Command do Discord com DiscordKit.
------------------------------------------------------------------------------------------------------------

Neste guia, você criará um comando `/ping` e conectará sua declaração a um handler do DiscordKit.

Isso apresenta um conceito importante do framework:

**declarar um comando no Discord e rotear sua interação são duas operações diferentes.**

## Declare o comando

DiscordKit fornece builders tipados para Application Commands.

Crie o comando:

```go
pingCommand := discordkit.Command(
    "ping",
    "Check whether the bot is responding",
)
```

Isso cria um `CommandBuilder`.

Nesse momento, nada foi enviado ao Discord.

## Construa o comando

O builder é convertido em um Application Command normal do discordgo:

```go
commands, err := discordkit.BuildCommands(
    pingCommand,
)
if err != nil {
    log.Fatal(err)
}
```

`BuildCommands` valida as declarações e retorna:

```go
[]*discordgo.ApplicationCommand
```

Portanto, DiscordKit não introduz uma representação própria de comandos em runtime. Os valores finais continuam sendo tipos normais do discordgo.

## Sincronize o comando com o Discord

Depois que a sessão estiver conectada, sincronize os comandos desejados:

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

`SyncCommands` compara os comandos declarados pela aplicação com aqueles atualmente registrados no Discord.

Ele pode:

* criar comandos ausentes;
* atualizar comandos modificados;
* manter comandos inalterados sem novas requisições;
* opcionalmente remover comandos obsoletos.

Por padrão, comandos obsoletos não são removidos.

Para remover comandos que não fazem mais parte da aplicação:

```go
discordkit.SyncOptions{
    Delete: true,
}
```

## Registre o handler

A sincronização apenas informa ao Discord que `/ping` existe.

Sua aplicação ainda precisa definir o que acontecerá quando alguém executar o comando.

Registre uma rota:

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

Quando o Discord enviar uma interação de `/ping`, o Router encontrará o caminho correspondente e executará o handler registrado.

## Aplicação completa

O programa pode ficar assim:

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

    if err := router.Command(
        "ping",
        func(c *discordkit.Context) error {
            return c.ReplyText("Pong!")
        },
    ); err != nil {
        log.Fatal(err)
    }

    session.AddHandler(router.Handle)

    if err := session.Open(); err != nil {
        log.Fatal(err)
    }
    defer session.Close()

    commands, err := discordkit.BuildCommands(
        discordkit.Command(
            "ping",
            "Check whether the bot is responding",
        ),
    )
    if err != nil {
        log.Fatal(err)
    }

    if _, err := discordkit.SyncCommands(
        session,
        session.State.User.ID,
        commands,
        discordkit.SyncOptions{},
    ); err != nil {
        log.Fatal(err)
    }

    log.Println("bot connected")

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt)
    <-stop
}
```

## Declaração e roteamento são propositalmente separados

Estas duas chamadas parecem semelhantes:

```go
discordkit.Command("ping", "Check whether the bot is responding")
```

e:

```go
router.Command("ping", pingHandler)
```

mas resolvem problemas diferentes.

A primeira declara o comando que deve existir no Discord.

A segunda registra o código da aplicação responsável por tratar a interação.

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

Separar essas responsabilidades permite testar declarações de comandos independentemente e tratar a sincronização separadamente do processamento das interações.

## Comandos de guild durante o desenvolvimento

DiscordKit permite sincronizar comandos apenas em uma guild:

```go
discordkit.SyncOptions{
    GuildID: guildID,
}
```

Isso é útil durante o desenvolvimento porque comandos específicos de guild geralmente ficam disponíveis mais rapidamente do que comandos globais.

A mesma declaração pode posteriormente ser sincronizada globalmente removendo `GuildID`.

## Planos de sincronização

A sincronização é baseada em uma operação de diff que não realiza I/O.

Você pode descobrir o que seria alterado sem fazer requisições à API:

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

Isso é especialmente útil em testes, ferramentas internas e diagnóstico de deploy.

## O que acontece quando `/ping` é executado?

Quando o usuário executa o comando:

1. o Discord envia uma interação para sua aplicação;
2. discordgo recebe a interação;
3. `router.Handle` cria um `Context` do DiscordKit;
4. o Router resolve o caminho do comando como `ping`;
5. os middleware correspondentes são aplicados;
6. seu handler é executado;
7. `ReplyText` envia a resposta inicial da interação.

Esse fluxo será explicado em mais detalhes em [O modelo de interações](/pt-br/concepts/interaction-model/).

## Próximo passo

Continue em [O modelo de interações](/pt-br/concepts/interaction-model/) para entender como comandos, componentes, modais, autocomplete, Router e handlers se relacionam.
