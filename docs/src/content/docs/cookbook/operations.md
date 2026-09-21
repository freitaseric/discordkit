---
title: "08 · Test and operate the bot"
description: "Ratings, export, failure handling, backup and hosting."
---

**Enable stage 6.** Ratings and export complete the workflow. The optional laboratory becomes available too.

`RequireInt` reads the rating; the store checks that the ticket is closed and the actor is its owner. Staff cannot rate on behalf of another user.

Export serializes the authorized `Store.List` result, never the raw database. A conservative 7 MiB guard avoids oversized requests; Discord determines the actual upload limit. Complete the private defer using `Edit`, then send a private file using `Followup`. The attachment URL must match the upload filename. `FollowupEdit` adds final details while retaining attachments. The dismiss button acknowledges with `DeferUpdate` and calls `DeleteResponse` for the component's message.

Deleting an export message neither removes tickets nor revokes downloaded copies.

<!-- cookbook-source: internal/bot/operations.go -->
```go title="internal/bot/operations.go"
package bot

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/bwmarrin/discordgo"
	dk "github.com/freitaseric/discordkit"
)

func (b *Bot) registerOperations(r *dk.Router) func() error {
	return func() error {
		if err := r.Autocomplete("ticket rate", "id", b.completeTicket); err != nil {
			return err
		}
		if err := r.Command("ticket rate", func(c *dk.Context) error {
			id, err := c.RequireString("id")
			if err != nil {
				return err
			}
			score, err := c.RequireInt("score")
			if err != nil {
				return err
			}
			if err = c.Defer(true); err != nil {
				return err
			}
			t, err := b.Store.Change(actor(c), id, "rate", int(score))
			if err != nil {
				return finish(c, dk.MessageSpec{}, err)
			}
			return finish(c, ticketView(t), nil)
		}); err != nil {
			return err
		}
		if err := r.Component("/export/dismiss", func(c *dk.Context) error {
			if err := c.DeferUpdate(); err != nil {
				return err
			}
			return c.DeleteResponse()
		}); err != nil {
			return err
		}
		return r.Command("ticket export", b.export)
	}
}
func (b *Bot) export(c *dk.Context) error {
	if err := c.Defer(true); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(b.Store.List(actor(c), ""), "", "  ")
	if err != nil {
		return finish(c, dk.MessageSpec{}, err)
	}
	if len(raw) > 7*1024*1024 {
		return finish(c, message(dk.Text("Exportação muito grande. Solicite um backup ao operador.")), nil)
	}
	// Complete the deferred response before creating an additional private message.
	if err = finish(c, message(dk.Text("Exportação preparada.")), nil); err != nil {
		return err
	}
	spec := message(dk.Text("## Exportação de chamados"), dk.File("attachment://tickets.json"), dk.Row(dk.Button("Ocultar exportação", "/export/dismiss").Secondary()))
	spec.Ephemeral = true
	spec.Files = []*discordgo.File{{Name: "tickets.json", ContentType: "application/json", Reader: bytes.NewReader(raw)}}
	sent, err := c.Followup(spec)
	if err != nil {
		return err
	}
	// Edit the same followup while retaining its attachment and controls.
	spec.Components = dk.Components(dk.Text(fmt.Sprintf("## Exportação pronta\n%d bytes · somente chamados autorizados.", len(raw))), dk.File("attachment://tickets.json"), dk.Row(dk.Button("Ocultar exportação", "/export/dismiss").Secondary()))
	spec.Ephemeral = false // Editing does not change the existing message visibility.
	spec.Files = nil
	spec.Attachments = sent.Attachments
	_, err = c.FollowupEdit(sent.ID, spec)
	return err
}
```
<!-- /cookbook-source -->

## Offline tests and live acceptance

```bash
go test ./...
go test -race ./...
go vet ./...
```

Tests cover store policies, persistence, all stage registrations, message validation and a modal submission through a fake HTTP transport. They require no real token. They cannot verify remote API behavior, installation or channel permissions.

Test these scenarios in Discord: open via panel and slash command; reject another member's manual ID lookup; claim as staff and reject a competing claim; close and rate as owner; reject rating by another actor; restart and read the same state; export only authorized data; dismiss the file while retaining its records.

## Extract your own module

Copy `cmd`, `internal`, `.gitignore` and `.env.example` into a new directory:

```bash
mkdir my-bot
cp -R cmd internal .gitignore .env.example my-bot/
cd my-bot
go mod init example.com/my-bot
```

Replace local import prefixes `github.com/freitaseric/discordkit/examples/community-bot/internal/` with `example.com/my-bot/internal/`. Keep the DiscordKit library imports unchanged. In the original clone, obtain the tested revision with `git rev-parse HEAD`. In the new module, run `go get github.com/freitaseric/discordkit@REVISION`, `go mod tidy`, `go test ./...` and `go run ./cmd/bot`. Never copy the database or real credentials into Git.

## Hosting and recovery

Vercel hosts this documentation. The Gateway bot needs a continuously running Go process on a VM, worker or container with persistent disk. An ephemeral HTTP function is unsuitable for this connection and JSON store.

Build with `go build -o bot ./cmd/bot`. Use an absolute `BOT_DATA_FILE`, persistent volume and one replica. Supply credentials through the service environment. See the [hosting guide](/guides/deployment/) for basic process supervision.

For backup, stop the service, copy the JSON to a protected timestamped file, and restart. To restore, stop, preserve the current file, restore the backup with correct ownership and permissions, restart, and query a known ID. Corrupt files must fail startup, never silently become `{}`.

This example has no automatic retention, anti-spam, job queue or complete handler draining. Add quotas, retention, monitoring and an appropriate database before serving a large community.

**Final exercise:** restore a backup into your test deployment and repeat the acceptance scenarios. Explain event → router → actor → store → MessageSpec → Discord.

Next: [complementary API laboratory](/cookbook/laboratory/).

---

[← Previous: 07 · Filters, pages and autocomplete](/cookbook/queue/) · [Next →: 09 · Explore the remaining API](/cookbook/laboratory/)
