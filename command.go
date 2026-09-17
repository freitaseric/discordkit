package discordkit

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// CommandBuilder builds a *discordgo.ApplicationCommand with compile-time typed
// option builders and validation deferred to Build.
type CommandBuilder struct {
	cmd     discordgo.ApplicationCommand
	options []Option
	err     error
}

// Command starts a chat (slash) command builder. Options may also be added with
// Options for readability.
func Command(name, description string, options ...Option) *CommandBuilder {
	b := &CommandBuilder{cmd: discordgo.ApplicationCommand{
		Type:        discordgo.ChatApplicationCommand,
		Name:        name,
		Description: description,
	}}
	if l := len(name); l < 1 || l > 32 {
		b.err = fmt.Errorf("%w: command name must be 1-32 characters", ErrInvalidCommand)
	} else if l := len(description); l < 1 || l > 100 {
		b.err = fmt.Errorf("%w: command description must be 1-100 characters", ErrInvalidCommand)
	}
	b.options = append(b.options, options...)
	return b
}

// UserCommand starts a user context menu command builder.
func UserCommand(name string) *CommandBuilder {
	b := &CommandBuilder{cmd: discordgo.ApplicationCommand{Type: discordgo.UserApplicationCommand, Name: name}}
	if l := len(name); l < 1 || l > 32 {
		b.err = fmt.Errorf("%w: command name must be 1-32 characters", ErrInvalidCommand)
	}
	return b
}

// MessageCommand starts a message context menu command builder.
func MessageCommand(name string) *CommandBuilder {
	b := &CommandBuilder{cmd: discordgo.ApplicationCommand{Type: discordgo.MessageApplicationCommand, Name: name}}
	if l := len(name); l < 1 || l > 32 {
		b.err = fmt.Errorf("%w: command name must be 1-32 characters", ErrInvalidCommand)
	}
	return b
}

// Options appends command options.
func (b *CommandBuilder) Options(options ...Option) *CommandBuilder {
	b.options = append(b.options, options...)
	return b
}

// NameLocalizations sets localized command names.
func (b *CommandBuilder) NameLocalizations(m map[discordgo.Locale]string) *CommandBuilder {
	loc := m
	b.cmd.NameLocalizations = &loc
	return b
}

// DescriptionLocalizations sets localized command descriptions.
func (b *CommandBuilder) DescriptionLocalizations(m map[discordgo.Locale]string) *CommandBuilder {
	b.cmd.DescriptionLocalizations = &m
	return b
}

// DefaultMemberPermissions sets the default member permission bitmask required
// to see and use the command.
func (b *CommandBuilder) DefaultMemberPermissions(perms int64) *CommandBuilder {
	p := perms
	b.cmd.DefaultMemberPermissions = &p
	return b
}

// NSFW marks the command as age-restricted.
func (b *CommandBuilder) NSFW() *CommandBuilder {
	nsfw := true
	b.cmd.NSFW = &nsfw
	return b
}

// Contexts sets the interaction contexts where the command is available.
func (b *CommandBuilder) Contexts(contexts ...discordgo.InteractionContextType) *CommandBuilder {
	c := contexts
	b.cmd.Contexts = &c
	return b
}

// IntegrationTypes sets the installation contexts for the command.
func (b *CommandBuilder) IntegrationTypes(types ...discordgo.ApplicationIntegrationType) *CommandBuilder {
	t := types
	b.cmd.IntegrationTypes = &t
	return b
}

// Build validates and returns the discordgo command.
func (b *CommandBuilder) Build() (*discordgo.ApplicationCommand, error) {
	if b.err != nil {
		return nil, b.err
	}
	if b.cmd.Type != discordgo.ChatApplicationCommand && len(b.options) > 0 {
		return nil, fmt.Errorf("%w: only chat commands may declare options", ErrInvalidCommand)
	}
	opts, err := buildOptions(b.options)
	if err != nil {
		return nil, err
	}
	cmd := b.cmd
	cmd.Options = opts
	return &cmd, nil
}

// MustBuild is Build with panic on error, for static command declarations.
func (b *CommandBuilder) MustBuild() *discordgo.ApplicationCommand {
	cmd, err := b.Build()
	if err != nil {
		panic(err)
	}
	return cmd
}

// buildOptions validates the shared rules for a set of sibling options: the
// 25-option limit, unique names, and required options preceding optional ones.
func buildOptions(options []Option) ([]*discordgo.ApplicationCommandOption, error) {
	if len(options) > 25 {
		return nil, fmt.Errorf("%w: at most 25 options allowed", ErrInvalidCommand)
	}
	out := make([]*discordgo.ApplicationCommandOption, 0, len(options))
	seen := map[string]bool{}
	sawOptional := false
	for _, o := range options {
		built, err := o.buildOption()
		if err != nil {
			return nil, err
		}
		if seen[built.Name] {
			return nil, fmt.Errorf("%w: duplicate option name %q", ErrInvalidCommand, built.Name)
		}
		seen[built.Name] = true
		isContainer := built.Type == discordgo.ApplicationCommandOptionSubCommand || built.Type == discordgo.ApplicationCommandOptionSubCommandGroup
		if !isContainer {
			if built.Required && sawOptional {
				return nil, fmt.Errorf("%w: required option %q must precede optional options", ErrInvalidCommand, built.Name)
			}
			if !built.Required {
				sawOptional = true
			}
		}
		out = append(out, built)
	}
	return out, nil
}

// SubCommandBuilder builds a subcommand option.
type SubCommandBuilder struct {
	opt     discordgo.ApplicationCommandOption
	options []Option
	err     error
}

// SubCommand starts a subcommand with its own leaf options.
func SubCommand(name, description string, options ...Option) *SubCommandBuilder {
	b := &SubCommandBuilder{opt: discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        name,
		Description: description,
	}}
	if err := validOptionName(name); err != nil {
		b.err = err
	} else if err := validOptionDescription(description); err != nil {
		b.err = err
	}
	b.options = options
	return b
}

// NameLocalizations sets localized names.
func (b *SubCommandBuilder) NameLocalizations(m map[discordgo.Locale]string) *SubCommandBuilder {
	b.opt.NameLocalizations = m
	return b
}

func (b *SubCommandBuilder) buildOption() (*discordgo.ApplicationCommandOption, error) {
	if b.err != nil {
		return nil, b.err
	}
	for _, o := range b.options {
		if _, ok := o.(*SubCommandBuilder); ok {
			return nil, fmt.Errorf("%w: subcommands cannot be nested in a subcommand", ErrInvalidCommand)
		}
		if _, ok := o.(*SubCommandGroupBuilder); ok {
			return nil, fmt.Errorf("%w: subcommand groups cannot be nested in a subcommand", ErrInvalidCommand)
		}
	}
	opts, err := buildOptions(b.options)
	if err != nil {
		return nil, err
	}
	opt := b.opt
	opt.Options = opts
	return &opt, nil
}

// SubCommandGroupBuilder builds a subcommand group option.
type SubCommandGroupBuilder struct {
	opt         discordgo.ApplicationCommandOption
	subcommands []*SubCommandBuilder
	err         error
}

// SubCommandGroup starts a subcommand group containing subcommands.
func SubCommandGroup(name, description string, subcommands ...*SubCommandBuilder) *SubCommandGroupBuilder {
	b := &SubCommandGroupBuilder{opt: discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
		Name:        name,
		Description: description,
	}}
	if err := validOptionName(name); err != nil {
		b.err = err
	} else if err := validOptionDescription(description); err != nil {
		b.err = err
	}
	b.subcommands = subcommands
	return b
}

// NameLocalizations sets localized names.
func (b *SubCommandGroupBuilder) NameLocalizations(m map[discordgo.Locale]string) *SubCommandGroupBuilder {
	b.opt.NameLocalizations = m
	return b
}

func (b *SubCommandGroupBuilder) buildOption() (*discordgo.ApplicationCommandOption, error) {
	if b.err != nil {
		return nil, b.err
	}
	if len(b.subcommands) == 0 {
		return nil, fmt.Errorf("%w: subcommand group %q requires at least one subcommand", ErrInvalidCommand, b.opt.Name)
	}
	options := make([]Option, len(b.subcommands))
	for i, s := range b.subcommands {
		options[i] = s
	}
	opts, err := buildOptions(options)
	if err != nil {
		return nil, err
	}
	opt := b.opt
	opt.Options = opts
	return &opt, nil
}
