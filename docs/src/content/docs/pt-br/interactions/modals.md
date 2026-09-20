---
title: "Modais"
description: "Abra um formulário e trate sua submissão."
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
Registre `openForm` em um comando ou botão e `receiveForm` com `router.Modal("/feedback", receiveForm)`. Trate erros de registro. Abrir um modal é a resposta inicial: não faça defer antes de `ShowModal`. A submissão é uma nova interação, com seu próprio ciclo de resposta.

O exemplo não persiste feedback. Adicione armazenamento e use `Defer(true)` na **submissão** antes de I/O demorado; finalize com `Edit` em vez de uma nova resposta inicial.
