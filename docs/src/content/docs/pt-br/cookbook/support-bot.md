---
title: "O projeto: uma central de atendimento"
description: "O que vamos construir, como estudar e como executar cada etapa."
---

Vamos construir um bot que resolve um fluxo completo: um membro abre um chamado, acompanha o atendimento, a equipe assume o trabalho e o solicitante avalia a solução. Os registros continuam disponíveis depois de reiniciar o processo.

O resultado é um **registro de chamados dentro do Discord**, com interfaces privadas. Ele não cria canais privados, não encaminha uma conversa em tempo real e não promete notificar automaticamente o solicitante. A equipe consulta a fila, vê quem abriu o chamado e combina o atendimento no servidor. Essas distinções importam para você saber o que está entregando.

## Como seguir o guia

Você precisa saber criar arquivos, executar comandos no terminal e reconhecer funções, structs e erros em Go. Explicaremos as decisões de organização e as APIs usadas. Se nunca conectou um bot, comece pelo capítulo de preparação.

Há dois modos de estudar:

- **Acompanhado:** clone o projeto, comece na etapa 1 e ative uma funcionalidade por vez. Leia e altere os arquivos indicados em cada capítulo. O esqueleto completo já existe para que cada parada compile.
- **Reconstrução:** depois da primeira leitura, crie seu próprio módulo e reescreva os arquivos por responsabilidade. As páginas mostram o código completo de cada arquivo; o capítulo de operação explica como trocar o caminho do módulo.

`COOKBOOK_STAGE` é uma ferramenta didática. Não é feature flag da DiscordKit nem sistema de migração. Ele decide quais comandos e rotas são registrados. O padrão é `6`, o bot completo. Aumente a etapa ao terminar o exercício; pare com Ctrl+C antes de iniciar novamente. Voltar a uma etapa anterior não apaga comandos antigos do Discord, porque a sincronização é propositalmente não destrutiva.

| Capítulo | Etapa executável | O que você entrega |
| --- | --- | --- |
| [01 · Preparação](/pt-br/cookbook/setup/) | 1 | Aplicação instalada e `/ping` funcionando |
| [02 · Arquitetura](/pt-br/cookbook/architecture/) | 1 | Configuração, ciclo de vida e módulos compreendidos |
| [03 · Comandos](/pt-br/cookbook/commands/) | 2 | Opções tipadas, respostas públicas e privadas |
| [04 · Painel](/pt-br/cookbook/components/) | 3 | Central pública, FAQ privado e permissões |
| [05 · JSON-db](/pt-br/cookbook/persistence/) | 3 + testes | Repositório persistente com regras de acesso |
| [06 · Chamados](/pt-br/cookbook/tickets/) | 4 | Modal, custom IDs, atribuição e encerramento |
| [07 · Fila](/pt-br/cookbook/queue/) | 5 | Filtro, paginação e autocomplete autorizado |
| [08 · Operação](/pt-br/cookbook/operations/) | 6 | Avaliação, exportação, testes e hospedagem |
| [09 · Laboratório](/pt-br/cookbook/laboratory/) | 6 | Opções resolvidas, menus de contexto e mídia |
| [Mapa da API](/pt-br/cookbook/api-map/) | Consulta | Onde cada família da API aparece e seus limites |

## Três camadas, três responsabilidades

**Go** fornece módulos, pacotes, structs, interfaces, goroutines, mutex, JSON e arquivos. **discordgo** conecta o Gateway e expõe os tipos e endpoints do Discord. **DiscordKit** organiza o fluxo de interações: builders, roteamento, contexto, middleware e respostas.

O JSON-db é código deste aplicativo. Não estamos adicionando uma abstração de banco à biblioteca. Também não há uma camada mágica de injeção de dependências: `main` constrói o store e o entrega ao bot.

## Critérios de conclusão

Ao terminar, você deve conseguir abrir e consultar um chamado, impedir outro membro de acessá-lo, assumir como atendente, encerrar como autor ou equipe, avaliar como autor, reiniciar sem perder os dados e exportar apenas os registros autorizados. Além disso, deve conseguir explicar por que `ShowModal` não vem depois de `Defer` e por que um custom ID não substitui uma autorização.

[Código executável completo](https://github.com/freitaseric/discordkit/tree/main/examples/community-bot). Os blocos deste guia são sincronizados com os arquivos do exemplo e conferidos no CI.

---

[Próximo →: 01 · Prepare e conecte o bot](/pt-br/cookbook/setup/)
