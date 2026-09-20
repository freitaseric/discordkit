---
title: "Botões"
description: "Crie uma mensagem Components V2 e atualize-a ao clicar."
---

```go
func panel(c *discordkit.Context) error {
    return c.Ephemeral(discordkit.MessageSpec{
        Components: discordkit.Components(
            discordkit.Text("Confirm this action?"),
            discordkit.Row(discordkit.Button("Confirm", "/confirm").Success()),
        ),
    })
}

func confirm(c *discordkit.Context) error {
    return c.Update(discordkit.MessageSpec{
        Components: discordkit.Components(discordkit.Text("Confirmed.")),
    })
}
```
Registre `panel` como handler de um slash command e `confirm` com `router.Component("/confirm", confirm)`, tratando os dois erros de registro. Inclua o slash command nas declarações sincronizadas.

`Text` ativa Components V2 por meio de `MessageSpec`. `Update` substitui a mensagem original e reconhece o clique. Não chame `Reply` depois de `Update` na mesma interação. Use `LinkButton` para URLs; botões de link não enviam interações ao router.
