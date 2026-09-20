---
title: "Gateway intents"
description: "Declare the events your bot receives."
---

For the command and component examples in this documentation, configure the session before opening it:

```go
session.Identify.Intents = discordgo.IntentsGuilds
```

Slash commands and buttons do not require reading ordinary message content. Do not request privileged intents for these recipes.

An event-driven feature may require additional intents. A welcome message for new members needs the member intent, enabled both in code and in the Developer Portal. Changing only one side can result in missing events or the Gateway closing the connection.

Intents control event delivery. Bot permissions control actions such as sending messages or moderating a server. Check both when adding a new feature.
