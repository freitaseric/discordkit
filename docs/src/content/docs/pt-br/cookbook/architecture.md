---
title: "02 · Organize a aplicação Go"
description: "Pacotes, dependências, configuração, Gateway e encerramento."
---

**Checkpoint:** mantenha `COOKBOOK_STAGE=1`. Neste capítulo você entende o esqueleto que continuará igual enquanto adicionamos funcionalidades.

## Uma responsabilidade por pacote

| Caminho | Responsabilidade | Não deve fazer |
| --- | --- | --- |
| `cmd/bot/main.go` | Montar dependências e controlar o processo | Implementar regras de chamados |
| `internal/config` | Ler e validar ambiente | Conectar ao Discord |
| `internal/store` | Validar, autorizar e persistir chamados | Conhecer `Context` ou tokens |
| `internal/bot` | Traduzir interações em operações e mensagens | Escrever diretamente no JSON |

`internal` é uma restrição de importação do Go: pacotes fora da árvore pai não podem importar essas implementações. Não é uma barreira de segurança de dados. `cmd/bot` é uma convenção para o executável. O nome do diretório não cria nenhum comportamento automático.

Os arquivos de `internal/bot` compartilham o mesmo pacote. `commands.go`, `tickets.go` e `queue.go` são uma divisão para leitura, não serviços separados. `Bot` recebe `*store.Store`: nenhuma variável global é necessária. Quando uma segunda implementação de armazenamento existir, introduza uma interface pequena no pacote consumidor, com as operações realmente usadas.

## Configuração explícita

`Load` retorna `(Config, error)`. Token ausente ou etapa inválida são erros de inicialização: é melhor encerrar do que abrir uma sessão parcialmente configurada. Não imprima `Config` inteiro, pois ele contém o token.

<!-- cookbook-source: internal/config/config.go -->
```go title="internal/config/config.go"
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Token, GuildID, DataFile string
	Stage                    int
}

func Load() (Config, error) {
	c := Config{Token: strings.TrimSpace(os.Getenv("DISCORD_TOKEN")), GuildID: strings.TrimSpace(os.Getenv("DISCORD_GUILD_ID")), DataFile: os.Getenv("BOT_DATA_FILE"), Stage: 6}
	if c.Token == "" || c.GuildID == "" {
		return c, fmt.Errorf("set DISCORD_TOKEN and DISCORD_GUILD_ID")
	}
	if c.DataFile == "" {
		c.DataFile = "data/community.json"
	}
	if v := os.Getenv("COOKBOOK_STAGE"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 6 {
			return c, fmt.Errorf("COOKBOOK_STAGE must be 1..6")
		}
		c.Stage = n
	}
	return c, nil
}
```
<!-- /cookbook-source -->

## Monte as dependências antes de abrir o Gateway

Leia o `run` nesta ordem: configuração → store → bot → builders → rotas → sessão → handlers → conexão → sincronização → espera por sinal. Validar comandos e rotas antes de abrir a conexão revela erros locais sem precisar receber uma interação.

`defer session.Close()` executa quando `run` retorna. `signal.NotifyContext` transforma Ctrl+C/SIGTERM em cancelamento. Isso encerra a conexão; o exemplo não implementa um protocolo completo de drenagem de handlers em andamento. Pare em períodos sem atendimento antes de um backup consistente.

O callback `Ready` pode acontecer após reconexões. Por isso ele só atualiza a presença. Criar painéis ou sincronizar comandos dentro dele causaria efeitos repetidos. `UpdateGameStatus` é discordgo puro: não há razão para DiscordKit duplicar essa API.

<!-- cookbook-source: cmd/bot/main.go -->
```go title="cmd/bot/main.go"
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	dk "github.com/freitaseric/discordkit"
	"github.com/freitaseric/discordkit/examples/community-bot/internal/bot"
	"github.com/freitaseric/discordkit/examples/community-bot/internal/config"
	"github.com/freitaseric/discordkit/examples/community-bot/internal/store"
)

func main() {
	if err := run(); err != nil {
		slog.Error("bot stopped", "error", err)
		os.Exit(1)
	}
}
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	db, err := store.Open(cfg.DataFile)
	if err != nil {
		return err
	}
	app := &bot.Bot{Store: db, GuildID: cfg.GuildID, Stage: cfg.Stage}
	commands, err := app.Commands()
	if err != nil {
		return err
	}
	router, err := app.Router()
	if err != nil {
		return err
	}
	session, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return err
	}
	session.Identify.Intents = discordgo.IntentsGuilds
	session.AddHandler(router.Handle)
	// Ready can run again after reconnects. Do not create panels or sync here.
	session.AddHandler(func(s *discordgo.Session, _ *discordgo.Ready) {
		if err := s.UpdateGameStatus(0, "/help · Central de atendimento"); err != nil {
			slog.Warn("presence failed", "error", err)
		}
	})
	if err = session.Open(); err != nil {
		return err
	}
	defer session.Close()
	// Sync is non-destructive: unrelated guild commands are not deleted.
	plan, err := dk.SyncCommands(session, session.State.User.ID, commands, dk.SyncOptions{GuildID: cfg.GuildID})
	if err != nil {
		return err
	}
	slog.Info("ready", "stage", cfg.Stage, "created", len(plan.Create), "updated", len(plan.Update), "unchanged", len(plan.Unchanged))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	return nil
}
```
<!-- /cookbook-source -->

## Roteador e middleware

`Router()` registra handlers uma única vez. O middleware recebe `next` e devolve outro handler. A ordem global é logging → escopo de servidor/contexto → RequireGuild → handler. Recovery já é aplicado pelo Router e transforma panics em erros; não devolve uma resposta sozinho.

`scope` restringe o bot ao servidor configurado e associa um prazo às requisições REST feitas pelo Context. Esse prazo não cancela `os.WriteFile` nem substitui o prazo de resposta inicial do Discord. O store faz operações locais curtas.

Erros antes da resposta tentam produzir uma mensagem privada. Erros depois de um defer são tratados por `finish`, que edita a resposta já reconhecida. Não usamos “tentar Reply, depois Followup sempre”: uma falha de rede pode deixar incerto se a primeira resposta chegou.

<!-- cookbook-source: internal/bot/bot.go -->
```go title="internal/bot/bot.go"
package bot

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
	dk "github.com/freitaseric/discordkit"
	"github.com/freitaseric/discordkit/examples/community-bot/internal/store"
)

type Bot struct {
	Store   *store.Store
	GuildID string
	Stage   int
}

func actor(c *dk.Context) store.Actor {
	a := store.Actor{GuildID: c.GuildID()}
	if u := c.User(); u != nil {
		a.UserID = u.ID
	}
	if m := c.Member(); m != nil {
		a.Staff = m.Permissions&(discordgo.PermissionManageMessages|discordgo.PermissionAdministrator) != 0
	}
	return a
}
func (b *Bot) scope(next dk.Handler) dk.Handler {
	return func(c *dk.Context) error {
		if c.GuildID() != b.GuildID {
			return c.EphemeralText("Este bot atende apenas o servidor configurado.")
		}
		ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
		defer cancel()
		c.SetContext(ctx)
		return next(c)
	}
}
func friendly(err error) string {
	switch {
	case errors.Is(err, store.ErrNotFound):
		return "Chamado não encontrado ou sem acesso."
	case errors.Is(err, store.ErrForbidden):
		return "Você não pode executar esta ação."
	case errors.Is(err, store.ErrClosed):
		return "Este chamado já foi encerrado."
	case errors.Is(err, store.ErrInvalid):
		return "Confira os campos e tente novamente."
	default:
		return "Não foi possível concluir. Consulte o chamado antes de repetir a ação."
	}
}

// finish handles failures after Defer/DeferUpdate without a second initial reply.
func finish(c *dk.Context, spec dk.MessageSpec, err error) error {
	if err != nil {
		slog.Error("ticket operation failed", "error", err)
		spec = message(dk.Text(friendly(err)))
	}
	_, sendErr := c.Edit(spec)
	return sendErr
}
func (b *Bot) Router() (*dk.Router, error) {
	r := dk.NewRouter(dk.Logging(nil), b.scope, dk.RequireGuild())
	r.OnError(func(c *dk.Context, err error) {
		slog.Error("interaction failed", "error", err)
		if c.Interaction.Type == discordgo.InteractionApplicationCommandAutocomplete {
			_ = c.Autocomplete()
			return
		}
		// ErrAlreadyResponded means an initial response was attempted, not necessarily delivered.
		// Do not blindly create a followup after a modal or an uncertain network failure.
		if e := c.EphemeralText(friendly(err)); e != nil && !errors.Is(e, dk.ErrAlreadyResponded) {
			slog.Error("error response failed", "error", e)
		}
	})
	registrations := []func() error{
		func() error {
			return r.Command("ping", func(c *dk.Context) error { return c.EphemeralText("Pong! DiscordKit está conectado.") })
		},
	}
	if b.Stage >= 2 {
		registrations = append(registrations, func() error { return r.Command("about", b.about) })
	}
	if b.Stage >= 3 {
		registrations = append(registrations,
			func() error {
				return r.Command("help", func(c *dk.Context) error { return c.Ephemeral(panel(b.Stage)) })
			},
			func() error {
				return r.Command("panel", func(c *dk.Context) error { return c.Reply(panel(b.Stage)) }, dk.RequirePermissions(discordgo.PermissionManageMessages))
			},
			func() error { return r.Component("/help/topic", b.topic) },
		)
	}
	if b.Stage >= 4 {
		registrations = append(registrations, b.registerTickets(r))
	}
	if b.Stage >= 5 {
		registrations = append(registrations, b.registerQueue(r))
	}
	if b.Stage >= 6 {
		registrations = append(registrations, b.registerOperations(r), func() error { return b.registerLab(r) })
	}
	for _, register := range registrations {
		if err := register(); err != nil {
			return nil, err
		}
	}
	return r, nil
}
```
<!-- /cookbook-source -->

**Confira:** rode `go test ./...`; passe `COOKBOOK_STAGE=0` e veja a inicialização falhar com uma mensagem clara. Volte para `1`. Reiniciar não deve publicar mensagens no canal.

**Exercício:** acrescente uma variável `BOT_LOG_LEVEL` em `config` e configure `slog` em `main`. Evite ler `os.Getenv` dentro de handlers: configuração espalhada torna testes e execução difíceis de reproduzir.

Próximo: [construir e registrar comandos](/pt-br/cookbook/commands/).

---

[← Anterior: 01 · Prepare e conecte o bot](/pt-br/cookbook/setup/) · [Próximo →: 03 · Comandos e opções](/pt-br/cookbook/commands/)
