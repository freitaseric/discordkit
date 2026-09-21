---
title: "05 · Construa o JSON-db"
description: "Concorrência, gravação, idempotência e autorização sem um servidor de banco."
---

**Objetivo:** entender e testar o armazenamento antes de ligar o formulário. Mantenha a etapa 3 e execute `go test -race ./internal/store`.

## Modele o que precisa sobreviver ao processo

Um chamado tem ID, servidor, autor, assunto, descrição, estado, atendente, avaliação e datas UTC. Não guardamos token, Context, sessão discordgo ou ponteiros para usuários em cache. Esses objetos têm outro ciclo de vida.

`Actor` representa quem está realizando a operação. O adapter Discord monta esse valor a partir do evento; nunca o preenche usando um ID de usuário fornecido livremente em custom IDs. A condição `visible` exige o mesmo servidor e, então, autor ou equipe. O usuário da equipe é identificado por Gerenciar Mensagens/Administrador neste exemplo; em uma aplicação real você pode adotar um cargo específico.

| Operação | Autor | Equipe do mesmo servidor | Outro membro/servidor |
| --- | --- | --- | --- |
| Consultar/listar/exportar | Seus chamados | Todos do servidor | Sem acesso |
| Assumir | Não, salvo se também for equipe | Chamado aberto e livre ou já atribuído a si | Sem acesso |
| Encerrar | Sim | Sim | Sem acesso |
| Avaliar | Próprio chamado encerrado, nota 1–5 | Somente se for o autor | Sem acesso |

A política mora no store, inclusive para operações chamadas pelos botões. Esconder controles seria apenas uma conveniência visual.

## Proteja a transição inteira

`RWMutex` permite múltiplas leituras, mas serializa alterações. Em `Change`, leitura, autorização e escrita acontecem sob o mesmo lock. Se você verificasse o estado fora do lock, dois atendentes poderiam assumir o mesmo chamado ao mesmo tempo.

`Create` usa o ID da interação de envio do modal como chave idempotente. Reprocessar o mesmo evento retorna o mesmo chamado. Um novo envio de formulário cria outro ID: isso não impede que um usuário abra dois chamados parecidos. Encerrar duas vezes também é seguro; alterar uma avaliação substitui a nota anterior.

## Grave uma nova versão antes de trocar a memória

`commit` copia o mapa, serializa e grava em um arquivo temporário **no mesmo diretório**. Depois chama `Sync`, fecha e renomeia. Só então troca o estado em memória. Se a gravação falhar, o erro volta ao handler e o mapa continua com os dados anteriores.

Em sistemas POSIX e filesystem local, a troca por rename evita que um leitor encontre metade do JSON. Isso **não é uma transação de banco completa**, não oferece lock entre processos e não garante durabilidade absoluta após queda de energia: não fazemos fsync do diretório. Não conte com a mesma atomicidade em Windows ou filesystem de rede. Use uma única instância Linux para hospedar este exemplo.

Arquivo ausente significa banco novo. Arquivo inválido ou versão desconhecida causa falha de inicialização. Jamais trate JSON corrompido como banco vazio: isso apagaria silenciosamente a evidência do problema.

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

## Confira as invariantes

Os testes criam diretórios temporários e verificam reinicialização, isolamento por servidor e autor, mudanças autorizadas, escrita concorrente, repetição de eventos e falha de escrita sem alteração de memória.

```bash
go test -race ./internal/store
```

**Exercício:** adicione prioridade (`low`, `normal`, `high`) à struct e à validação. Decida o valor padrão para registros antigos antes de alterar a versão do arquivo. Criar um campo novo sem planejar leitura de dados antigos é uma mudança de schema, mesmo em JSON.

## Quando sair do JSON

Cada gravação copia e reescreve o banco inteiro; cada consulta percorre os registros, e não há limite de retenção automático. É uma escolha didática para servidores pequenos. Ao precisar de múltiplos processos, maior volume, consultas complexas ou garantias mais fortes, substitua por SQLite/PostgreSQL e preserve as operações do domínio. Um mutex não coordena dois containers.

Não edite o JSON com o bot rodando. Para backup, pare o processo e copie o arquivo; para restauração, valide uma cópia com os testes e só depois substitua o arquivo de produção. O capítulo de operação detalha o procedimento.

Próximo: [ligar o formulário aos chamados](/pt-br/cookbook/tickets/).

---

[← Anterior: 04 · Painel, componentes e permissões](/pt-br/cookbook/components/) · [Próximo →: 06 · Abra e acompanhe chamados](/pt-br/cookbook/tickets/)
