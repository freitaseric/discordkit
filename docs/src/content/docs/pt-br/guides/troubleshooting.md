---
title: "Solução de problemas"
description: "Identifique a etapa com falha antes de mudar a configuração."
---

| Sintoma | O que verificar |
| --- | --- |
| 401 / Unauthorized | Redefina tokens inválidos; use o token do bot, não a chave pública |
| 403 / Missing Access na sincronização | Aplicação instalada na guild escolhida; ID do servidor correto |
| Comando não aparece | Escopo, sincronização concluída e permissões do comando |
| Gateway 4014 | Intents privilegiadas solicitadas habilitadas no portal; remova as desnecessárias |
| O aplicativo não respondeu | Logs do handler, rota correspondente e reconhecimento rápido |
| Já respondeu | Apenas uma resposta inicial; depois use Edit ou Followup |
| Componente rejeitado | Validação de MessageSpec; V2 não aceita embeds legados nem polls |

Registre erros de rotas e sincronização na inicialização. Para trabalho demorado, reconheça com `Defer` e finalize com `Edit`. Não registre o objeto inteiro da interação: ele pode conter tokens e dados do usuário.

Se duas cópias do bot estiverem rodando, encerre a instância inesperada antes de depurar respostas. Reproduza em um servidor de teste com um processo e um comando.
