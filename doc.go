// Package discordkit provides an interaction framework built on top of discordgo.
// discordgo owns Discord models and transport; DiscordKit provides routing,
// typed command options, component validation, forms and response lifecycle.
//
// Register routes before connecting a discordgo.Session. Attach a Router with
// Session.AddHandler(router.Handle). Handlers return errors to Router.OnError.
// Recovery is enabled by default. Access Context.Session for raw discordgo APIs.
package discordkit
