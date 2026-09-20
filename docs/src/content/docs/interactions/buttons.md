---
title: "Buttons"
description: "Build a Components V2 message and update it on click."
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
Register `panel` as a slash-command handler and `confirm` with `router.Component("/confirm", confirm)`, checking both registration errors. Include the slash command in the synchronized declarations.

`Text` activates Components V2 through `MessageSpec`. `Update` replaces the original message and acknowledges the click. Do not call `Reply` after `Update` for the same interaction. Use `LinkButton` for URLs; link buttons do not dispatch component interactions.
