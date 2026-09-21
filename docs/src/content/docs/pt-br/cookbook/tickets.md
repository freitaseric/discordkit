---
title: "06 · Abra e acompanhe chamados"
description: "Forms, rotas parametrizadas, defer e mudanças de estado."
---

**Ative a etapa 4.** Agora `/ticket new` abre um formulário e `/ticket show` consulta um registro. Publique um novo `/panel` para incluir o botão de abertura.

## Uma interação abre; outra envia

`newTicket` constrói um `Form` com dois `Field` e `TextInput`. `Build` valida o layout. `ShowModal` é a resposta inicial ao comando ou botão: não faça `Defer` antes dela. O envio do formulário chega depois, como uma nova interação, na rota `/tickets/create`.

O modal usa Labels (`Field`) para os campos, de acordo com o modelo atual. `Paragraph`, `MinLen` e `MaxLen` controlam a experiência de preenchimento. O store ainda valida os limites e remove espaços nas extremidades.

`c.Form().RequireString` retorna erro se o campo não existir. Depois de extrair os dados, `createTicket` reconhece a interação com `Defer(true)` **antes da gravação em disco**. A resposta final vem por `Edit`, não por outro `Reply`.

O Discord exige uma resposta inicial em até 3 segundos; tokens de interação duram 15 minutos. Autocomplete e abertura de modal não têm o mesmo fluxo de defer de mensagens. Veja o [ciclo oficial de respostas](https://docs.discord.com/developers/interactions/receiving-and-responding).

## Roteie IDs, não permissões

`Route("/tickets/:id/close").Param(...).Build()` monta o custom ID, com escape e validação. O handler usa `RequireParam("id")`, que recebe o parâmetro decodificado. IDs têm limite de 100 caracteres; não coloque descrição, token ou dados pessoais neles.

O botão contém somente a referência. `Store.Change` verifica servidor, autor/equipe e estado atual **a cada clique**. Mesmo que alguém tente reutilizar um ID, não ganha acesso ao chamado de outra pessoa.

Usamos `DeferUpdate` nos botões de assumir/encerrar porque queremos atualizar a mensagem privada que contém o botão. O painel público nunca é substituído pelo texto de um chamado. Após persistir, `Edit` redesenha a resposta com o estado atual. Se outra pessoa já tiver assumido, a operação é recusada.

<!-- cookbook-source: internal/bot/tickets.go -->
```go title="internal/bot/tickets.go"
package bot

import (
	"strings"

	dk "github.com/freitaseric/discordkit"
)

func (b *Bot) registerTickets(r *dk.Router) func() error {
	return func() error {
		group := r.Group("ticket")
		registrations := []func() error{
			func() error { return group.Command("new", b.newTicket) },
			func() error { return group.Command("show", b.showTicket) },
			func() error { return group.Autocomplete("show", "id", b.completeTicket) },
			func() error { return r.Component("/tickets/new", b.newTicket) },
			func() error { return r.Modal("/tickets/create", b.createTicket) },
			func() error { return r.Component("/tickets/:id/close", b.changeTicket("close")) },
			func() error { return r.Component("/tickets/:id/claim", b.changeTicket("claim")) },
		}
		for _, register := range registrations {
			if err := register(); err != nil {
				return err
			}
		}
		return nil
	}
}
func (b *Bot) newTicket(c *dk.Context) error {
	form, err := dk.Form("/tickets/create", "Novo chamado",
		dk.Field("Assunto", dk.TextInput("subject").Required(true).MinLen(3).MaxLen(80)),
		dk.Field("Descrição", dk.TextInput("description").Paragraph().Required(true).MinLen(10).MaxLen(1000)).Description("Explique como reproduzir. Não inclua senhas."),
	).Build()
	if err != nil {
		return err
	}
	return c.ShowModal(form)
}
func (b *Bot) createTicket(c *dk.Context) error {
	subject, err := c.Form().RequireString("subject")
	if err != nil {
		return err
	}
	description, err := c.Form().RequireString("description")
	if err != nil {
		return err
	}
	if err = c.Defer(true); err != nil {
		return err
	}
	t, err := b.Store.Create(actor(c), c.Interaction.ID, subject, description)
	if err != nil {
		return finish(c, dk.MessageSpec{}, err)
	}
	return finish(c, ticketView(t), nil)
}
func (b *Bot) showTicket(c *dk.Context) error {
	id, err := c.RequireString("id")
	if err != nil {
		return err
	}
	if err = c.Defer(true); err != nil {
		return err
	}
	t, err := b.Store.Get(actor(c), id)
	if err != nil {
		return finish(c, dk.MessageSpec{}, err)
	}
	return finish(c, ticketView(t), nil)
}
func (b *Bot) changeTicket(action string) dk.Handler {
	return func(c *dk.Context) error {
		id, err := c.RequireParam("id")
		if err != nil {
			return err
		}
		// These controls only occur in private ticket responses, never on the public panel.
		if err = c.DeferUpdate(); err != nil {
			return err
		}
		t, err := b.Store.Change(actor(c), id, action, 0)
		if err != nil {
			return finish(c, dk.MessageSpec{}, err)
		}
		return finish(c, ticketView(t), nil)
	}
}
func (b *Bot) completeTicket(c *dk.Context) error {
	focused, ok := c.FocusedOption()
	if !ok {
		return c.Autocomplete()
	}
	query, _ := focused.Value.(string)
	query = strings.ToLower(query)
	choices := []dk.ChoiceValue{}
	for _, t := range b.Store.List(actor(c), "") {
		if strings.Contains(strings.ToLower(t.Subject), query) || strings.Contains(t.ID, query) {
			// IDs are short Discord snowflakes; truncate the title to stay below 100 characters.
			title := []rune(t.Subject)
			for len(string(title)) > 60 {
				title = title[:len(title)-1]
			}
			choices = append(choices, dk.Choice(string(title)+" · "+t.ID, t.ID))
			if len(choices) == 25 {
				break
			}
		}
	}
	return c.Autocomplete(choices...)
}
```
<!-- /cookbook-source -->

## O que acontece quando algo falha

Se o disco falhar antes da gravação, `finish` edita a resposta de espera com uma mensagem genérica e registra o erro no servidor. Se a gravação funcionar mas a resposta Discord falhar, o chamado **já existe**. O usuário deve consultar `/ticket show`/`list` antes de repetir o formulário. Não há transação distribuída entre um arquivo e a API do Discord.

As telas antigas são fotografias do estado. Encerrar um chamado em outra mensagem não atualiza todas as cópias já enviadas. Porém um novo clique consulta o store e respeita o estado atual.

## Confira com autor, equipe e terceiro

1. Abra um chamado e anote o ID. Verifique `data/community.json`.
2. Consulte pelo ID como autor. Como terceiro sem permissão, o ID não deve aparecer no autocomplete e a consulta manual deve ser recusada.
3. Como equipe, consulte o ID e clique **Assumir**. Outro atendente não deve sobrescrever essa atribuição.
4. Encerre como autor ou equipe. O estado passa para `closed`.
5. Pare e reinicie o bot. Consulte o mesmo ID. O chamado deve continuar encerrado.

**Exercício:** acrescente um botão “Atualizar” que apenas consulta e redesenha o chamado. Reaplique a autorização mesmo sendo uma leitura.

Próximo: [fila e navegação](/pt-br/cookbook/queue/).

---

[← Anterior: 05 · Construa o JSON-db](/pt-br/cookbook/persistence/) · [Próximo →: 07 · Filtros, páginas e autocomplete](/pt-br/cookbook/queue/)
