---
title: "Parâmetros de rota"
description: "Conecte ações de componentes aos registros da aplicação."
---

```go
id, err := discordkit.Route("/jobs/:id/save").Param("id", "42").Build()
if err != nil { return err }
button := discordkit.Button("Save", id)
_ = button // Add to a Row in the outgoing MessageSpec.

if err := router.Component("/jobs/:id/save", func(c *discordkit.Context) error {
    jobID, err := c.RequireParam("id")
    if err != nil { return err }
    // Check c.User() is allowed to access jobID before changing data.
    return c.EphemeralText("Selected job: " + jobID)
}); err != nil { return err }
```
Use o builder de rotas em vez de concatenar entradas do usuário em custom IDs. Respeite o limite de custom ID do Discord e nunca coloque segredos nele. Um parâmetro identifica um registro; ele não autoriza o acesso. Busque o registro e confira propriedade ou permissões antes de alterá-lo.
