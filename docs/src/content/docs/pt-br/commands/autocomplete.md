---
title: "Autocomplete"
description: "Sugira valores enquanto o usuário digita."
---

```go
command := discordkit.Command("docs", "Find documentation",
    discordkit.StringOption("topic", "Topic").Required().Autocomplete(),
)
_ = command // Include in BuildCommands and SyncCommands.
if err := router.Autocomplete("docs", "topic", func(c *discordkit.Context) error {
    query, _ := c.String("topic")
    topics := []string{"commands", "components", "modals"}
    choices := make([]discordkit.ChoiceValue, 0, len(topics))
    for _, topic := range topics {
        if strings.Contains(topic, strings.ToLower(query)) {
            choices = append(choices, discordkit.Choice(topic, topic))
        }
    }
    return c.Autocomplete(choices...)
}); err != nil { return err }
```
Importe `strings`. Registre separadamente o handler final do comando `docs`; o handler de autocomplete apenas devolve sugestões. Retorne no máximo 25 resultados, responda rapidamente e valide o valor final no comando, pois o usuário pode digitar algo fora das sugestões. `Choices` fixas e `Autocomplete` não podem ser usados na mesma opção.
