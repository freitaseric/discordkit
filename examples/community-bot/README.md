# Community support bot

A modular, runnable DiscordKit cookbook project. It registers tickets with private
views, staff claiming, owner/staff closing, owner ratings, filtering, autocomplete
and authorized JSON exports. It does not create private ticket channels or relay
conversations. Bot UI text is Portuguese; the tutorial is available in both languages.

- [Português: guia completo](https://discordkit.freitaseric.com/pt-br/cookbook/)
- [English: complete guide](https://discordkit.freitaseric.com/cookbook/)

## Run

From this directory, with Go 1.26+:

```bash
read -rs -p 'Bot token: ' DISCORD_TOKEN; echo
export DISCORD_TOKEN
export DISCORD_GUILD_ID='YOUR_TEST_GUILD_ID'
export BOT_DATA_FILE='data/community.json'
export COOKBOOK_STAGE=1
go run ./cmd/bot
```

`.env.example` documents variables; the application does not load `.env` files.
Use the existing repository module, not `go mod init` in this directory.

| Stage | Adds |
| --- | --- |
| 1 | Gateway, lifecycle, /ping |
| 2 | /about and typed boolean option |
| 3 | Private help, permission-protected public panel, topic select |
| 4 | Ticket modal, persistent records, show, autocomplete, claim and close |
| 5 | Private queue, status filters and pagination |
| 6 (default) | Rating, private export/dismiss and optional /lab experiments |

Stop before changing stages. Sync does not delete unrelated commands, so moving
backwards can leave commands visible without active handlers. Use a dedicated
test application/guild. Publishing a panel is explicit; startup never creates one.

Staff means Manage Messages or Administrator. Every store operation checks guild
and actor access. Only the owner rates a closed ticket. Replaying the same modal
interaction ID does not duplicate its ticket. Fresh submissions do create new ones.

## Verify

```bash
go test ./...
go test -race ./...
go vet ./...
```

Tests are offline, including a fake HTTP transport for modal → defer → edit → disk.
Test installation, client rendering, permissions and exports in a real test guild
before relying on this deployment. See the guide's manual acceptance checklist.

## Storage boundaries

The JSON store is part of this example, not a library database API. Run **one
process**, preferably Linux on a local persistent filesystem. It uses an in-process
mutex and temporary-file + sync + rename writes. It has no distributed lock,
directory fsync, automatic retention, anti-spam or unlimited capacity. Every write
rewrites all records. Corrupt/unsupported files fail startup instead of being reset.

Use an absolute data path in production. Never commit credentials or ticket data.
Stop before backups, restore, or manual edits. Use SQLite/PostgreSQL when scale or
stronger durability warrants it. Vercel serves the documentation; run the Gateway
bot on a continuously running worker, VM or container with a persistent volume.

## Maintaining documentation

After changing the example, from the repository root:

```bash
python3 scripts/sync-cookbook.py
python3 scripts/sync-cookbook.py --check
```

The CI gate prevents tutorial listings from drifting away from the executable code.
