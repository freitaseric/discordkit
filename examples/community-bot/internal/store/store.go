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
