package discordkit

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// ChoiceValue is a single predefined choice for a command option or an
// autocomplete response. Build one with Choice.
type ChoiceValue struct {
	name              string
	value             any
	nameLocalizations map[discordgo.Locale]string
}

// Choice creates a named choice. The value must be a string, integer or float.
func Choice(name string, value any) ChoiceValue {
	return ChoiceValue{name: name, value: value}
}

// Localized attaches name localizations to a choice.
func (c ChoiceValue) Localized(localizations map[discordgo.Locale]string) ChoiceValue {
	c.nameLocalizations = localizations
	return c
}

func (c ChoiceValue) build() (*discordgo.ApplicationCommandOptionChoice, error) {
	if l := len(c.name); l < 1 || l > 100 {
		return nil, fmt.Errorf("%w: choice name must be 1-100 characters", ErrInvalidCommand)
	}
	switch c.value.(type) {
	case string, int, int64, float64:
	default:
		return nil, fmt.Errorf("%w: choice %q value must be string, int or float", ErrInvalidCommand, c.name)
	}
	return &discordgo.ApplicationCommandOptionChoice{
		Name:              c.name,
		NameLocalizations: c.nameLocalizations,
		Value:             c.value,
	}, nil
}

// buildChoices validates and converts a set of choices, enforcing the 25-choice
// autocomplete limit.
func buildChoices(choices []ChoiceValue) ([]*discordgo.ApplicationCommandOptionChoice, error) {
	if len(choices) > 25 {
		return nil, fmt.Errorf("%w: at most 25 choices allowed", ErrInvalidCommand)
	}
	out := make([]*discordgo.ApplicationCommandOptionChoice, 0, len(choices))
	for _, c := range choices {
		built, err := c.build()
		if err != nil {
			return nil, err
		}
		out = append(out, built)
	}
	return out, nil
}

// Option is a command option builder that produces a discordgo option.
type Option interface {
	buildOption() (*discordgo.ApplicationCommandOption, error)
}

func validOptionName(name string) error {
	if l := len(name); l < 1 || l > 32 {
		return fmt.Errorf("%w: option name %q must be 1-32 characters", ErrInvalidCommand, name)
	}
	return nil
}

func validOptionDescription(desc string) error {
	if l := len(desc); l < 1 || l > 100 {
		return fmt.Errorf("%w: option description must be 1-100 characters", ErrInvalidCommand)
	}
	return nil
}

// baseOption holds fields shared by every leaf option builder.
type baseOption struct {
	opt discordgo.ApplicationCommandOption
	err error
}

func (b *baseOption) buildBase(kind discordgo.ApplicationCommandOptionType, name, description string) {
	b.opt.Type = kind
	b.opt.Name = name
	b.opt.Description = description
	if err := validOptionName(name); err != nil {
		b.err = err
		return
	}
	if err := validOptionDescription(description); err != nil {
		b.err = err
	}
}

// StringOptionBuilder builds a string command option.
type StringOptionBuilder struct {
	baseOption
	choices []ChoiceValue
}

// StringOption starts a string option builder.
func StringOption(name, description string) *StringOptionBuilder {
	b := &StringOptionBuilder{}
	b.buildBase(discordgo.ApplicationCommandOptionString, name, description)
	return b
}

// Required marks the option as required.
func (b *StringOptionBuilder) Required() *StringOptionBuilder { b.opt.Required = true; return b }

// MinLen sets the minimum string length.
func (b *StringOptionBuilder) MinLen(n int) *StringOptionBuilder {
	if n < 0 || n > 6000 {
		b.err = fmt.Errorf("%w: min length out of range", ErrInvalidCommand)
	}
	b.opt.MinLength = &n
	return b
}

// MaxLen sets the maximum string length.
func (b *StringOptionBuilder) MaxLen(n int) *StringOptionBuilder {
	if n < 1 || n > 6000 {
		b.err = fmt.Errorf("%w: max length out of range", ErrInvalidCommand)
	}
	b.opt.MaxLength = n
	return b
}

// Choices restricts the option to a predefined set. Mutually exclusive with Autocomplete.
func (b *StringOptionBuilder) Choices(choices ...ChoiceValue) *StringOptionBuilder {
	b.choices = choices
	return b
}

// Autocomplete enables dynamic autocomplete. Mutually exclusive with Choices.
func (b *StringOptionBuilder) Autocomplete() *StringOptionBuilder {
	b.opt.Autocomplete = true
	return b
}

// NameLocalizations sets localized names.
func (b *StringOptionBuilder) NameLocalizations(m map[discordgo.Locale]string) *StringOptionBuilder {
	b.opt.NameLocalizations = m
	return b
}

func (b *StringOptionBuilder) buildOption() (*discordgo.ApplicationCommandOption, error) {
	if b.err != nil {
		return nil, b.err
	}
	choices, err := buildChoices(b.choices)
	if err != nil {
		return nil, err
	}
	if len(choices) > 0 && b.opt.Autocomplete {
		return nil, fmt.Errorf("%w: choices and autocomplete are mutually exclusive", ErrInvalidCommand)
	}
	opt := b.opt
	opt.Choices = choices
	return &opt, nil
}

// IntegerOptionBuilder builds an integer command option.
type IntegerOptionBuilder struct {
	baseOption
	choices  []ChoiceValue
	hasMax   bool
	maxValue int64
}

// IntegerOption starts an integer option builder.
func IntegerOption(name, description string) *IntegerOptionBuilder {
	b := &IntegerOptionBuilder{}
	b.buildBase(discordgo.ApplicationCommandOptionInteger, name, description)
	return b
}

// Required marks the option as required.
func (b *IntegerOptionBuilder) Required() *IntegerOptionBuilder { b.opt.Required = true; return b }

// Min sets the minimum value.
func (b *IntegerOptionBuilder) Min(n int64) *IntegerOptionBuilder {
	v := float64(n)
	b.opt.MinValue = &v
	return b
}

// Max sets the maximum value.
func (b *IntegerOptionBuilder) Max(n int64) *IntegerOptionBuilder {
	b.hasMax = true
	b.maxValue = n
	return b
}

// Choices restricts the option to a predefined set. Mutually exclusive with Autocomplete.
func (b *IntegerOptionBuilder) Choices(choices ...ChoiceValue) *IntegerOptionBuilder {
	b.choices = choices
	return b
}

// Autocomplete enables dynamic autocomplete. Mutually exclusive with Choices.
func (b *IntegerOptionBuilder) Autocomplete() *IntegerOptionBuilder {
	b.opt.Autocomplete = true
	return b
}

// NameLocalizations sets localized names.
func (b *IntegerOptionBuilder) NameLocalizations(m map[discordgo.Locale]string) *IntegerOptionBuilder {
	b.opt.NameLocalizations = m
	return b
}

func (b *IntegerOptionBuilder) buildOption() (*discordgo.ApplicationCommandOption, error) {
	if b.err != nil {
		return nil, b.err
	}
	choices, err := buildChoices(b.choices)
	if err != nil {
		return nil, err
	}
	if len(choices) > 0 && b.opt.Autocomplete {
		return nil, fmt.Errorf("%w: choices and autocomplete are mutually exclusive", ErrInvalidCommand)
	}
	opt := b.opt
	if b.hasMax {
		if b.maxValue == 0 {
			return nil, fmt.Errorf("%w: an explicit maximum of zero cannot be represented by discordgo", ErrInvalidCommand)
		}
		opt.MaxValue = float64(b.maxValue)
	}
	opt.Choices = choices
	return &opt, nil
}

// NumberOptionBuilder builds a floating-point command option.
type NumberOptionBuilder struct {
	baseOption
	choices  []ChoiceValue
	hasMax   bool
	maxValue float64
}

// NumberOption starts a number (float) option builder.
func NumberOption(name, description string) *NumberOptionBuilder {
	b := &NumberOptionBuilder{}
	b.buildBase(discordgo.ApplicationCommandOptionNumber, name, description)
	return b
}

// Required marks the option as required.
func (b *NumberOptionBuilder) Required() *NumberOptionBuilder { b.opt.Required = true; return b }

// Min sets the minimum value.
func (b *NumberOptionBuilder) Min(n float64) *NumberOptionBuilder { b.opt.MinValue = &n; return b }

// Max sets the maximum value.
func (b *NumberOptionBuilder) Max(n float64) *NumberOptionBuilder {
	b.hasMax = true
	b.maxValue = n
	return b
}

// Choices restricts the option to a predefined set. Mutually exclusive with Autocomplete.
func (b *NumberOptionBuilder) Choices(choices ...ChoiceValue) *NumberOptionBuilder {
	b.choices = choices
	return b
}

// Autocomplete enables dynamic autocomplete. Mutually exclusive with Choices.
func (b *NumberOptionBuilder) Autocomplete() *NumberOptionBuilder {
	b.opt.Autocomplete = true
	return b
}

// NameLocalizations sets localized names.
func (b *NumberOptionBuilder) NameLocalizations(m map[discordgo.Locale]string) *NumberOptionBuilder {
	b.opt.NameLocalizations = m
	return b
}

func (b *NumberOptionBuilder) buildOption() (*discordgo.ApplicationCommandOption, error) {
	if b.err != nil {
		return nil, b.err
	}
	choices, err := buildChoices(b.choices)
	if err != nil {
		return nil, err
	}
	if len(choices) > 0 && b.opt.Autocomplete {
		return nil, fmt.Errorf("%w: choices and autocomplete are mutually exclusive", ErrInvalidCommand)
	}
	opt := b.opt
	if b.hasMax {
		if b.maxValue == 0 {
			return nil, fmt.Errorf("%w: an explicit maximum of zero cannot be represented by discordgo", ErrInvalidCommand)
		}
		opt.MaxValue = b.maxValue
	}
	opt.Choices = choices
	return &opt, nil
}

// simpleOptionBuilder builds boolean, user, role, mentionable and attachment
// options, which accept only a name, description and required flag.
type simpleOptionBuilder struct {
	baseOption
}

func newSimpleOption(kind discordgo.ApplicationCommandOptionType, name, description string) *simpleOptionBuilder {
	b := &simpleOptionBuilder{}
	b.buildBase(kind, name, description)
	return b
}

// Required marks the option as required.
func (b *simpleOptionBuilder) Required() *simpleOptionBuilder { b.opt.Required = true; return b }

// NameLocalizations sets localized names.
func (b *simpleOptionBuilder) NameLocalizations(m map[discordgo.Locale]string) *simpleOptionBuilder {
	b.opt.NameLocalizations = m
	return b
}

func (b *simpleOptionBuilder) buildOption() (*discordgo.ApplicationCommandOption, error) {
	if b.err != nil {
		return nil, b.err
	}
	opt := b.opt
	return &opt, nil
}

// BooleanOption starts a boolean option builder.
func BooleanOption(name, description string) *simpleOptionBuilder {
	return newSimpleOption(discordgo.ApplicationCommandOptionBoolean, name, description)
}

// UserOption starts a user option builder.
func UserOption(name, description string) *simpleOptionBuilder {
	return newSimpleOption(discordgo.ApplicationCommandOptionUser, name, description)
}

// RoleOption starts a role option builder.
func RoleOption(name, description string) *simpleOptionBuilder {
	return newSimpleOption(discordgo.ApplicationCommandOptionRole, name, description)
}

// MentionableOption starts a mentionable (user or role) option builder.
func MentionableOption(name, description string) *simpleOptionBuilder {
	return newSimpleOption(discordgo.ApplicationCommandOptionMentionable, name, description)
}

// AttachmentOption starts an attachment option builder.
func AttachmentOption(name, description string) *simpleOptionBuilder {
	return newSimpleOption(discordgo.ApplicationCommandOptionAttachment, name, description)
}

// ChannelOptionBuilder builds a channel command option with channel type limits.
type ChannelOptionBuilder struct {
	baseOption
}

// ChannelOption starts a channel option builder.
func ChannelOption(name, description string) *ChannelOptionBuilder {
	b := &ChannelOptionBuilder{}
	b.buildBase(discordgo.ApplicationCommandOptionChannel, name, description)
	return b
}

// Required marks the option as required.
func (b *ChannelOptionBuilder) Required() *ChannelOptionBuilder { b.opt.Required = true; return b }

// ChannelTypes restricts selectable channel types.
func (b *ChannelOptionBuilder) ChannelTypes(types ...discordgo.ChannelType) *ChannelOptionBuilder {
	b.opt.ChannelTypes = types
	return b
}

// NameLocalizations sets localized names.
func (b *ChannelOptionBuilder) NameLocalizations(m map[discordgo.Locale]string) *ChannelOptionBuilder {
	b.opt.NameLocalizations = m
	return b
}

func (b *ChannelOptionBuilder) buildOption() (*discordgo.ApplicationCommandOption, error) {
	if b.err != nil {
		return nil, b.err
	}
	opt := b.opt
	return &opt, nil
}
