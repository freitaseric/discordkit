---
title: "07 · Filters, pages and autocomplete"
description: "Navigate real records without leaking ticket titles."
---

**Enable stage 5.** `/ticket list` shows a private queue with five records per page.

Filter and page travel through custom IDs. Validate both and clamp the page against the current result set; records may change between clicks. `Store.List` authorizes before rendering. Members see their records; staff see their guild's records.

`Update` replaces the component's original message immediately. The lookup is in memory. A slower remote repository would require `DeferUpdate` followed by `Edit`.

<!-- cookbook-source: internal/bot/queue.go -->
```go title="internal/bot/queue.go"
package bot

import (
	"strconv"

	dk "github.com/freitaseric/discordkit"
)

func validStatus(s string) bool { return s == "open" || s == "closed" || s == "all" }
func (b *Bot) queue(c *dk.Context, status string, page int) dk.MessageSpec {
	filter := status
	if filter == "all" {
		filter = ""
	}
	return queueView(b.Store.List(actor(c), filter), status, page)
}
func (b *Bot) registerQueue(r *dk.Router) func() error {
	return func() error {
		if err := r.Command("ticket list", func(c *dk.Context) error {
			status, ok := c.String("status")
			if !ok {
				status = "open"
			}
			if !validStatus(status) {
				return c.EphemeralText("Filtro inválido.")
			}
			return c.Ephemeral(b.queue(c, status, 0))
		}); err != nil {
			return err
		}
		if err := r.Component("/queue/filter", func(c *dk.Context) error {
			values := c.Interaction.MessageComponentData().Values
			if len(values) != 1 || !validStatus(values[0]) {
				return c.EphemeralText("Filtro inválido.")
			}
			return c.Update(b.queue(c, values[0], 0))
		}); err != nil {
			return err
		}
		return r.Component("/queue/:status/:page", func(c *dk.Context) error {
			status, err := c.RequireParam("status")
			if err != nil {
				return err
			}
			raw, err := c.RequireParam("page")
			if err != nil {
				return err
			}
			page, err := strconv.Atoi(raw)
			if err != nil || !validStatus(status) {
				return c.EphemeralText("Página inválida.")
			}
			return c.Update(b.queue(c, status, page))
		})
	}
}
```
<!-- /cookbook-source -->

## Autocomplete can leak data too

Read `completeTicket` in the [previous chapter](/cookbook/tickets/). It reads `FocusedOption`, searches only authorized records, caps results at 25 and truncates choice labels. It sends an autocomplete response, not a message or deferred response. No matches means `c.Autocomplete()`.

Suggestions do not authorize a later request: users can type arbitrary IDs. The show handler must call the authorized store lookup again. The example makes no REST calls per suggestion; a remote search would need an appropriate latency budget and index.

**Verify:** create six records; navigate from five results to one; check disabled first/last controls; close everything and inspect the empty open filter; switch to all. Repeat as a user without tickets and confirm no data leaks. Search a partial subject in `/ticket show id:`.

**Exercise:** add an assigned-to-me filter using the authenticated actor ID, preserving the guild boundary.

Next: [ratings, export and operations](/cookbook/operations/).

---

[← Previous: 06 · Open and track tickets](/cookbook/tickets/) · [Next →: 08 · Test and operate the bot](/cookbook/operations/)
