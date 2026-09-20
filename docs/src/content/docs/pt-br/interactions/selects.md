---
title: "Menus de seleção"
description: "Ofereça opções e valide a seleção recebida."
---

```go
func choose(c *discordkit.Context) error {
    return c.Ephemeral(discordkit.MessageSpec{
        Components: discordkit.Components(
            discordkit.Text("Choose a topic"),
            discordkit.Row(discordkit.StringSelect("/help/topic").Options(
                discordkit.SelectOption("Rules", "rules"),
                discordkit.SelectOption("Support", "support"),
            )),
        ),
    })
}

func selected(c *discordkit.Context) error {
    data := c.Interaction.MessageComponentData()
    if len(data.Values) != 1 {
        return c.EphemeralText("Choose one topic.")
    }
    switch data.Values[0] {
    case "rules":
        return c.EphemeralText("Read the server rules before posting.")
    case "support":
        return c.EphemeralText("Contact a moderator in the support channel.")
    default:
        return c.EphemeralText("Unknown topic.")
    }
}
```
Registre `selected` com `router.Component("/help/topic", selected)`. Use `choose` como handler de um comando declarado e sincronizado. Uma row com select deve conter apenas esse select. Valide valores recebidos mesmo quando o menu oferece um conjunto fixo.
