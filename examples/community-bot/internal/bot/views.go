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
