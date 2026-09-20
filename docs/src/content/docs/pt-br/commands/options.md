---
title: "Opções e subcomandos"
description: "Use entradas tipadas em vez de interpretar texto de mensagens."
---

```go
command := discordkit.Command("greet", "Greet someone",
    discordkit.StringOption("name", "Name").Required().MaxLen(80),
)
commands, err := discordkit.BuildCommands(command)
if err != nil { return err }
_ = commands // Pass to SyncCommands during startup.

if err := router.Command("greet", func(c *discordkit.Context) error {
    name, err := c.RequireString("name")
    if err != nil { return err }
    return c.EphemeralText("Hello, " + name + "!")
}); err != nil { return err }
```
Estes trechos ficam dentro da função de inicialização, onde `router` já existe. Declare opções obrigatórias antes das opcionais. `BuildCommands` verifica a declaração; `RequireString` valida a entrada da interação.

Para `Command("config", ..., SubCommand("show", ...))`, registre a rota como `router.Command("config show", handler)`. Caminhos de comando usam espaços, sem a barra inicial. Declarar um subcomando não registra seu handler.

Menus de contexto de usuário e mensagem usam os builders `UserCommand` e `MessageCommand`. Quando necessário, leia o alvo resolvido na interação original `c.Interaction` com os tipos do discordgo.
