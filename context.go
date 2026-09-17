package discordkit

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"

	"github.com/bwmarrin/discordgo"
)

// Context represents one interaction. Session and Interaction are escape hatches.
// Do not copy Context after use. Raw response calls bypass its lifecycle tracking.
type Context struct {
	Session      *discordgo.Session
	Interaction  *discordgo.InteractionCreate
	ctx          context.Context
	params       map[string]string
	mu           sync.Mutex
	state        responseState
	responseKind discordgo.InteractionResponseType
	v2           bool
	ephemeral    bool
}

// NewContext creates a pending interaction context. Router normally calls this.
func NewContext(s *discordgo.Session, i *discordgo.InteractionCreate) *Context {
	return &Context{Session: s, Interaction: i, ctx: context.Background()}
}

// Context returns the standard context used by DiscordKit's REST requests.
func (c *Context) Context() context.Context {
	if c.ctx == nil {
		return context.Background()
	}
	return c.ctx
}

// SetContext sets the request context before response operations; nil means Background.
func (c *Context) SetContext(ctx context.Context) { c.ctx = ctx }

// Param returns a decoded route parameter.
func (c *Context) Param(name string) (string, bool) { v, ok := c.params[name]; return v, ok }
func (c *Context) interaction() *discordgo.Interaction {
	if c == nil || c.Interaction == nil {
		return nil
	}
	return c.Interaction.Interaction
}

// User returns the invoking user, in guilds or DMs, without network access.
func (c *Context) User() *discordgo.User {
	i := c.interaction()
	if i == nil {
		return nil
	}
	if i.Member != nil {
		return i.Member.User
	}
	return i.User
}

// Member returns the invoking guild member, or nil in DMs.
func (c *Context) Member() *discordgo.Member {
	i := c.interaction()
	if i == nil {
		return nil
	}
	return i.Member
}

// GuildID returns the guild ID, or an empty string in DMs.
func (c *Context) GuildID() string {
	i := c.interaction()
	if i == nil {
		return ""
	}
	return i.GuildID
}

// ChannelID returns the interaction's channel ID.
func (c *Context) ChannelID() string {
	i := c.interaction()
	if i == nil {
		return ""
	}
	return i.ChannelID
}

// Locale returns the invoking user's locale.
func (c *Context) Locale() discordgo.Locale {
	i := c.interaction()
	if i == nil {
		return ""
	}
	return i.Locale
}

// IsGuild reports whether the interaction was invoked in a guild.
func (c *Context) IsGuild() bool { return c.GuildID() != "" }

// IsDM reports whether a non-nil interaction was invoked outside a guild.
func (c *Context) IsDM() bool { return c.interaction() != nil && !c.IsGuild() }
func (c *Context) commandData() (discordgo.ApplicationCommandInteractionData, bool) {
	i := c.interaction()
	if i == nil {
		return discordgo.ApplicationCommandInteractionData{}, false
	}
	switch d := i.Data.(type) {
	case discordgo.ApplicationCommandInteractionData:
		return d, true
	case *discordgo.ApplicationCommandInteractionData:
		if d != nil {
			return *d, true
		}
	}
	return discordgo.ApplicationCommandInteractionData{}, false
}
func leafOptions(opts []*discordgo.ApplicationCommandInteractionDataOption) []*discordgo.ApplicationCommandInteractionDataOption {
	for _, o := range opts {
		if o != nil && (o.Type == 1 || o.Type == 2) {
			return leafOptions(o.Options)
		}
	}
	return opts
}

// Option returns a leaf option from the selected subcommand, without fetching data.
func (c *Context) Option(name string) (*discordgo.ApplicationCommandInteractionDataOption, bool) {
	d, ok := c.commandData()
	if !ok {
		return nil, false
	}
	for _, o := range leafOptions(d.Options) {
		if o != nil && o.Name == name {
			return o, true
		}
	}
	return nil, false
}

// FocusedOption returns the currently focused autocomplete option.
func (c *Context) FocusedOption() (*discordgo.ApplicationCommandInteractionDataOption, bool) {
	d, ok := c.commandData()
	if !ok {
		return nil, false
	}
	for _, o := range leafOptions(d.Options) {
		if o != nil && o.Focused {
			return o, true
		}
	}
	return nil, false
}
func (c *Context) value(name string, t discordgo.ApplicationCommandOptionType) (any, bool) {
	o, ok := c.Option(name)
	if !ok || o.Type != t {
		return nil, false
	}
	return o.Value, true
}

// String reads a string option, including a supplied empty string.
func (c *Context) String(name string) (string, bool) {
	v, _ := c.value(name, 3)
	s, ok := v.(string)
	return s, ok
}

// Int reads an integer option without truncation or precision loss.
func (c *Context) Int(name string) (int64, bool) {
	v, ok := c.value(name, 4)
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		if math.IsNaN(n) || math.IsInf(n, 0) || math.Trunc(n) != n || math.Abs(n) > 9007199254740992 {
			return 0, false
		}
		return int64(n), true
	case int64:
		return n, n >= -9007199254740992 && n <= 9007199254740992
	case int:
		return int64(n), int64(n) >= -9007199254740992 && int64(n) <= 9007199254740992
	case json.Number:
		x, e := n.Int64()
		return x, e == nil && x >= -9007199254740992 && x <= 9007199254740992
	}
	return 0, false
}

// Float reads a finite number option.
func (c *Context) Float(name string) (float64, bool) {
	v, _ := c.value(name, 10)
	n, ok := v.(float64)
	if j, yes := v.(json.Number); yes {
		var e error
		n, e = j.Float64()
		ok = e == nil
	}
	return n, ok && !math.IsNaN(n) && !math.IsInf(n, 0)
}

// Bool reads a boolean option, preserving a supplied false value.
func (c *Context) Bool(name string) (bool, bool) {
	v, _ := c.value(name, 5)
	b, ok := v.(bool)
	return b, ok
}
func (c *Context) resolvedID(name string, t discordgo.ApplicationCommandOptionType) (string, *discordgo.ApplicationCommandInteractionDataResolved, bool) {
	v, ok := c.value(name, t)
	id, s := v.(string)
	d, has := c.commandData()
	return id, d.Resolved, ok && s && id != "" && has && d.Resolved != nil
}

// UserOption returns a resolved user, never performing an implicit REST request.
func (c *Context) UserOption(name string) (*discordgo.User, bool) {
	id, r, ok := c.resolvedID(name, 6)
	if !ok {
		return nil, false
	}
	u := r.Users[id]
	return u, u != nil
}

// MemberOption joins a resolved member with its user without mutating discordgo data.
func (c *Context) MemberOption(name string) (*discordgo.Member, bool) {
	id, r, ok := c.resolvedID(name, 6)
	if !ok || r.Members[id] == nil {
		return nil, false
	}
	m := *r.Members[id]
	if m.User == nil {
		m.User = r.Users[id]
	}
	return &m, true
}

// Role returns a resolved role option.
func (c *Context) Role(name string) (*discordgo.Role, bool) {
	id, r, ok := c.resolvedID(name, 8)
	if !ok {
		return nil, false
	}
	v := r.Roles[id]
	return v, v != nil
}

// Channel returns a resolved channel option.
func (c *Context) Channel(name string) (*discordgo.Channel, bool) {
	id, r, ok := c.resolvedID(name, 7)
	if !ok {
		return nil, false
	}
	v := r.Channels[id]
	return v, v != nil
}

// Attachment returns a resolved attachment option.
func (c *Context) Attachment(name string) (*discordgo.MessageAttachment, bool) {
	id, r, ok := c.resolvedID(name, 11)
	if !ok {
		return nil, false
	}
	v := r.Attachments[id]
	return v, v != nil
}

// Mentionable is a resolved user-or-role union. Valid reports its invariant.
type Mentionable struct {
	User *discordgo.User
	Role *discordgo.Role
}

// Valid reports whether exactly one of User and Role is present.
func (m Mentionable) Valid() bool { return (m.User != nil) != (m.Role != nil) }

// Mentionable returns a resolved mentionable without assuming it is a user.
func (c *Context) Mentionable(name string) (Mentionable, bool) {
	id, r, ok := c.resolvedID(name, 9)
	if !ok {
		return Mentionable{}, false
	}
	m := Mentionable{User: r.Users[id], Role: r.Roles[id]}
	if !m.Valid() {
		return Mentionable{}, false
	}
	return m, true
}
func optionError(name string) error { return fmt.Errorf("%w: %q", ErrMissingOption, name) }

// RequireString returns an error when the value is absent or invalid.
func (c *Context) RequireString(name string) (string, error) {
	v, ok := c.String(name)
	if !ok {
		return "", optionError(name)
	}
	return v, nil
}

// MustString panics when the value is absent or invalid.
func (c *Context) MustString(name string) string {
	v, e := c.RequireString(name)
	if e != nil {
		panic(e)
	}
	return v
}

// RequireInt returns an error when the value is absent or invalid.
func (c *Context) RequireInt(name string) (int64, error) {
	v, ok := c.Int(name)
	if !ok {
		return 0, optionError(name)
	}
	return v, nil
}

// MustInt panics when the value is absent or invalid.
func (c *Context) MustInt(name string) int64 {
	v, e := c.RequireInt(name)
	if e != nil {
		panic(e)
	}
	return v
}

// RequireFloat returns an error when the value is absent or invalid.
func (c *Context) RequireFloat(name string) (float64, error) {
	v, ok := c.Float(name)
	if !ok {
		return 0, optionError(name)
	}
	return v, nil
}

// MustFloat panics when the value is absent or invalid.
func (c *Context) MustFloat(name string) float64 {
	v, e := c.RequireFloat(name)
	if e != nil {
		panic(e)
	}
	return v
}

// RequireBool returns an error when the value is absent or invalid.
func (c *Context) RequireBool(name string) (bool, error) {
	v, ok := c.Bool(name)
	if !ok {
		return false, optionError(name)
	}
	return v, nil
}

// MustBool panics when the value is absent or invalid.
func (c *Context) MustBool(name string) bool {
	v, e := c.RequireBool(name)
	if e != nil {
		panic(e)
	}
	return v
}

// RequireUserOption returns an error when the value is absent or invalid.
func (c *Context) RequireUserOption(name string) (*discordgo.User, error) {
	v, ok := c.UserOption(name)
	if !ok {
		return nil, optionError(name)
	}
	return v, nil
}

// MustUserOption panics when the value is absent or invalid.
func (c *Context) MustUserOption(name string) *discordgo.User {
	v, e := c.RequireUserOption(name)
	if e != nil {
		panic(e)
	}
	return v
}

// RequireMemberOption returns an error when the value is absent or invalid.
func (c *Context) RequireMemberOption(name string) (*discordgo.Member, error) {
	v, ok := c.MemberOption(name)
	if !ok {
		return nil, optionError(name)
	}
	return v, nil
}

// MustMemberOption panics when the value is absent or invalid.
func (c *Context) MustMemberOption(name string) *discordgo.Member {
	v, e := c.RequireMemberOption(name)
	if e != nil {
		panic(e)
	}
	return v
}

// RequireRole returns an error when the value is absent or invalid.
func (c *Context) RequireRole(name string) (*discordgo.Role, error) {
	v, ok := c.Role(name)
	if !ok {
		return nil, optionError(name)
	}
	return v, nil
}

// MustRole panics when the value is absent or invalid.
func (c *Context) MustRole(name string) *discordgo.Role {
	v, e := c.RequireRole(name)
	if e != nil {
		panic(e)
	}
	return v
}

// RequireChannel returns an error when the value is absent or invalid.
func (c *Context) RequireChannel(name string) (*discordgo.Channel, error) {
	v, ok := c.Channel(name)
	if !ok {
		return nil, optionError(name)
	}
	return v, nil
}

// MustChannel panics when the value is absent or invalid.
func (c *Context) MustChannel(name string) *discordgo.Channel {
	v, e := c.RequireChannel(name)
	if e != nil {
		panic(e)
	}
	return v
}

// RequireAttachment returns an error when the value is absent or invalid.
func (c *Context) RequireAttachment(name string) (*discordgo.MessageAttachment, error) {
	v, ok := c.Attachment(name)
	if !ok {
		return nil, optionError(name)
	}
	return v, nil
}

// MustAttachment panics when the value is absent or invalid.
func (c *Context) MustAttachment(name string) *discordgo.MessageAttachment {
	v, e := c.RequireAttachment(name)
	if e != nil {
		panic(e)
	}
	return v
}

// RequireMentionable returns an error when the value is absent or invalid.
func (c *Context) RequireMentionable(name string) (Mentionable, error) {
	v, ok := c.Mentionable(name)
	if !ok {
		return Mentionable{}, optionError(name)
	}
	return v, nil
}

// MustMentionable panics when the value is absent or invalid.
func (c *Context) MustMentionable(name string) Mentionable {
	v, e := c.RequireMentionable(name)
	if e != nil {
		panic(e)
	}
	return v
}

// RequireParam returns an error when the value is absent or invalid.
func (c *Context) RequireParam(name string) (string, error) {
	v, ok := c.Param(name)
	if !ok {
		return "", optionError(name)
	}
	return v, nil
}

// MustParam panics when the value is absent or invalid.
func (c *Context) MustParam(name string) string {
	v, e := c.RequireParam(name)
	if e != nil {
		panic(e)
	}
	return v
}

type responseState uint8

const (
	responsePending responseState = iota
	responseSent
	responseDeferred
)
