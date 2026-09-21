---
title: "05 · Build the JSON store"
description: "Concurrency, writes, idempotency and authorization without a database server."
---

**Goal:** test storage before connecting the form. Keep stage 3 and run `go test -race ./internal/store`.

Tickets persist IDs, guild, owner, subject, description, status, assignee, rating and UTC timestamps. Do not persist sessions, interaction tokens, Context objects or cached Discord user pointers.

`Actor` is derived from the interaction, not from user-controlled custom ID parameters. Visibility requires the same guild and either ownership or staff access. Staff means Manage Messages or Administrator in this example.

| Operation | Owner | Same-guild staff | Other members/guilds |
| --- | --- | --- | --- |
| Read/list/export | Own records | Guild records | Denied |
| Claim | Only if also staff | Open and unassigned or already assigned to self | Denied |
| Close | Allowed | Allowed | Denied |
| Rate | Own closed ticket, 1–5 | Only if also owner | Denied |

## Make changes under one lock

`RWMutex` permits concurrent reads and serializes writes. `Change` reads, authorizes and mutates under the same lock to prevent two staff members claiming one ticket simultaneously.

The modal interaction ID is the creation idempotency key. Replaying the same interaction returns the existing ticket; submitting a new form creates a new ticket. Repeated close operations are harmless; ratings can be updated by the owner.

## Save before changing memory

`commit` copies the map, serializes it to a temporary file in the same directory, syncs, closes and renames it. Only then does it replace memory. Failed writes leave the in-memory state unchanged.

On local POSIX filesystems, rename avoids exposing a partially written destination file. This is not a full database transaction: there is no inter-process lock or directory fsync, and absolute power-loss durability is not promised. Windows and network filesystems do not offer the same guarantees. Host this example as a single Linux process.

Missing files mean a new database. Corrupt or unsupported files fail startup instead of silently creating an empty database.

<!-- cookbook-source: internal/store/store.go -->
```go title="internal/store/store.go"
// Package store is an example-specific JSON repository, not a DiscordKit API.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var (
	ErrNotFound  = errors.New("ticket not found or inaccessible")
	ErrForbidden = errors.New("action not allowed")
	ErrClosed    = errors.New("ticket is closed")
	ErrInvalid   = errors.New("invalid ticket data")
)

type Actor struct {
	GuildID, UserID string
	Staff           bool
}
type Ticket struct {
	ID          string    `json:"id"`
	GuildID     string    `json:"guild_id"`
	OwnerID     string    `json:"owner_id"`
	Subject     string    `json:"subject"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	AssigneeID  string    `json:"assignee_id,omitempty"`
	Rating      int       `json:"rating,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
type database struct {
	Version int               `json:"version"`
	Tickets map[string]Ticket `json:"tickets"`
}
type Store struct {
	mu   sync.RWMutex
	path string
	db   database
}

func Open(path string) (*Store, error) {
	s := &Store{path: path, db: database{Version: 1, Tickets: map[string]Ticket{}}}
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &s.db); err != nil {
		return nil, fmt.Errorf("read database: %w", err)
	}
	if s.db.Version != 1 || s.db.Tickets == nil {
		return nil, fmt.Errorf("unsupported or incomplete database")
	}
	for id, t := range s.db.Tickets {
		if id == "" || t.ID != id || t.GuildID == "" || t.OwnerID == "" || (t.Status != "open" && t.Status != "closed") || t.Rating < 0 || t.Rating > 5 {
			return nil, fmt.Errorf("invalid stored ticket %q", id)
		}
	}
	return s, nil
}

func visible(a Actor, t Ticket) bool {
	return a.GuildID != "" && a.UserID != "" && a.GuildID == t.GuildID && (a.Staff || a.UserID == t.OwnerID)
}
func (s *Store) Get(a Actor, id string) (Ticket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.db.Tickets[id]
	if !ok || !visible(a, t) {
		return Ticket{}, ErrNotFound
	}
	return t, nil
}
func (s *Store) List(a Actor, status string) []Ticket {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Ticket{}
	for _, t := range s.db.Tickets {
		if visible(a, t) && (status == "" || t.Status == status) {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

// Create uses the modal interaction ID as an idempotency key.
func (s *Store) Create(a Actor, id, subject, description string) (Ticket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	subject, description = strings.TrimSpace(subject), strings.TrimSpace(description)
	if a.GuildID == "" || a.UserID == "" || id == "" || utf8.RuneCountInString(subject) < 3 || utf8.RuneCountInString(subject) > 80 || utf8.RuneCountInString(description) < 10 || utf8.RuneCountInString(description) > 1000 {
		return Ticket{}, ErrInvalid
	}
	if t, ok := s.db.Tickets[id]; ok {
		if t.GuildID != a.GuildID || t.OwnerID != a.UserID {
			return Ticket{}, ErrNotFound
		}
		return t, nil
	}
	now := time.Now().UTC()
	t := Ticket{ID: id, GuildID: a.GuildID, OwnerID: a.UserID, Subject: subject, Description: description, Status: "open", CreatedAt: now, UpdatedAt: now}
	return t, s.commit(t)
}

// Change authorizes and changes a ticket while holding the same lock.
func (s *Store) Change(a Actor, id, action string, rating int) (Ticket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.db.Tickets[id]
	if !ok || !visible(a, t) {
		return Ticket{}, ErrNotFound
	}
	switch action {
	case "claim":
		if !a.Staff {
			return Ticket{}, ErrForbidden
		}
		if t.Status != "open" {
			return Ticket{}, ErrClosed
		}
		if t.AssigneeID != "" && t.AssigneeID != a.UserID {
			return Ticket{}, ErrForbidden
		}
		t.AssigneeID = a.UserID
	case "close":
		if t.Status == "closed" {
			return t, nil
		}
		t.Status = "closed"
	case "rate":
		if a.UserID != t.OwnerID || t.Status != "closed" || rating < 1 || rating > 5 {
			return Ticket{}, ErrForbidden
		}
		t.Rating = rating
	default:
		return Ticket{}, ErrInvalid
	}
	t.UpdatedAt = time.Now().UTC()
	return t, s.commit(t)
}

// Write a copy first. A failed save must not change the in-memory database.
// One process only: the mutex is not an inter-process or distributed lock.
func (s *Store) commit(t Ticket) error {
	next := database{Version: 1, Tickets: make(map[string]Ticket, len(s.db.Tickets)+1)}
	for id, v := range s.db.Tickets {
		next.Tickets[id] = v
	}
	next.Tickets[t.ID] = t
	raw, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(s.path)
	if err = os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".community-*.tmp")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.Write(raw); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(name, s.path); err != nil {
		return err
	}
	s.db = next
	return nil
}
```
<!-- /cookbook-source -->

## Verify and extend

```bash
go test -race ./internal/store
```

Tests cover reopening, authorization, concurrent writes, replayed interactions, corrupt files and failed saves without memory changes.

**Exercise:** add a priority field with three allowed values. Decide how old records get a default before changing the schema version.

Every write copies and rewrites the whole database; reads scan records and there is no automatic retention policy. This is suitable for a small teaching deployment. Move to SQLite/PostgreSQL when you need scale, multiple processes or stronger guarantees. A mutex does not coordinate containers.

Do not edit the file while the bot runs. Stop before backup or restore; Operations provides the procedure.

Next: [connect forms and tickets](/cookbook/tickets/).

---

[← Previous: 04 · Panel, components and permissions](/cookbook/components/) · [Next →: 06 · Open and track tickets](/cookbook/tickets/)
