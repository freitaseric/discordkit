---
title: "Modals"
description: "Open a form and handle its submission."
---

```go
func openForm(c *discordkit.Context) error {
    form, err := discordkit.Form("/feedback", "Feedback",
        discordkit.Field("Your suggestion",
            discordkit.TextInput("message").Paragraph().Required(true).MaxLen(500)),
    ).Build()
    if err != nil { return err }
    return c.ShowModal(form)
}

func receiveForm(c *discordkit.Context) error {
    message, err := c.Form().RequireString("message")
    if err != nil { return err }
    // Validate and persist message in your application before acknowledging success.
    _ = message
    return c.EphemeralText("Form received for this demonstration; nothing was saved.")
}
```
Register `openForm` on a command or button, and `receiveForm` using `router.Modal("/feedback", receiveForm)`. Check registration errors. Opening a modal is the initial response: do not defer that interaction before `ShowModal`. The submission is a new interaction with its own response lifecycle.

The example intentionally does not persist feedback. Add storage and use `Defer(true)` on the **submission** before slow I/O; finish it with `Edit` instead of another initial reply.
