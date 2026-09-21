# DiscordKit

DiscordKit is an ergonomic framework layer built **on top of**
[`discordgo`](https://github.com/bwmarrin/discordgo). It keeps discordgo as the
transport and data-model layer and adds the pieces application developers write
by hand over and over: a unified interaction router, typed command and option
builders, a Components V2 DSL, modal forms, and a response lifecycle that turns
easy-to-miss API rules into compile-time and run-time guarantees.

DiscordKit never wraps discordgo's domain types. A user is a
`*discordgo.User`, a guild is a `*discordgo.Guild`, and every builder finalizes
into plain discordgo structs, so you can always drop down to the underlying
library.

## Learn by building

Follow the [cookbook in English](https://discordkit.freitaseric.com/cookbook/)
or [português](https://discordkit.freitaseric.com/pt-br/cookbook/): build a modular
support bot with ticket forms, authorization, a local JSON store, filters,
autocomplete, ratings, exports, tests and deployment guidance.
The [executable project](examples/community-bot) has six runnable learning stages.

## Installation

```bash
go get github.com/freitaseric/discordkit
```

DiscordKit requires **Go 1.26+** and a recent discordgo (Components V2, modal
labels and file uploads rely on discordgo master; the exact minimum commit is
pinned in `go.mod`).

## Quick start

```go
package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/freitaseric/discordkit"
)

func main() {
	s, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	r := discordkit.NewRouter(discordkit.Recovery(), discordkit.Logging(nil))
	r.Command("ping", func(c *discordkit.Context) error {
		return c.ReplyText("Pong!")
	})
	s.AddHandler(r.Handle)

	if err := s.Open(); err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	// Register the slash command with Discord.
	cmds, _ := discordkit.BuildCommands(discordkit.Command("ping", "Check I am alive"))
	if _, err := discordkit.SyncCommands(s, s.State.User.ID, cmds, discordkit.SyncOptions{}); err != nil {
		log.Fatal(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}
```

## Commands and synchronization

Build commands with typed option builders. Validation (name/length limits, the
25-option cap, unique names, required-before-optional ordering, and
`choices`/`autocomplete` mutual exclusion) happens at `Build`:

```go
cmd := discordkit.Command("greet", "Greet a member",
	discordkit.UserOption("who", "Who to greet").Required(),
	discordkit.StringOption("tone", "How to greet").
		Choices(discordkit.Choice("Formal", "formal"), discordkit.Choice("Casual", "casual")),
)

sub := discordkit.Command("config", "Server config",
	discordkit.SubCommandGroup("channel", "Channel settings",
		discordkit.SubCommand("set", "Set the channel",
			discordkit.ChannelOption("target", "Channel").Required(),
		),
	),
)
```

`SyncCommands` diffs your declared commands against what is registered with
Discord and only issues the create/update/delete calls that are actually needed.
The diff is exposed as a pure, testable function:

```go
plan := discordkit.DiffCommands(remote, desired, true) // deleteObsolete = true
fmt.Println(len(plan.Create), len(plan.Update), len(plan.Delete))
```

Set `SyncOptions.GuildID` to scope to a single guild and `SyncOptions.Delete` to
control whether obsolete commands are removed.

## Context

Every handler receives a `*Context` for exactly one interaction. It exposes
typed, network-free option accessors that read from the interaction's resolved
data:

```go
name, _ := c.String("tone")           // (string, bool)
who, _ := c.UserOption("who")         // *discordgo.User
count := c.MustInt("count")           // panics if missing/invalid
target, err := c.RequireChannel("target")
```

`Session` and `Interaction` are exported escape hatches; the raw discordgo API is
always one field away.

## Routing, groups and middleware

`Router` dispatches command, component, modal and autocomplete interactions
through a single tree. Component and modal routes match on custom-ID patterns
with named parameters:

```go
r := discordkit.NewRouter(discordkit.Recovery())

r.Command("greet", greetHandler)
r.Component("/jobs/:jobID/cancel", cancelHandler)  // matches custom_id built with Route
r.Autocomplete("greet", "tone", toneSuggestions)

admin := r.Group("admin", discordkit.RequirePermissions(discordgo.PermissionManageServer))
admin.Command("ban", banHandler)
```

Built-in middleware: `Recovery()` (turns panics into a `*PanicError`),
`Logging(*slog.Logger)`, `RequireGuild()`, and `RequirePermissions(bitmask)`.
Middleware is `func(Handler) Handler`, so custom middleware is trivial.

## Custom IDs

Build and parse component custom IDs safely instead of using `fmt.Sprintf` and
`strings.Split`:

```go
id, err := discordkit.Route("/jobs/:jobID/cancel").Param("jobID", job.ID).Build()
btn := discordkit.Button("Cancel", string(id))
// In the handler registered at "/jobs/:jobID/cancel":
jobID, _ := c.Param("jobID")
```

Values are escaped exactly once and the final ID is validated against Discord's
100-character limit.

## Components V2

The component DSL finalizes into discordgo components and is validated as a tree.
When any Components V2 component is present, DiscordKit sets the
`IS_COMPONENTS_V2` message flag, moves `Content` into a leading `TextDisplay`,
and rejects the combinations Discord forbids (legacy embeds and polls alongside
V2):

```go
msg := discordkit.MessageSpec{
	Components: discordkit.Components(
		discordkit.Container(
			discordkit.Text("## Deploy finished"),
			discordkit.Separator().Divider(true),
			discordkit.Section(discordkit.Text("Logs are ready.")).
				Accessory(discordkit.Button("Open", "logs:open")),
		),
	),
}
c.Reply(msg)
```

Available builders include `Text`, `Row`, `Button`/`LinkButton`/`PremiumButton`,
`StringSelect`, `UserSelect`/`RoleSelect`/`MentionableSelect`, `ChannelSelect`,
`Section`, `Thumbnail`, `Gallery`/`MediaItem`, `File`, `Separator` and
`Container`. `Raw(...)` drops any discordgo component straight into the tree as an
escape hatch.

Validation rules enforced by `ValidateMessageComponents` include: interactive
components must live in an action row, a select is alone in its row, buttons need
a custom ID (link buttons need a URL, premium buttons neither), sections have
1-3 text displays plus an accessory, thumbnails appear only as section
accessories, containers do not nest, and the overall component budget is capped.

## Forms and modals

Modals are built from `Label`-wrapped fields and validated on build:

```go
modal, _ := discordkit.Form("feedback:new", "Send feedback",
	discordkit.Field("Subject", discordkit.TextInput("subject")),
	discordkit.Field("Details", discordkit.TextInput("body").Paragraph()),
	discordkit.Field("Attachment", discordkit.FileUpload("file").FileTypes("png", "pdf")),
).Build()
c.ShowModal(modal)
```

`FileUpload().FileTypes(...)` is a DiscordKit request-only extension: discordgo
does not model `file_types`, so DiscordKit marshals it for you while submissions
still decode into `discordgo.FileUpload`.

Read submissions without reflection through `Context.Form()`:

```go
f := c.Form()
subject, _ := f.String("subject")
files, _ := f.Files("file")   // resolved []*discordgo.MessageAttachment
```

## Responses, followups and autocomplete

`Context` tracks the response lifecycle so an accidental double-acknowledge fails
loudly with a sentinel error instead of a confusing Discord API error:

| Method | Purpose |
| --- | --- |
| `Reply` / `ReplyText` | Initial message response |
| `Ephemeral` / `EphemeralText` | Initial response only the caller sees |
| `Defer(ephemeral)` | Acknowledge a command, edit later |
| `DeferUpdate` | Acknowledge a component without changing the message |
| `Update` | Replace a component's message |
| `Edit` | Edit the original response after Reply/Defer |
| `Followup` / `FollowupEdit` | Additional messages |
| `DeleteResponse` | Delete the original response |
| `Autocomplete(...ChoiceValue)` | Reply to an autocomplete interaction |
| `ShowModal(Modal)` | Open a modal |

Calling an initial response twice returns `ErrAlreadyResponded`; calling
`Edit`/`Followup` before acknowledging returns `ErrNotResponded`; `Autocomplete`
and `ShowModal` guard against the wrong interaction type.

## Error handling

DiscordKit exposes sentinel errors for use with `errors.Is`: `ErrRouteNotFound`,
`ErrRouteConflict`, `ErrInvalidCustomID`, `ErrAlreadyResponded`,
`ErrNotResponded`, `ErrInvalidInteraction`, `ErrInvalidComponent`,
`ErrInvalidMessage`, `ErrMissingOption`, and `ErrInvalidCommand`. Register a
router-wide handler with `Router.OnError`.

## discordgo interop

DiscordKit is additive. You keep your existing `*discordgo.Session`, add
`router.Handle` with `Session.AddHandler`, and use discordgo types everywhere.
Every builder has a `Build`/`component`/conversion method that returns plain
discordgo structs, and `Raw` / the exported `Context.Session` and
`Context.Interaction` fields give you a full escape hatch at any time.

## Compatibility

See [`docs/compatibility.md`](docs/compatibility.md) for the mapping between
Discord API features, discordgo support, and DiscordKit coverage, including the
components DiscordKit intentionally does not build (Radio Group, Checkbox Group
and Checkbox, which discordgo cannot yet decode).

## License

MIT. See [LICENSE](LICENSE).
