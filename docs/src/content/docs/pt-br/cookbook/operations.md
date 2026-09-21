---
title: "08 · Teste e opere o bot"
description: "Avaliação, exportação, tratamento de falhas, backup e hospedagem."
---

**Ative a etapa 6.** `/ticket rate` e `/ticket export` completam o fluxo. O laboratório também fica disponível, mas não é necessário para atender chamados.

## Avaliação e exportação

A nota é lida por `RequireInt`. O store exige que o chamado esteja encerrado e que o ator seja o autor. Um atendente não pode avaliar o próprio trabalho em nome de outra pessoa.

A exportação serializa `Store.List(actor, "")`, nunca o arquivo bruto do banco. Assim, um usuário recebe apenas seus dados; a equipe recebe os do servidor. O limite conservador de 7 MiB evita tentar enviar arquivos grandes; o limite efetivo de upload ainda depende do Discord.

`Defer(true)` reconhece a operação. Primeiro `Edit` conclui a espera; depois `Followup` envia o arquivo privado. `File("attachment://tickets.json")` referencia exatamente o nome em `Files`. `FollowupEdit` adiciona os detalhes finais, preservando os attachments recebidos. O botão de ocultar reconhece o clique com `DeferUpdate` e apaga a resposta associada ao componente com `DeleteResponse`.

Baixar uma exportação cria outra cópia dos dados. Ocultar a mensagem não revoga arquivos já baixados e não exclui chamados do store.

<!-- cookbook-source: internal/bot/operations.go -->
```go title="internal/bot/operations.go"
package bot

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/bwmarrin/discordgo"
	dk "github.com/freitaseric/discordkit"
)

func (b *Bot) registerOperations(r *dk.Router) func() error {
	return func() error {
		if err := r.Autocomplete("ticket rate", "id", b.completeTicket); err != nil {
			return err
		}
		if err := r.Command("ticket rate", func(c *dk.Context) error {
			id, err := c.RequireString("id")
			if err != nil {
				return err
			}
			score, err := c.RequireInt("score")
			if err != nil {
				return err
			}
			if err = c.Defer(true); err != nil {
				return err
			}
			t, err := b.Store.Change(actor(c), id, "rate", int(score))
			if err != nil {
				return finish(c, dk.MessageSpec{}, err)
			}
			return finish(c, ticketView(t), nil)
		}); err != nil {
			return err
		}
		if err := r.Component("/export/dismiss", func(c *dk.Context) error {
			if err := c.DeferUpdate(); err != nil {
				return err
			}
			return c.DeleteResponse()
		}); err != nil {
			return err
		}
		return r.Command("ticket export", b.export)
	}
}
func (b *Bot) export(c *dk.Context) error {
	if err := c.Defer(true); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(b.Store.List(actor(c), ""), "", "  ")
	if err != nil {
		return finish(c, dk.MessageSpec{}, err)
	}
	if len(raw) > 7*1024*1024 {
		return finish(c, message(dk.Text("Exportação muito grande. Solicite um backup ao operador.")), nil)
	}
	// Complete the deferred response before creating an additional private message.
	if err = finish(c, message(dk.Text("Exportação preparada.")), nil); err != nil {
		return err
	}
	spec := message(dk.Text("## Exportação de chamados"), dk.File("attachment://tickets.json"), dk.Row(dk.Button("Ocultar exportação", "/export/dismiss").Secondary()))
	spec.Ephemeral = true
	spec.Files = []*discordgo.File{{Name: "tickets.json", ContentType: "application/json", Reader: bytes.NewReader(raw)}}
	sent, err := c.Followup(spec)
	if err != nil {
		return err
	}
	// Edit the same followup while retaining its attachment and controls.
	spec.Components = dk.Components(dk.Text(fmt.Sprintf("## Exportação pronta\n%d bytes · somente chamados autorizados.", len(raw))), dk.File("attachment://tickets.json"), dk.Row(dk.Button("Ocultar exportação", "/export/dismiss").Secondary()))
	spec.Ephemeral = false // Editing does not change the existing message visibility.
	spec.Files = nil
	spec.Attachments = sent.Attachments
	_, err = c.FollowupEdit(sent.ID, spec)
	return err
}
```
<!-- /cookbook-source -->

## Testes sem credenciais

```bash
go test ./...
go test -race ./...
go vet ./...
```

Os testes do store cobrem políticas e persistência. Os testes de bot montam rotas de todas as etapas, validam layouts e usam um transporte HTTP falso para verificar o envio de formulário: defer privado, edição e registro em disco. Nenhum teste precisa de token real.

Isso não substitui testar no Discord. API remota, permissões de canal, instalação e comportamento do cliente dependem de uma aplicação real. Execute esta aceitação no seu servidor:

| Cenário | Resultado esperado |
| --- | --- |
| Abrir pelo painel ou `/ticket new` | Formulário, resposta privada e registro persistido |
| Terceiro consulta um ID conhecido | Recusa sem conteúdo do chamado |
| Atendente assume; outro tenta assumir | Primeiro mantém a atribuição |
| Autor encerra e avalia | Estado fechado e nota persistida |
| Terceiro/equipe tenta avaliar pelo autor | Recusa |
| Reiniciar e consultar | Mesmos registros e estados |
| Exportar como membro comum | Somente seus chamados no arquivo |
| Ocultar exportação | Mensagem some; chamado permanece |

## Extraia para seu próprio repositório

Depois de estudar, copie `cmd`, `internal`, `.gitignore` e `.env.example` para uma pasta nova. Use um nome real para seu módulo e substitua **apenas os imports locais**:

```bash
mkdir meu-bot
cp -R cmd internal .gitignore .env.example meu-bot/
cd meu-bot
go mod init example.com/meu-bot
```

Nos arquivos Go copiados, troque `github.com/freitaseric/discordkit/examples/community-bot/internal/` por `example.com/meu-bot/internal/`. Mantenha os imports da biblioteca `github.com/freitaseric/discordkit`. A partir da raiz do clone original, descubra a revisão testada com `git rev-parse HEAD`; na pasta nova, instale essa revisão usando `go get github.com/freitaseric/discordkit@REVISAO` e execute `go mod tidy`, `go test ./...` e `go run ./cmd/bot`.

Não copie `data` ou um token para o novo repositório. Fixar a revisão evita que uma alteração futura na branch mude o tutorial durante sua reconstrução.

## Hospede um processo com disco persistente

A Vercel hospeda **esta documentação**. O bot usa Gateway e requer um processo Go continuamente ativo: uma VM, container ou serviço de workers com volume persistente. Uma função HTTP efêmera não mantém essa conexão nem esse arquivo.

Compile `go build -o bot ./cmd/bot`. Defina `BOT_DATA_FILE` como um caminho absoluto em volume persistente e configure o supervisor com uma única réplica. Mantenha as variáveis no ambiente do serviço; nunca em argumentos de linha de comando com o token. Confira o [guia de hospedagem](/pt-br/guides/deployment/) para o ciclo de vida básico.

Backups: pare o serviço, copie o JSON para um local protegido com data no nome e reinicie. Para restaurar, pare, preserve uma cópia do arquivo atual, reponha o backup, mantenha o mesmo proprietário/permissões e inicie. Confirme no log que o banco abriu e consulte um ID conhecido. Um arquivo corrompido deve bloquear o início; não o substitua automaticamente por `{}`.

O exemplo não inclui antispam, retenção automática, fila de trabalhos ou drenagem completa de handlers. Antes de abrir para um servidor grande, adicione limites por usuário, política de retenção, monitoramento e um banco adequado ao volume. Não prometa armazenamento ilimitado em um JSON que é reescrito inteiro.

**Exercício final:** apague sua cópia de teste somente depois de guardar um backup, restaure e execute toda a aceitação. Você deve conseguir explicar o caminho completo: evento → router → ator → store → MessageSpec → Discord.

Próximo: [laboratório das APIs complementares](/pt-br/cookbook/laboratory/).

---

[← Anterior: 07 · Filtros, páginas e autocomplete](/pt-br/cookbook/queue/) · [Próximo →: 09 · Explore o restante da API](/pt-br/cookbook/laboratory/)
