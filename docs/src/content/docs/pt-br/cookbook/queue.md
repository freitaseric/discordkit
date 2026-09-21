---
title: "07 · Filtros, páginas e autocomplete"
description: "Navegação por dados reais sem vazar títulos de chamados."
---

**Ative a etapa 5.** `/ticket list` passa a mostrar uma fila privada, com cinco registros por página.

## Derive a tela a partir dos dados

O servidor recebe o estado do filtro e o número da página pelo custom ID. Ele não confia que “Próxima” veio de uma tela legítima: valida o estado, converte a página e limita o índice à faixa atual. Se os dados mudarem entre cliques, a última página válida é recalculada.

`Store.List(actor, filtro)` aplica autorização **antes** de a lista virar texto. Um membro comum vê os seus registros; a equipe vê os do servidor. `queueView` apenas pagina o resultado já autorizado. Não envie todos os chamados para o cliente esperando que o select esconda os proibidos.

`Update` substitui a mensagem original do componente. É adequado aqui porque a consulta está em memória e a resposta é imediata. Se você trocar por uma consulta remota demorada, faça `DeferUpdate` antes e depois `Edit`.

<!-- cookbook-source: internal/bot/queue.go -->
```go title="internal/bot/queue.go"
package bot

import (
	"strconv"

	dk "github.com/freitaseric/discordkit"
)

func validStatus(s string) bool { return s == "open" || s == "closed" || s == "all" }
func (b *Bot) queue(c *dk.Context, status string, page int) dk.MessageSpec {
	filter := status
	if filter == "all" {
		filter = ""
	}
	return queueView(b.Store.List(actor(c), filter), status, page)
}
func (b *Bot) registerQueue(r *dk.Router) func() error {
	return func() error {
		if err := r.Command("ticket list", func(c *dk.Context) error {
			status, ok := c.String("status")
			if !ok {
				status = "open"
			}
			if !validStatus(status) {
				return c.EphemeralText("Filtro inválido.")
			}
			return c.Ephemeral(b.queue(c, status, 0))
		}); err != nil {
			return err
		}
		if err := r.Component("/queue/filter", func(c *dk.Context) error {
			values := c.Interaction.MessageComponentData().Values
			if len(values) != 1 || !validStatus(values[0]) {
				return c.EphemeralText("Filtro inválido.")
			}
			return c.Update(b.queue(c, values[0], 0))
		}); err != nil {
			return err
		}
		return r.Component("/queue/:status/:page", func(c *dk.Context) error {
			status, err := c.RequireParam("status")
			if err != nil {
				return err
			}
			raw, err := c.RequireParam("page")
			if err != nil {
				return err
			}
			page, err := strconv.Atoi(raw)
			if err != nil || !validStatus(status) {
				return c.EphemeralText("Página inválida.")
			}
			return c.Update(b.queue(c, status, page))
		})
	}
}
```
<!-- /cookbook-source -->

## Autocomplete também é uma superfície de dados

Releia `completeTicket` no [capítulo anterior](/pt-br/cookbook/tickets/). `FocusedOption` identifica o campo em edição. Filtramos ID/assunto entre os chamados autorizados, limitamos a lista a 25 escolhas e truncamos o título para respeitar o limite de nome.

Não responda autocomplete com uma mensagem normal e não use defer. Se não houver resultado, devolva `c.Autocomplete()` vazio. A sugestão não é uma prova de autorização: o usuário pode digitar qualquer ID, por isso `/ticket show` chama `Store.Get` novamente.

Não fazemos requisições REST por sugestão. Uma busca remota precisaria de índice, timeout e orçamento de latência apropriados; a API do Discord não espera uma consulta lenta.

## Confira estados vazios e mudança de páginas

Crie seis chamados com assuntos diferentes. A primeira página deve ter cinco; a segunda, um. “Anterior” deve ficar desativado na primeira página, e “Próxima” na última. Encerre todos, escolha “Abertos” e confira o estado vazio. Troque para “Todos” e confirme que os registros continuam presentes.

Teste também com um membro sem chamados: nenhuma tela deve mostrar dados de outras pessoas. Na conta do autor, digite parte do assunto em `/ticket show id:` e selecione a sugestão.

**Exercício:** adicione filtro por “atribuídos a mim”. Faça o filtro com o ID do ator autenticado, não com um parâmetro livre de usuário, e mantenha a restrição de servidor.

Próximo: [avaliação, exportação e operação](/pt-br/cookbook/operations/).

---

[← Anterior: 06 · Abra e acompanhe chamados](/pt-br/cookbook/tickets/) · [Próximo →: 08 · Teste e opere o bot](/pt-br/cookbook/operations/)
