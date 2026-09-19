---
title: Router
description: Entenda como o DiscordKit resolve comandos, componentes, modais, autocomplete, middleware, grupos e parâmetros de rota.
---

O `Router` é o componente central de roteamento de interações do DiscordKit.

Ele conecta interações recebidas do Discord aos handlers da aplicação.

Um único Router pode tratar Application Commands, componentes, submissões de modais e autocomplete.

## Criando um Router

```go
router := discordkit.NewRouter()
```

Com middleware global:

```go
router := discordkit.NewRouter(
    discordkit.Logging(nil),
    discordkit.RequireGuild(),
)
```

## Conectando ao discordgo

```go
session.AddHandler(router.Handle)
```

## Rotas de comandos

```go
router.Command("ping", pingHandler)
router.Command("admin ban", banHandler)
router.Command("config channel set", setChannelHandler)
```

Caminhos de comandos usam espaços.

## Rotas de componentes

Literal:

```go
router.Component("settings:save", saveSettings)
```

Parametrizada:

```go
router.Component(
    "/jobs/:jobID/save",
    saveJob,
)
```

Para `/jobs/42/save`:

```go
jobID := c.MustParam("jobID")
```

## Rotas de modais

```go
router.Modal(
    "/jobs/:jobID/edit",
    updateJob,
)
```

## Rotas de autocomplete

```go
router.Autocomplete(
    "jobs search",
    "query",
    autocompleteJobs,
)
```

## Middleware por rota

```go
router.Command(
    "admin ban",
    banUser,
    discordkit.RequireGuild(),
    discordkit.RequirePermissions(
        discordgo.PermissionBanMembers,
    ),
)
```

## Adicionando middleware global depois

```go
router.Use(
    discordkit.Logging(nil),
)
```

Middleware adicionado com `Use` se aplica às rotas registradas antes e depois dele.

## Groups

```go
admin := router.Group(
    "admin",
    discordkit.RequireGuild(),
)

admin.Command(
    "ban",
    banHandler,
)
```

Groups compartilham prefixos e middleware, mas registram suas rotas no Router original.

## Parâmetros de rota

```text
/jobs/:jobID/save
```

Acesse com:

```go
jobID, ok := c.Param("jobID")
jobID, err := c.RequireParam("jobID")
jobID := c.MustParam("jobID")
```

## Conflitos de rota

DiscordKit retorna `ErrRouteConflict` quando rotas entram em conflito.

Estes padrões são estruturalmente equivalentes:

```text
/jobs/:jobID/save
/jobs/:id/save
```

Sempre trate os erros de registro durante o startup.

## Rotas não encontradas

Interações válidas sem rota correspondente retornam `ErrRouteNotFound`.

## Tratamento de erros

```go
router.OnError(func(
    c *discordkit.Context,
    err error,
) {
    slog.Error("interaction failed", "error", err)
})
```

Sem hook configurado, DiscordKit utiliza `slog`.

## Recovery

O dispatch do Router aplica `Recovery()` automaticamente.

Um panic vira `*discordkit.PanicError`.

## Dispatch direto

Para testes ou integrações em baixo nível:

```go
err := router.Dispatch(ctx)
```

`Dispatch` não executa `OnError`; ele retorna o erro diretamente.

## Próximo passo

Continue com Context para aprender como handlers acessam dados das interações e parâmetros de rota.
