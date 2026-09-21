---
title: "04 · Painel, componentes e permissões"
description: "Construa uma interface V2 sem confundir visibilidade com autorização."
---

**Ative a etapa 3.** O bot ganha `/help` privado e `/panel` público. A tela é construída pela mesma função: a diferença está no método de resposta.

## Da árvore de componentes à mensagem

Leia `message` e `panel` em `views.go` primeiro. `Text` produz texto V2, `Container` agrupa visualmente, `Separator` separa seções e `Row` organiza controles. Uma row comporta até cinco botões ou um select. `LinkButton` abre uma URL e não gera uma interação para seu router.

`Components` transforma builders em componentes nativos do discordgo. `MessageSpec` concentra o payload e aplica a flag V2 quando necessário. Não misture layouts V2 com embeds ou polls legados. A validação local detecta vários erros de árvore antes da chamada REST; o Discord ainda é responsável por aceitar o payload final.

`AllowedMentions` vazio impede que texto digitado por usuários dispare menções. `safe` escapa marcações nos títulos e descrições dos chamados; não é um sanitizador universal para qualquer contexto. Os IDs interpolados pelo aplicativo são snowflakes gerados pelo Discord.

As funções `ticketView` e `queueView` deste arquivo serão usadas nos capítulos seguintes. Observe que elas recebem dados e devolvem uma mensagem: não fazem I/O, autorização ou persistência. Essa separação permite testar layouts sem conectar o bot.

<!-- cookbook-source: internal/bot/views.go -->
```go title="internal/bot/views.go"
package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	dk "github.com/freitaseric/discordkit"
	"github.com/freitaseric/discordkit/examples/community-bot/internal/store"
)

func safe(text string) string {
	return strings.NewReplacer("\\", "\\\\", "`", "\\`", "*", "\\*", "_", "\\_", "~", "\\~", ">", "\\>", "|", "\\|", "[", "\\[", "]", "\\]", "#", "\\#").Replace(text)
}
func message(nodes ...dk.Component) dk.MessageSpec {
	return dk.MessageSpec{Components: dk.Components(nodes...), AllowedMentions: &discordgo.MessageAllowedMentions{Parse: []discordgo.AllowedMentionType{}}}
}
func panel(stage int) dk.MessageSpec {
	nodes := []dk.Component{dk.Text("## Central de atendimento\nConsulte a ajuda ou abra um chamado privado."), dk.Separator(), dk.Row(dk.StringSelect("/help/topic").Placeholder("Escolha um assunto").Options(dk.SelectOption("Regras", "rules"), dk.SelectOption("Como pedir ajuda", "support"))), dk.Row(dk.LinkButton("Documentação", "https://discordkit.freitaseric.com/pt-br/"))}
	if stage >= 4 {
		nodes = append(nodes, dk.Row(dk.Button("Abrir chamado", "/tickets/new").Primary()))
	}
	return message(dk.Container(nodes...).AccentColor(0x38D9B0))
}
func ticketView(t store.Ticket) dk.MessageSpec {
	text := fmt.Sprintf("## %s\nID: `%s`\nEstado: **%s**\nAutor: `%s`\nResponsável: `%s`\nAvaliação: %d/5\n\n%s", safe(t.Subject), t.ID, t.Status, t.OwnerID, t.AssigneeID, t.Rating, safe(t.Description))
	nodes := []dk.Component{dk.Text(text)}
	if t.Status == "open" {
		closeID := string(dk.Route("/tickets/:id/close").Param("id", t.ID).MustBuild())
		claimID := string(dk.Route("/tickets/:id/claim").Param("id", t.ID).MustBuild())
		nodes = append(nodes, dk.Row(dk.Button("Assumir (equipe)", claimID).Secondary(), dk.Button("Encerrar", closeID).Danger()))
	}
	return message(dk.Container(nodes...).AccentColor(0x38D9B0))
}
func queueView(tickets []store.Ticket, status string, page int) dk.MessageSpec {
	const size = 5
	pages := (len(tickets) + size - 1) / size
	if pages == 0 {
		pages = 1
	}
	if page < 0 {
		page = 0
	}
	if page >= pages {
		page = pages - 1
	}
	text := fmt.Sprintf("## Chamados — %s\nPágina %d/%d · %d registros", status, page+1, pages, len(tickets))
	for i := page * size; i < len(tickets) && i < (page+1)*size; i++ {
		t := tickets[i]
		text += fmt.Sprintf("\n`%s` · %s · %s", t.ID, safe(t.Subject), t.Status)
	}
	if len(tickets) == 0 {
		text += "\nNenhum chamado neste filtro."
	}
	id := func(n int) string {
		return string(dk.Route("/queue/:status/:page").Param("status", status).Param("page", fmt.Sprint(n)).MustBuild())
	}
	prev, next := dk.Button("Anterior", id(page-1)).Secondary(), dk.Button("Próxima", id(page+1)).Secondary()
	if page == 0 {
		prev.Disabled()
	}
	if page+1 >= pages {
		next.Disabled()
	}
	return message(dk.Text(text), dk.Row(dk.StringSelect("/queue/filter").Options(dk.SelectOption("Abertos", "open"), dk.SelectOption("Encerrados", "closed"), dk.SelectOption("Todos", "all"))), dk.Row(prev, next))
}
```
<!-- /cookbook-source -->

## O select não escolhe uma rota diferente

O menu envia sempre `/help/topic` e uma lista de valores. `topic` valida que existe exatamente um valor e decide a resposta. O valor do menu é entrada externa, mesmo que você tenha criado as opções.

## Duas verificações de permissão

`DefaultMemberPermissions` limita a disponibilidade padrão de `/panel` na interface. `RequirePermissions` verifica o membro na execução, incluindo a exceção de Administrador. A segunda verificação é necessária porque visibilidade de comando não é uma garantia de autorização.

Não dê esse middleware globalmente ao router: isso impediria membros comuns de abrir chamados. O middleware de publicação pertence apenas à rota `/panel`.

Na etapa 4, o botão “Abrir chamado” aparece no painel. Um painel já publicado contém o layout antigo; publique um novo quando quiser atualizar os controles. Reiniciar o processo não altera mensagens já existentes. Custom IDs estáveis permitem que os controles antigos continuem roteáveis.

**Confira com duas contas:** publique o painel como alguém com Gerenciar Mensagens; clique nos tópicos como membro comum e veja respostas privadas. Sem a permissão, `/panel` deve estar indisponível por padrão ou ser recusado pelo handler.

**Exercício:** acrescente um tópico “Horários” no select e um case correspondente no handler. Clique nele e depois em um tópico antigo. O menu deve continuar público e cada resposta deve ser privada.

Próximo: [persistência e regras de acesso](/pt-br/cookbook/persistence/).

---

[← Anterior: 03 · Comandos e opções](/pt-br/cookbook/commands/) · [Próximo →: 05 · Construa o JSON-db](/pt-br/cookbook/persistence/)
