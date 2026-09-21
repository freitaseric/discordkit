---
title: "03 · Comandos e opções"
description: "Separe o contrato registrado no Discord do handler que responde."
---

**Ative:** `export COOKBOOK_STAGE=2` e `go run ./cmd/bot` (no PowerShell, `$env:COOKBOOK_STAGE='2'`). Agora `/about` se junta a `/ping`.

## Definir não é executar

`Command` constrói a definição que o Discord usa para exibir o comando. `router.Command` registra o código que trata a interação recebida. É preciso ter os dois. O nome canônico deve coincidir, inclusive os espaços dos subcomandos: `ticket show`, não `/ticket/show`.

`Commands()` mostra também as definições que vamos habilitar depois. Leia agora apenas `ping` e `about`. `BooleanOption("private", ...)` é opcional. `c.Bool` devolve `(valor, encontrado)`: `false` e “não informado” são situações diferentes, por isso adotamos `true` como padrão explícito no handler.

`BuildCommands` retorna um erro validável no bootstrap. Prefira a versão que retorna erro quando o conteúdo depende de configuração. `MustBuild` é conveniente para constantes já validadas, mas panica se houver erro.

<!-- cookbook-source: internal/bot/commands.go -->
```go title="internal/bot/commands.go"
package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	dk "github.com/freitaseric/discordkit"
)

func (b *Bot) Commands() ([]*discordgo.ApplicationCommand, error) {
	commands := []*dk.CommandBuilder{dk.Command("ping", "Verifique a conexão")}
	if b.Stage >= 2 {
		commands = append(commands, dk.Command("about", "Conheça o bot", dk.BooleanOption("private", "Responder apenas para você")))
	}
	if b.Stage >= 3 {
		commands = append(commands, dk.Command("help", "Abra a central de ajuda"), dk.Command("panel", "Publique a central de ajuda").DefaultMemberPermissions(discordgo.PermissionManageMessages))
	}
	if b.Stage >= 4 {
		subs := []dk.Option{
			dk.SubCommand("new", "Abra um chamado"),
			dk.SubCommand("show", "Consulte um chamado", dk.StringOption("id", "ID do chamado").Required().Autocomplete()),
		}
		if b.Stage >= 5 {
			subs = append(subs, dk.SubCommand("list", "Consulte a fila", dk.StringOption("status", "Filtro").Choices(dk.Choice("Abertos", "open"), dk.Choice("Encerrados", "closed"), dk.Choice("Todos", "all"))))
		}
		if b.Stage >= 6 {
			subs = append(subs, dk.SubCommand("rate", "Avalie um chamado encerrado", dk.StringOption("id", "ID do chamado").Required().Autocomplete(), dk.IntegerOption("score", "Nota de 1 a 5").Required().Min(1).Max(5)), dk.SubCommand("export", "Exporte os chamados que você pode acessar"))
		}
		commands = append(commands, dk.Command("ticket", "Gerencie chamados", subs...))
	}
	if b.Stage >= 6 {
		commands = append(commands, labCommands()...)
	}
	return dk.BuildCommands(commands...)
}
func (b *Bot) about(c *dk.Context) error {
	private, ok := c.Bool("private")
	if !ok {
		private = true
	}
	spec := message(dk.Text(fmt.Sprintf("## Central DiscordKit\nEtapa %d/6 · Go + discordgo + DiscordKit\nIdioma da interação: %s", b.Stage, c.Locale())))
	if private {
		return c.Ephemeral(spec)
	}
	return c.Reply(spec)
}
func (b *Bot) topic(c *dk.Context) error {
	data := c.Interaction.MessageComponentData()
	if len(data.Values) != 1 {
		return c.EphemeralText("Escolha um assunto.")
	}
	switch data.Values[0] {
	case "rules":
		return c.EphemeralText("Respeite os membros. Não envie spam nem credenciais.")
	case "support":
		return c.EphemeralText("Descreva o problema e como reproduzi-lo. Consulte /ticket list para acompanhar.")
	default:
		return c.EphemeralText("Assunto desconhecido.")
	}
}
```
<!-- /cookbook-source -->

## Resposta pública ou privada

`about` usa `c.Locale()` para mostrar o idioma recebido. `c.Ephemeral(spec)` restringe a resposta ao solicitante; `c.Reply(spec)` cria uma resposta pública. Não existe como transformar uma resposta pública já enviada em privada mudando um booleano depois.

Na etapa 3, `topic` lerá os valores de um menu pelo `Interaction.MessageComponentData()` do discordgo. Ter acesso ao objeto nativo é intencional: DiscordKit complementa, não substitui, discordgo.

## Subcomandos, escolhas e autocomplete

Nas etapas seguintes, `/ticket` agrupa `new`, `show`, `list`, `rate` e `export`. Isso mantém a navegação relacionada sem criar vários comandos no topo. Os subcomandos são definidos no builder e roteados pelo caminho completo.

`Choices` serve para um conjunto fixo, como estado aberto/encerrado. `Autocomplete` consulta dados dinâmicos, como IDs de chamados. Não combine os dois na mesma opção. `IntegerOption` delimita a nota de 1 a 5, mas o store ainda valida a nota: dados externos não devem depender somente da interface do Discord.

## Sincronize sem apagar o servidor

`SyncCommands` compara definições remotas e desejadas; por padrão não remove comandos extras. Criar o handler local não sincroniza nada por conta própria. O exemplo sincroniza na inicialização, uma vez, no servidor de teste.

Se voltar de uma etapa avançada para uma anterior, comandos antigos podem continuar visíveis sem handler ativo. Avance novamente ou remova explicitamente apenas os comandos do seu experimento. Não ative `Delete: true` em uma aplicação compartilhada sem revisar um plano de exclusão. O capítulo de laboratório mostra como pré-visualizar com `DiffCommands`.

**Confira:** `/about` sem opção é privado; `/about private:false` é público. Modifique a descrição no builder, reinicie e confira a mudança na interface do Discord.

**Exercício:** adicione `/about verbose` como opção booleana opcional e use `c.Bool` para exibir informação adicional. Não registre um subcomando se a intenção é uma opção do comando atual.

Próximo: [painel e componentes](/pt-br/cookbook/components/).

---

[← Anterior: 02 · Organize a aplicação Go](/pt-br/cookbook/architecture/) · [Próximo →: 04 · Painel, componentes e permissões](/pt-br/cookbook/components/)
