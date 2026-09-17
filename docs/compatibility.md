# Compatibility audit — 2026-09-17

DiscordKit is a framework layer built on top of discordgo. It does not replace
its REST, Gateway, voice, session, cache or fundamental resource models.

The upstream master head was checked through GitHub's commits API before
implementation: `f43dd94faaacd5b163e9e783f14b5bd8be639fc9`, dated 2026-02-14.
The minimum supported module version is
`v0.29.1-0.20260214123928-f43dd94faaac`. No fork or replace directive is needed.

Audited sources:

- https://github.com/bwmarrin/discordgo/commit/f43dd94faaacd5b163e9e783f14b5bd8be639fc9
- https://docs.discord.com/developers/components/reference
- https://docs.discord.com/developers/interactions/receiving-and-responding
- https://docs.discord.com/developers/interactions/application-commands

Components 1–14 (excluding nonexistent types), 17, 18 and 19 are represented
and decoded upstream. Label, FileUpload and resolved modal data are available.
Radio Group (21), Checkbox Group (22), and Checkbox (23) are documented by
Discord but rejected by discordgo's component decoder. DiscordKit deliberately
does not provide one-way builders for them.

FileUpload.file_types is request-only and absent upstream; an isolated component
extension can marshal it while submissions still decode into discordgo.FileUpload.
Filtering is by extension, not content validation.

Other upstream limitations: WebhookParams and WebhookEdit have no Poll field;
DiscordKit must reject polls for these conversions instead of silently dropping
them. ApplicationCommandOption.MaxValue uses omitempty, so an explicit maximum
of zero cannot be represented by that upstream model. Such a builder setting
must fail explicitly. TextDisplay lacks its optional numeric ID. Activity launch
and interaction callback with_response are outside the v0.1.0 framework scope.

Message Components V2 is irreversible once enabled on a message. It disables
legacy content, embeds, polls and stickers. Content is normalized to TextDisplay;
files must be explicitly referenced by media/file components to be rendered.
A message permits 40 total components, including nested layout components.
Modern modal inputs use Label; deprecated ActionRow/TextInput is not generated.
