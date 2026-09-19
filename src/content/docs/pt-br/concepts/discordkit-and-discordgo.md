---
title: DiscordKit e discordgo
description: Entenda o que o DiscordKit abstrai, o que continua sendo responsabilidade do discordgo e quando usar cada camada.
---

DiscordKit e discordgo resolvem partes diferentes do mesmo problema.

DiscordKit é construído **sobre o discordgo** e mantém a biblioteca base propositalmente visível.

## As camadas

```text
┌───────────────────────────┐
│        Sua aplicação      │
├───────────────────────────┤
│         DiscordKit        │
│ Router · Context          │
│ Commands · Components     │
│ Forms · Middleware        │
│ Ciclo de respostas        │
├───────────────────────────┤
│          discordgo        │
│ Gateway · REST            │
│ Sessions · tipos Discord  │
├───────────────────────────┤
│        Discord API        │
└───────────────────────────┘
```

discordgo fornece a camada de protocolo e modelo de dados.

DiscordKit fornece estrutura de aplicação em torno das interações.

## O que continua sendo discordgo

DiscordKit não cria modelos alternativos para as entidades do Discord.

`RequireUserOption` retorna `*discordgo.User`, canais continuam sendo `*discordgo.Channel`, membros continuam sendo `*discordgo.Member` e sua aplicação continua controlando a `*discordgo.Session` original.

Isso é proposital.

Criar wrappers para todos esses tipos resultaria em outro modelo de objetos que precisaria ser constantemente convertido de volta para discordgo.

## O que o DiscordKit adiciona

DiscordKit foca nos padrões acima do discordgo:

- roteamento;
- helpers de Context;
- builders tipados;
- ciclo de vida das interações;
- middleware;
- Components V2;
- formulários e parsing de modais.

## Você sempre pode acessar discordgo diretamente

DiscordKit não é uma abstração fechada.

Dentro de um handler, a sessão e a interação originais permanecem acessíveis.

Esse escape hatch é útil quando o Discord introduz um novo recurso, quando discordgo implementa algo antes do DiscordKit ou quando você precisa de controle em baixo nível.

## Aplicações discordgo existentes

Você não precisa reescrever uma aplicação existente.

```go
session.AddHandler(existingMessageHandler)
session.AddHandler(existingReadyHandler)
session.AddHandler(router.Handle)
```

A adoção incremental é suportada.

## Regra prática

Use DiscordKit quando o problema estiver principalmente na **estrutura de interações da aplicação**.

Use discordgo diretamente quando o problema estiver principalmente no **Discord em si**.

Na prática, boas aplicações DiscordKit usam ambos.
