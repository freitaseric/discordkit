---

title: Introdução
description: Entenda o que é o DiscordKit, quais problemas ele resolve e como ele se relaciona com o discordgo.
---------------------------------------------------------------------------------------------------------------

DiscordKit é uma camada de framework ergonômica para o desenvolvimento de
aplicações para Discord em Go.

Ele é construído **sobre o discordgo**, e não no lugar dele.

O discordgo continua responsável pela conexão com o Discord, pela exposição das
estruturas de dados da plataforma, pelo Gateway e pela API REST, além da
representação de entidades como usuários, membros, canais, servidores,
interações e mensagens.

DiscordKit atua principalmente na camada de aplicação construída sobre esses
recursos.

## Por que o DiscordKit existe

Um bot pequeno para Discord pode ser desenvolvido diretamente com discordgo sem
grandes dificuldades.

Conforme a aplicação cresce, porém, é comum que desenvolvedores comecem a
implementar repetidamente a mesma infraestrutura:

* roteamento de interações;
* registro e sincronização de comandos;
* leitura tipada de opções;
* interpretação de custom IDs;
* middleware;
* recuperação de erros e panics;
* builders para Components V2;
* criação e leitura de modais;
* roteamento de autocomplete;
* controle do ciclo de vida das respostas.

DiscordKit fornece essas funcionalidades através de uma API consistente.

O objetivo não é esconder o Discord. O objetivo é tornar padrões comuns de
aplicações orientadas a interações mais fáceis de expressar e mais difíceis de
usar incorretamente.

## O que o DiscordKit adiciona

### Roteamento unificado de interações

Um único `Router` pode encaminhar:

* slash commands;
* comandos de contexto de usuário e mensagem;
* componentes;
* submissões de modais;
* interações de autocomplete.

Todos os handlers utilizam a mesma assinatura:

```go
func(c *discordkit.Context) error
```

### Builders tipados para comandos

Comandos e opções podem ser declarados através de builders:

```go
cmd := discordkit.Command(
    "greet",
    "Greet another user",
    discordkit.UserOption("user", "User to greet").Required(),
)
```

DiscordKit valida diversas restrições da API do Discord durante a construção do
comando.

### Helpers de Context

Cada handler recebe um `*discordkit.Context`, que oferece acesso tipado aos
dados da interação:

```go
user, err := c.RequireUserOption("user")
if err != nil {
    return err
}
```

Objetos resolvidos do Discord são obtidos diretamente dos dados enviados na
interação, sem requisições REST implícitas.

### Components V2

DiscordKit oferece uma DSL para construção de Components V2:

```go
discordkit.Container(
    discordkit.Text("## Deployment complete"),
    discordkit.Row(
        discordkit.Button("Open logs", "logs:open"),
    ),
)
```

Os componentes resultantes continuam sendo tipos normais do discordgo.

### Formulários e modais

Modais podem ser definidos através de builders tipados e suas submissões podem
ser acessadas sem reflection:

```go
form := c.Form()

subject, ok := form.String("subject")
```

### Middleware

O modelo de middleware segue o padrão familiar de composição de handlers em Go:

```go
type Middleware func(Handler) Handler
```

DiscordKit inclui middleware para recuperação de panics, logging, exigência de
guild e verificação de permissões, além de permitir middleware definido pela
própria aplicação.

### Segurança no ciclo de vida das respostas

Interações do Discord possuem regras rígidas de reconhecimento e resposta.

DiscordKit acompanha se a interação ainda está pendente, já recebeu uma resposta
inicial ou foi adiada.

Operações inválidas retornam erros claros como:

```go
discordkit.ErrAlreadyResponded
discordkit.ErrNotResponded
```

em vez de depender apenas de erros posteriores retornados pela API do Discord.

## DiscordKit não substitui discordgo

DiscordKit mantém o discordgo propositalmente acessível.

Um usuário continua sendo:

```go
*discordgo.User
```

Um canal continua sendo:

```go
*discordgo.Channel
```

Sua aplicação continua criando e controlando:

```go
*discordgo.Session
```

E os handlers podem acessar diretamente a sessão e a interação original sempre
que precisarem utilizar algum recurso que o DiscordKit não abstraia.

Isso permite utilizar DiscordKit tanto em aplicações novas quanto em projetos
discordgo existentes que desejam uma camada de aplicação mais estruturada.

## Próximo passo

Continue em [Instalação](./installation/) para adicionar o DiscordKit a um
projeto Go.
