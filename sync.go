package discordkit

import (
	"reflect"

	"github.com/bwmarrin/discordgo"
)

// SyncOptions configures command synchronization scope and deletion behavior.
type SyncOptions struct {
	// GuildID scopes synchronization to a guild. Empty means global commands.
	GuildID string
	// Delete removes remote commands that are no longer declared. When false,
	// obsolete commands are reported but left in place.
	Delete bool
}

// SyncPlan is the result of diffing declared commands against remote commands.
// It is computed without any Discord API calls and is safe to inspect in tests.
type SyncPlan struct {
	Create    []*discordgo.ApplicationCommand
	Update    []*discordgo.ApplicationCommand
	Unchanged []*discordgo.ApplicationCommand
	Delete    []*discordgo.ApplicationCommand
}

// Empty reports whether the plan requires no create, update or delete actions.
func (p SyncPlan) Empty() bool {
	return len(p.Create) == 0 && len(p.Update) == 0 && len(p.Delete) == 0
}

// DiffCommands computes the synchronization plan between remote and desired
// commands, matching by name and type. When deleteObsolete is false, obsolete
// remote commands are omitted from Delete. This function performs no I/O.
func DiffCommands(remote, desired []*discordgo.ApplicationCommand, deleteObsolete bool) SyncPlan {
	var plan SyncPlan
	remoteByKey := map[string]*discordgo.ApplicationCommand{}
	for _, r := range remote {
		if r != nil {
			remoteByKey[commandKey(r)] = r
		}
	}
	desiredKeys := map[string]bool{}
	for _, d := range desired {
		if d == nil {
			continue
		}
		key := commandKey(d)
		desiredKeys[key] = true
		existing, ok := remoteByKey[key]
		switch {
		case !ok:
			plan.Create = append(plan.Create, d)
		case commandsEqual(existing, d):
			plan.Unchanged = append(plan.Unchanged, d)
		default:
			merged := *d
			merged.ID = existing.ID
			plan.Update = append(plan.Update, &merged)
		}
	}
	if deleteObsolete {
		for _, r := range remote {
			if r != nil && !desiredKeys[commandKey(r)] {
				plan.Delete = append(plan.Delete, r)
			}
		}
	}
	return plan
}

func commandKey(c *discordgo.ApplicationCommand) string {
	t := c.Type
	if t == 0 {
		t = discordgo.ChatApplicationCommand
	}
	return string(rune(t)) + ":" + c.Name
}

// commandsEqual compares the user-settable fields of two commands. Server-only
// fields such as ID, ApplicationID, GuildID and Version are ignored.
func commandsEqual(a, b *discordgo.ApplicationCommand) bool {
	if a.Name != b.Name || a.Description != b.Description {
		return false
	}
	if normalizeType(a.Type) != normalizeType(b.Type) {
		return false
	}
	if boolValue(a.NSFW) != boolValue(b.NSFW) {
		return false
	}
	if !equalInt64(a.DefaultMemberPermissions, b.DefaultMemberPermissions) {
		return false
	}
	if !reflect.DeepEqual(a.NameLocalizations, b.NameLocalizations) {
		return false
	}
	if !reflect.DeepEqual(a.DescriptionLocalizations, b.DescriptionLocalizations) {
		return false
	}
	if !reflect.DeepEqual(a.Contexts, b.Contexts) {
		return false
	}
	if !reflect.DeepEqual(a.IntegrationTypes, b.IntegrationTypes) {
		return false
	}
	return reflect.DeepEqual(a.Options, b.Options)
}

func normalizeType(t discordgo.ApplicationCommandType) discordgo.ApplicationCommandType {
	if t == 0 {
		return discordgo.ChatApplicationCommand
	}
	return t
}

func boolValue(b *bool) bool { return b != nil && *b }

func equalInt64(a, b *int64) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// SyncCommands synchronizes declared commands with Discord: creating missing
// commands, updating divergent ones, and optionally deleting obsolete ones.
// It returns the plan that was applied. Commands with no diff are left untouched.
func SyncCommands(s *discordgo.Session, appID string, desired []*discordgo.ApplicationCommand, opts SyncOptions) (SyncPlan, error) {
	remote, err := s.ApplicationCommands(appID, opts.GuildID)
	if err != nil {
		return SyncPlan{}, err
	}
	plan := DiffCommands(remote, desired, opts.Delete)
	for _, c := range plan.Create {
		if _, err := s.ApplicationCommandCreate(appID, opts.GuildID, c); err != nil {
			return plan, err
		}
	}
	for _, c := range plan.Update {
		if _, err := s.ApplicationCommandEdit(appID, opts.GuildID, c.ID, c); err != nil {
			return plan, err
		}
	}
	for _, c := range plan.Delete {
		if err := s.ApplicationCommandDelete(appID, opts.GuildID, c.ID); err != nil {
			return plan, err
		}
	}
	return plan, nil
}

// BuildCommands builds a set of command builders, returning the first error.
func BuildCommands(builders ...*CommandBuilder) ([]*discordgo.ApplicationCommand, error) {
	out := make([]*discordgo.ApplicationCommand, 0, len(builders))
	for _, b := range builders {
		cmd, err := b.Build()
		if err != nil {
			return nil, err
		}
		out = append(out, cmd)
	}
	return out, nil
}
