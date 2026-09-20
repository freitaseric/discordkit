---
title: Visão geral da API
description: Mapa das principais APIs do DiscordKit.
---

Esta página apresenta as APIs principais do pacote. Consulte também os guias de [roteamento](/pt-br/concepts/router/), [interações](/pt-br/concepts/interaction-model/) e [instalação](/pt-br/getting-started/installation/).

## Comandos

Crie comandos de aplicação com `Command`, `UserCommand` e `MessageCommand`. Comandos de barra aceitam builders tipados, como `StringOption`, `IntegerOption`, `BooleanOption`, `UserOption` e `ChannelOption`. Use `SubCommand` e `SubCommandGroup` para criar subcomandos e grupos.

`Build` retorna um `*discordgo.ApplicationCommand` e um erro; `MustBuild` causa panic em declarações inválidas; `BuildCommands` monta várias declarações. As validações cobrem nomes, descrições, quantidade e duplicidade de opções, hierarquia e a ordem de opções obrigatórias antes das opcionais. O registro no Discord continua sob controle da aplicação.

```go
ping, err := discordkit.Command("ping", "Verifica se o bot está respondendo").Build()
if err != nil {
    return err
}
```

## Roteador e middleware

Um `Router` despacha comandos, componentes, envios de modais e autocomplete. Registre handlers com `Command`, `Component`, `Modal` e `Autocomplete`, e conecte `router.Handle` à sessão discordgo com `Session.AddHandler`. Handlers usam a assinatura `func(*discordkit.Context) error`.

`Group` compartilha prefixo e middleware. `Use` adiciona middleware global e cada rota pode receber middleware próprio. `Recovery` recupera panics e os transforma em erros. `OnError` permite observar erros de despacho; sem um hook, os erros são registrados em log. A [página de roteamento](/pt-br/concepts/router/) detalha parâmetros nomeados e precedência de rotas.

## Contexto e ciclo de resposta

O `Context` oferece a sessão e interação do discordgo, acesso tipado às opções, parâmetros de rota e métodos de resposta. Use `ReplyText` ou `Reply` para a resposta inicial. Use `Defer` para confirmar uma interação cujo processamento levará mais tempo e depois conclua com `EditReply` ou `Followup`. Handlers de componentes podem usar `Update` ou `DeferUpdate`; `Ephemeral` controla a visibilidade quando permitido.

O contexto acompanha as confirmações e rejeita operações incompatíveis com o estado atual. `Defer` não move a execução para segundo plano. A aplicação é responsável por prazos e trabalhos externos.

## Mensagens e Components V2

`MessageSpec` descreve conteúdo, embeds, flags, arquivos e componentes. A DSL `Components` inclui linhas, botões interativos/de link/premium, selects de string/entidade/canal, seções, containers, displays de texto, galerias, separadores, miniaturas e componentes de arquivo. `Raw` permite usar diretamente um componente discordgo sem builder dedicado.

Antes do envio, a validação verifica posicionamento, action rows, custom IDs, tamanho de galerias e limite recursivo. Quando há componentes V2, a flag necessária é aplicada automaticamente. Mensagens V2 seguem as restrições do Discord e não devem combinar conteúdo ou embeds legados com esse layout.

## Modais e formulários

O pacote oferece inputs de texto, selects de string e uploads de arquivo em modais, conforme a API do Discord e a versão usada do discordgo. Os valores enviados podem ser lidos pelos helpers de modal do contexto. A árvore precisa ter pelo menos um componente e cada label precisa conter um input compatível. A aplicação decide como armazenar ou processar anexos enviados.

## Custom IDs

`CustomID` monta e interpreta IDs com segmentos separados por barras e parâmetros nomeados. Use-os nas rotas `Router.Component` e `Router.Modal`. IDs são visíveis para clientes, portanto não coloque segredos ou dados sensíveis neles e respeite o limite de tamanho do Discord.

## Sincronização de comandos

`DiffCommands` compara comandos desejados com os remotos sem fazer chamadas de rede e retorna um `SyncPlan`. `SyncCommands` consulta o Discord e aplica criações e atualizações. Comandos obsoletos só são excluídos quando `SyncOptions.Delete` está habilitado. `GuildID` vazio indica comandos globais; preenchido, comandos de servidor. Confira a política de exclusão antes de usá-la em produção.

## Erros e discordgo

Erros de builders e validadores podem ser verificados com `errors.Is` e sentinelas como `ErrInvalidCommand`, `ErrInvalidComponent`, `ErrInvalidModal`, `ErrAlreadyAcknowledged` e `ErrRouteNotFound`. Acesso à sessão e aos tipos originais do discordgo permanece disponível para recursos ainda não cobertos pela DSL.
