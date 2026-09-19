---
title: O modelo de interações
description: Entenda como interações do Discord percorrem discordgo, Router, middleware, Context e handlers no DiscordKit.
---

Grande parte do DiscordKit é construída em torno de uma ideia:

**o Discord envia interações, e sua aplicação as encaminha para handlers.**

Comandos são apenas um dos tipos de interação.

DiscordKit utiliza o mesmo modelo para comandos, componentes, submissões de modais e autocomplete.

## O fluxo completo

```text
Discord
   │
   │ Interação
   ▼
discordgo.Session
   │
   ▼
Router.Handle
   │
   ▼
Context
   │
   ▼
Resolução da rota
   │
   ▼
Middleware global
   │
   ▼
Middleware da rota
   │
   ▼
Handler
   │
   ▼
Resposta
   │
   ▼
Discord
```

## Categorias de rota

### Comandos

```go
router.Command("ping", pingHandler)
router.Command("admin ban", banHandler)
```

### Componentes

```go
router.Component(
    "/jobs/:jobID/save",
    saveJobHandler,
)
```

### Modais

```go
router.Modal(
    "/jobs/:jobID/edit",
    editJobHandler,
)
```

### Autocomplete

```go
router.Autocomplete(
    "search",
    "query",
    searchAutocomplete,
)
```

## Middleware envolve handlers

```go
type Middleware func(Handler) Handler
```

Middleware global:

```go
router := discordkit.NewRouter(
    discordkit.Logging(nil),
    discordkit.RequireGuild(),
)
```

Middleware de rota:

```go
router.Command(
    "admin",
    adminHandler,
    discordkit.RequirePermissions(
        discordgo.PermissionAdministrator,
    ),
)
```

## Recovery sempre é aplicado

O dispatch do Router aplica Recovery automaticamente.

Um panic em handler vira `*discordkit.PanicError`.

Recovery não envia automaticamente uma mensagem de erro ao usuário.

## Handlers retornam erros

```go
type Handler func(*Context) error
```

Erros podem atravessar middleware e chegar ao hook central:

```go
router.OnError(func(c *discordkit.Context, err error) {
    slog.Error("interaction failed", "error", err)
})
```

## Respostas reconhecem a interação

```go
return c.ReplyText("Done")
```

ou:

```go
if err := c.Defer(false); err != nil {
    return err
}

_, err := c.Edit(discordkit.MessageSpec{
    Content: "Done",
})
return err
```

DiscordKit acompanha o ciclo de vida para impedir sequências inválidas.

## Próximo passo

Continue em [Router](/pt-br/concepts/router/).
