---
title: "Estrutura do projeto"
description: "Organize o bot quando ele crescer além do main.go."
---

Comece com um `main.go` enquanto aprende. Quando os handlers crescerem, use pacotes Go comuns:

| Caminho | Responsabilidade |
| --- | --- |
| `cmd/bot/main.go` | Ler configuração, conectar a sessão e aguardar encerramento |
| `internal/discord/commands.go` | Declarar comandos e registrar rotas |
| `internal/discord/components.go` | Tratar botões, selects e modais |
| `internal/service/` | Regras de negócio independentes do Discord |
| `internal/store/` | Persistência e consultas |

DiscordKit não varre diretórios nem registra handlers automaticamente. Exporte uma função de registro e chame-a explicitamente antes de `session.Open()`.

```go title="internal/discord/commands.go"
package discord

import "github.com/freitaseric/discordkit"

func Register(r *discordkit.Router) error {
    return r.Command("ping", func(c *discordkit.Context) error {
        return c.ReplyText("Pong!")
    })
}
```

Importe o pacote com o caminho do seu módulo, como `example.com/meu-bot/internal/discord`. Leia a configuração na inicialização e injete os serviços nos handlers. Evite variáveis globais mutáveis: discordgo pode executar handlers simultaneamente.
