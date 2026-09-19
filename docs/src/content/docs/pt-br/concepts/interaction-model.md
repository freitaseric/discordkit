---

title: O modelo de interações
description: Entenda como interações do Discord percorrem discordgo, Router, middleware, Context e handlers no DiscordKit.
--------------------------------------------------------------------------------------------------------------------------

Grande parte do DiscordKit é construída em torno de uma ideia:

**o Discord envia interações, e sua aplicação as encaminha para handlers.**

Comandos são apenas um dos tipos de interação.

DiscordKit utiliza o mesmo modelo de aplicação para comandos, componentes, submissões de modais e autocomplete.

## O fluxo completo

Em alto nível:

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

Entender cada etapa facilita bastante o restante do framework.

## O Discord envia uma interação

Interações podem ser geradas por ações como:

* executar um slash command;
* usar um comando de contexto de usuário;
* usar um comando de contexto de mensagem;
* clicar em um botão;
* selecionar uma opção em um select;
* enviar um modal;
* digitar em uma opção com autocomplete.

discordgo recebe essas interações diretamente do Discord.

DiscordKit não substitui essa conexão.

## `Router.Handle` é um handler do discordgo

Um Router do DiscordKit é conectado ao discordgo assim:

```go
session.AddHandler(router.Handle)
```

`Router.Handle` adapta a interação recebida para o modelo de aplicação do DiscordKit.

Ele cria um `Context` e encaminha a interação pelo Router.

## Context representa uma única interação

Todo handler recebe:

```go
*discordkit.Context
```

Um Context existe para exatamente uma interação.

Ele oferece acesso a:

* `*discordgo.Session`;
* interação original;
* opções de comandos;
* usuários, roles, canais, membros e anexos resolvidos;
* parâmetros de rota;
* dados de formulários;
* helpers de resposta;
* `context.Context` padrão do Go.

Exemplo:

```go
func greet(c *discordkit.Context) error {
    user, err := c.RequireUserOption("user")
    if err != nil {
        return err
    }

    return c.ReplyText("Hello, " + user.Username)
}
```

## A resolução da rota depende do tipo da interação

DiscordKit possui quatro categorias principais de rota.

### Comandos

```go
router.Command("ping", pingHandler)
```

Subcomandos usam o caminho completo:

```go
router.Command("admin ban", banHandler)
```

### Componentes

Componentes são roteados através do `custom_id`:

```go
router.Component(
    "/jobs/:jobID/save",
    saveJobHandler,
)
```

Parâmetros nomeados ficam disponíveis no Context:

```go
jobID := c.MustParam("jobID")
```

### Modais

Submissões de modais usam o mesmo modelo de rotas parametrizadas:

```go
router.Modal(
    "/jobs/:jobID/edit",
    editJobHandler,
)
```

### Autocomplete

Rotas de autocomplete combinam o caminho do comando com o nome da opção em foco:

```go
router.Autocomplete(
    "search",
    "query",
    searchAutocomplete,
)
```

## Middleware envolve handlers

Middleware permite executar lógica antes ou depois dos handlers.

DiscordKit utiliza o padrão tradicional de composição do Go:

```go
type Middleware func(Handler) Handler
```

Exemplo:

```go
router := discordkit.NewRouter(
    discordkit.Logging(nil),
    discordkit.RequireGuild(),
)
```

Middleware global é aplicado às rotas correspondentes.

Também é possível adicionar middleware apenas a uma rota:

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

O Router aplica automaticamente o mecanismo de recovery do DiscordKit ao redor da execução.

Se um handler causar panic, ele é convertido em:

```go
*discordkit.PanicError
```

em vez de escapar diretamente do handler de interação.

Recovery **não envia automaticamente** uma mensagem de erro ao usuário.

A aplicação continua responsável por decidir como erros devem ser apresentados.

## Handlers retornam erros

Um Handler possui a forma:

```go
type Handler func(*Context) error
```

Esse modelo permite que erros atravessem middleware e cheguem até o tratamento central do Router.

Por exemplo:

```go
router.OnError(func(c *discordkit.Context, err error) {
    slog.Error("interaction failed", "error", err)
})
```

Sem um hook personalizado, DiscordKit registra o erro usando `slog`.

## Respostas reconhecem a interação

Interações precisam ser reconhecidas conforme as regras da API do Discord.

Um handler pode responder imediatamente:

```go
return c.ReplyText("Done")
```

ou adiar a resposta:

```go
if err := c.Defer(false); err != nil {
    return err
}

// trabalho...

_, err := c.Edit(
    discordkit.MessageSpec{
        Content: "Done",
    },
)

return err
```

DiscordKit acompanha esse estado e impede sequências inválidas, como tentar enviar duas respostas iniciais para a mesma interação.

## Um Router, vários tipos de interação

Uma aplicação maior pode registrar diferentes tipos de interação no mesmo Router:

```go
router := discordkit.NewRouter(
    discordkit.Logging(nil),
)

router.Command(
    "jobs search",
    searchJobs,
)

router.Component(
    "/jobs/:jobID/save",
    saveJob,
)

router.Modal(
    "/jobs/:jobID/edit",
    editJob,
)

router.Autocomplete(
    "jobs search",
    "query",
    autocompleteJobs,
)
```

Todas essas rotas utilizam o mesmo modelo de Context, Handler, middleware, erros e respostas.

Essa consistência é um dos objetivos centrais do DiscordKit.

## Modelo mental

Uma forma útil de pensar no DiscordKit é:

```text
Interação do Discord
        +
informação de roteamento
        +
middleware da aplicação
        =
execução do handler
```

O framework não tenta esconder o Discord.

Ele fornece um caminho estruturado entre a interação recebida e o código da aplicação.

## Próximo passo

Continue em [Router](/pt-br/concepts/router/) para entender registro de rotas, parâmetros, grupos, conflitos e composição de middleware.
