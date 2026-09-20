---
title: "Intents do Gateway"
description: "Declare quais eventos seu bot recebe."
---

Para os exemplos de comandos e componentes desta documentação, configure a sessão antes de abri-la:

```go
session.Identify.Intents = discordgo.IntentsGuilds
```

Slash commands e botões não precisam ler o conteúdo de mensagens comuns. Não solicite intents privilegiadas para essas receitas.

Uma funcionalidade baseada em eventos pode precisar de intents adicionais. Dar boas-vindas a novos membros exige a intent de membros, habilitada tanto no código quanto no Developer Portal. Alterar somente um dos lados pode causar ausência de eventos ou fechamento da conexão.

Intents controlam o recebimento de eventos. Permissões do bot controlam ações como enviar mensagens ou moderar um servidor. Verifique as duas coisas ao adicionar uma funcionalidade.
