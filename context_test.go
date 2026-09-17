package discordkit

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"math"
	"testing"
)

func TestContextOptions(t *testing.T) {
	r := &discordgo.ApplicationCommandInteractionDataResolved{Users: map[string]*discordgo.User{"u": {ID: "u"}}, Members: map[string]*discordgo.Member{"u": {}}, Roles: map[string]*discordgo.Role{"r": {ID: "r"}}, Channels: map[string]*discordgo.Channel{"c": {ID: "c"}}, Attachments: map[string]*discordgo.MessageAttachment{"a": {ID: "a"}}}
	options := []*discordgo.ApplicationCommandInteractionDataOption{{Name: "text", Type: 3, Value: ""}, {Name: "count", Type: 4, Value: float64(42)}, {Name: "price", Type: 10, Value: 1.5}, {Name: "private", Type: 5, Value: false}, {Name: "user", Type: 6, Value: "u"}, {Name: "role", Type: 8, Value: "r"}, {Name: "channel", Type: 7, Value: "c"}, {Name: "file", Type: 11, Value: "a"}, {Name: "mention", Type: 9, Value: "r"}}
	c := interaction(2, discordgo.ApplicationCommandInteractionData{Resolved: r, Options: []*discordgo.ApplicationCommandInteractionDataOption{{Name: "group", Type: 2, Options: []*discordgo.ApplicationCommandInteractionDataOption{{Name: "sub", Type: 1, Options: options}}}}})
	if v, ok := c.String("text"); !ok || v != "" {
		t.Fatal(v, ok)
	}
	if c.MustInt("count") != 42 || c.MustFloat("price") != 1.5 || c.MustBool("private") {
		t.Fatal("scalar options")
	}
	if c.MustUserOption("user").ID != "u" || c.MustRole("role").ID != "r" || c.MustChannel("channel").ID != "c" || c.MustAttachment("file").ID != "a" {
		t.Fatal("resolved options")
	}
	if c.MustMemberOption("user").User.ID != "u" || r.Members["u"].User != nil {
		t.Fatal("member hydration mutated original")
	}
	m := c.MustMentionable("mention")
	if m.User != nil || m.Role == nil {
		t.Fatal(m)
	}
	r.Users["r"] = &discordgo.User{ID: "r"}
	if _, ok := c.Mentionable("mention"); ok {
		t.Fatal("ambiguous mentionable")
	}
	if _, ok := c.String("count"); ok {
		t.Fatal("wrong type accepted")
	}
	if _, err := c.RequireString("missing"); !errors.Is(err, ErrMissingOption) {
		t.Fatal(err)
	}
	for _, v := range []any{1.5, math.Inf(1), math.NaN(), 1e20, "42"} {
		options[1].Value = v
		if _, ok := c.Int("count"); ok {
			t.Fatal(v)
		}
	}
	defer func() {
		if recover() == nil {
			t.Error("Must did not panic")
		}
	}()
	c.MustString("missing")
}
func TestContextIdentityAndNil(t *testing.T) {
	c := NewContext(nil, nil)
	if c.User() != nil || c.Member() != nil || c.IsDM() || c.IsGuild() {
		t.Fatal("nil identity")
	}
	if _, ok := c.UserOption("user"); ok {
		t.Fatal("nil option")
	}
	c.Interaction = &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{User: &discordgo.User{ID: "dm"}}}
	if !c.IsDM() || c.User().ID != "dm" {
		t.Fatal("DM")
	}
	c.Interaction.GuildID = "g"
	c.Interaction.Member = &discordgo.Member{User: &discordgo.User{ID: "guild"}}
	if !c.IsGuild() || c.User().ID != "guild" {
		t.Fatal("guild")
	}
}
