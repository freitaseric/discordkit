---
title: "Route parameters"
description: "Connect component actions to application records."
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
Use the route builder instead of concatenating user input into custom IDs. Keep IDs within the Discord custom ID limit, and never put secrets in them. A route parameter identifies a record; it does not authorize access. Fetch the record and check ownership or permissions before making changes.
