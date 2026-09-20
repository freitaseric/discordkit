---
title: "Inicialização"
description: "Conecte configuração, rotas e sincronização de comandos."
---

Use esta ordem na inicialização:

1. Valide as variáveis de ambiente.
2. Construa as declarações dos comandos e trate erros de validação.
3. Crie a sessão e configure intents explicitamente.
4. Registre todas as rotas, verificando cada erro retornado.
5. Anexe `router.Handle` e abra a sessão.
6. Sincronize comandos com o ID da aplicação e o servidor escolhido.
7. Aguarde o encerramento e feche a conexão.

Use `func run() error` para controlar a sessão e `defer session.Close()` depois da conexão. Retorne erros de inicialização de `run` e registre o erro final em `main`, permitindo executar a limpeza antes de encerrar o processo.

`SyncOptions{GuildID: guildID}` registra no servidor de teste. `GuildID` vazio significa comandos globais. `Delete: true` remove comandos remotos ausentes da declaração; habilite-o apenas se este processo controlar todos os comandos daquele escopo.

Siga o [cookbook do bot de atendimento](/pt-br/cookbook/support-bot/) para um programa completo com esse ciclo.
