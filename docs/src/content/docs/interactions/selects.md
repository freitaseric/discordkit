---
title: "Select menus"
description: "Offer choices and validate the submitted selection."
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
Register `selected` with `router.Component("/help/topic", selected)`. Use `choose` as a declared and synchronized command handler. A row containing a select must contain only that select. Values from users still need validation even when the menu offers a fixed set.
