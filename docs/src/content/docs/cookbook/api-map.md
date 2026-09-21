---
title: "Cookbook API map"
description: "Coverage, variants and boundaries of the API used in this project."
---

This map groups the public API by family. **Core usage, lab experiments and explained variants are different levels of coverage.** The support workflow implements the core; the lab demonstrates complementary features. Cosmetic and monetization-dependent variants are described without pretending they are implemented product features.

| API | Chapter and application |
| --- | --- |
| Command, Options, Build, MustBuild, BuildCommands | [Definitions and validation; MustBuild panics on errors.](/cookbook/commands/) |
| SubCommand, SubCommandGroup | [Ticket groups actions; lab options inspect adds another level.](/cookbook/laboratory/) |
| UserCommand, MessageCommand | [Context menus; target comes from discordgo TargetID/resolved.](/cookbook/laboratory/) |
| NameLocalizations, DescriptionLocalizations, Choice.Localized | [Lab localizes its name. Descriptions and choices use Locale maps; router paths remain canonical.](/cookbook/laboratory/) |
| DefaultMemberPermissions, Contexts, IntegrationTypes, NSFW | [Panel uses default permissions. Contexts/IntegrationTypes control distribution, primarily for global commands; NSFW is unnecessary here. Do not blindly copy them into guild-scoped sync.](/cookbook/components/) |
| StringOption, IntegerOption, NumberOption, BooleanOption | [ID, rating, weight and visibility; Required, bounds, Choices or Autocomplete.](/cookbook/commands/) |
| UserOption, RoleOption, ChannelOption, MentionableOption, AttachmentOption | [Resolved objects; ChannelTypes limits selection.](/cookbook/laboratory/) |
| Choice, Choices, Autocomplete | [Fixed statuses and dynamic lookup; choices and autocomplete are mutually exclusive on one option.](/cookbook/queue/) |
| NewRouter, Use, Group, Command, Component, Modal, Autocomplete | [Feature registration. Use adds global middleware; Group shares a prefix and middleware.](/cookbook/architecture/) |
| Handle, Dispatch, NewContext | [Handle adapts Gateway events; Dispatch/NewContext enable fake-transport testing.](/cookbook/operations/) |
| OnError, Handler, Middleware, Recovery, Logging, PanicError | [Error handling and composition; Recovery is automatic in Router. PanicError stacks belong in protected logs, never public responses.](/cookbook/architecture/) |
| RequireGuild, RequirePermissions | [Guild and invoking-member checks; they do not check bot channel permissions.](/cookbook/components/) |
| Context, SetContext, Session, Interaction | [REST deadline and native escape hatches; do not mix raw and tracked responses on one interaction.](/cookbook/architecture/) |
| User, Member, GuildID, ChannelID, Locale, IsGuild, IsDM | [Event metadata. This bot rejects DMs and guilds outside its configuration.](/cookbook/architecture/) |
| Option, FocusedOption, String, Int, Float, Bool | [Option presence is separate from zero values; FocusedOption identifies active autocomplete input.](/cookbook/queue/) |
| UserOption, MemberOption, Role, Channel, Attachment, Mentionable | [Context getters read resolved data; Mentionable.Valid requires exactly one user or role.](/cookbook/laboratory/) |
| Require… e Must… / Require… and Must… | [Require returns errors; Must panics. Prefer Require for external input. Variants cover primitives, entities and route parameters.](/cookbook/tickets/) |
| CustomID.Validate, Route, Param, Build, MustBuild | [Stable IDs, parameter encoding and 100-character limit; not authorization.](/cookbook/tickets/) |
| Context.Param, RequireParam, MustParam | [Read the parameter decoded by the router.](/cookbook/tickets/) |
| MessageSpec, InteractionResponseData, WebhookParams, WebhookEdit, MessageEdit | [One payload with endpoint-specific conversions; Files/Attachments/AllowedMentions are explicit fields.](/cookbook/laboratory/) |
| Reply, ReplyText, Ephemeral, EphemeralText | [Initial response and visibility; text methods are MessageSpec conveniences.](/cookbook/commands/) |
| Defer, Edit, DeferUpdate, Update | [Defer before disk I/O; Update for immediate navigation; Edit after acknowledgment.](/cookbook/tickets/) |
| Followup, FollowupEdit, DeleteResponse | [Additional export message, followup revision and dismiss button.](/cookbook/operations/) |
| Context.Autocomplete, ShowModal | [Specialized initial responses; they do not follow a message defer.](/cookbook/tickets/) |
| Components, Raw, Text, Row, Container, Separator | [V2 tree; container accent/spoiler and separator divider/size customize presentation.](/cookbook/components/) |
| Button, LinkButton, PremiumButton | [Routed actions and links in the bot. PremiumButton requires a real SKU/monetization; no custom ID or fake purchase handler.](/cookbook/components/) |
| Primary, Secondary, Success, Danger, Emoji, Disabled | [Button variants: style is not permission. Disabled prevents normal clicks but the server still revalidates state.](/cookbook/components/) |
| SelectOption, StringSelect | [Options, Description, Emoji, Default and Placeholder; validate values in handlers.](/cookbook/queue/) |
| UserSelect, RoleSelect, MentionableSelect, ChannelSelect | [Entity selection, bounds, defaults and channel types; the lab demonstrates three and exercises the mentionable variant.](/cookbook/laboratory/) |
| Section, Thumbnail, Gallery, MediaItem, File | [V2 media, descriptions, spoiler and attachment://; sections require an accessory.](/cookbook/laboratory/) |
| Form, Fields, Field, TextInput, Modal, ModalOf | [Modal builders and native conversion; input values, placeholders, paragraph, required and bounds configure fields.](/cookbook/tickets/) |
| FileUpload, FileTypes, MinValues, MaxValues, Required | [Form upload. FileTypes constrains UI; server-side size/type validation is still needed before processing files.](/cookbook/laboratory/) |
| Context.Form, FormData.String, Strings, Files, RequireString, MustString, RequireStrings, RequireFiles | [Text and resolved form values; Require methods detect missing input.](/cookbook/tickets/) |
| ValidateMessageComponents, ValidateModalComponents | [Local validation, also used by builders/converters. It does not replace remote validation.](/cookbook/components/) |
| SyncCommands, SyncOptions, DiffCommands, SyncPlan.Empty | [Non-destructive sync by default; preview with remote list + DiffCommands. The actual deletion option is named Delete.](/cookbook/laboratory/) |

## Errors are part of the contract

Use `errors.Is` with errors.go sentinels and `errors.As` for PanicError instead of matching text. ErrAlreadyResponded means Context initiated a response, not that Discord received it. ErrNotResponded means a later operation was attempted before acknowledgment. Configuration-related route and validation errors should fail tests or startup.

## Deliberate boundaries

Persistence, migrations, queues, domain permissions and quotas belong to the application. The cookbook adds none of these services to the root package. Raw does not create support for Discord features missing from the pinned discordgo version. Consult the library reference and upstream notes before experimenting with new components.

[Return to the learning path](/cookbook/support-bot/).
