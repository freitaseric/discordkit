---
title: "Autocomplete"
description: "Suggest values as the user types."
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
Import `strings`. Register the final `docs` command handler separately; the autocomplete handler only returns suggestions. Keep results at 25 or fewer, respond promptly, and validate the final value in the command handler because users can enter a value outside the suggestions. Fixed `Choices` and `Autocomplete` cannot be used on the same option.
