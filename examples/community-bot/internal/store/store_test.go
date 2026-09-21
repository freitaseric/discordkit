package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestPersistenceAndAuthorization(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data", "db.json")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	owner := Actor{GuildID: "guild", UserID: "owner"}
	staff := Actor{GuildID: "guild", UserID: "staff", Staff: true}
	ticket, err := s.Create(owner, "123", "Ajuda", "Não consigo acessar o painel")
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range []Actor{{GuildID: "guild", UserID: "other"}, {GuildID: "other", UserID: "owner", Staff: true}} {
		if _, err = s.Get(a, ticket.ID); !errors.Is(err, ErrNotFound) {
			t.Fatal("data leaked", err)
		}
		if len(s.List(a, "")) != 0 {
			t.Fatal("list leaked")
		}
		if _, err = s.Change(a, ticket.ID, "close", 0); !errors.Is(err, ErrNotFound) {
			t.Fatal("unauthorized change", err)
		}
	}
	if _, err = s.Change(owner, ticket.ID, "claim", 0); !errors.Is(err, ErrForbidden) {
		t.Fatal(err)
	}
	if _, err = s.Change(staff, ticket.ID, "claim", 0); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Change(Actor{GuildID: "guild", UserID: "staff2", Staff: true}, ticket.ID, "claim", 0); !errors.Is(err, ErrForbidden) {
		t.Fatal("claim overwritten", err)
	}
	if _, err = s.Change(owner, ticket.ID, "rate", 5); !errors.Is(err, ErrForbidden) {
		t.Fatal(err)
	}
	if _, err = s.Change(owner, ticket.ID, "close", 0); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Change(staff, ticket.ID, "rate", 5); !errors.Is(err, ErrForbidden) {
		t.Fatal(err)
	}
	if _, err = s.Change(owner, ticket.ID, "rate", 5); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	got, err := reopened.Get(owner, ticket.ID)
	if err != nil || got.Status != "closed" || got.Rating != 5 || got.AssigneeID != "staff" {
		t.Fatalf("lost data: %+v %v", got, err)
	}
	if _, err = reopened.Create(owner, "123", "Outro assunto", "Uma descrição diferente"); err != nil {
		t.Fatal(err)
	}
	if len(reopened.List(owner, "")) != 1 {
		t.Fatal("duplicate interaction created ticket")
	}
}
func TestConcurrentWrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.json")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	a := Actor{GuildID: "guild", UserID: "user"}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := s.Create(a, fmt.Sprint(i), "Assunto", "Descrição detalhada"); err != nil {
				t.Error(err)
			}
			s.List(a, "")
		}()
	}
	wg.Wait()
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.List(a, "")) != 20 {
		t.Fatal("lost concurrent write")
	}
}
func TestCorruptDatabaseAndFailedWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "db.json")
	if err := os.WriteFile(path, []byte("{broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(path); err == nil {
		t.Fatal("silently accepted corrupt database")
	}
	path = filepath.Join(dir, "blocker", "db.json")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "blocker"), []byte("file"), 0600); err != nil {
		t.Fatal(err)
	}
	a := Actor{GuildID: "g", UserID: "u"}
	if _, err = s.Create(a, "1", "Assunto", "Descrição detalhada"); err == nil {
		t.Fatal("expected failed save")
	}
	if len(s.List(a, "")) != 0 {
		t.Fatal("failed write changed memory")
	}
}
