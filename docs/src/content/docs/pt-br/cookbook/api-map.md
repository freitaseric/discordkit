---
title: "Mapa da API no cookbook"
description: "Cobertura, variantes e limites da API usada no projeto."
---

Este mapa organiza a API pública por famílias. **Uso no bot, experimento e variante explicada não são a mesma coisa.** O caminho principal implementa o atendimento; o laboratório demonstra recursos complementares; variantes cosméticas ou dependentes de monetização são explicadas sem simular funcionalidades que não existem.

| API | Capítulo e aplicação |
| --- | --- |
| Command, Options, Build, MustBuild, BuildCommands | [Definições e validação; MustBuild panica em caso de erro.](/pt-br/cookbook/commands/) |
| SubCommand, SubCommandGroup | [Ticket agrupa ações; lab options inspect mostra o segundo nível.](/pt-br/cookbook/laboratory/) |
| UserCommand, MessageCommand | [Menus de contexto; alvo em TargetID/resolved do discordgo.](/pt-br/cookbook/laboratory/) |
| NameLocalizations, DescriptionLocalizations, Choice.Localized | [Lab localiza o nome. Descrições e escolhas seguem mapas Locale; o caminho do router permanece canônico.](/pt-br/cookbook/laboratory/) |
| DefaultMemberPermissions, Contexts, IntegrationTypes, NSFW | [Permissão padrão de panel é usada. Contexts/IntegrationTypes são configurações de distribuição (sobretudo global); NSFW não é necessário neste bot. Não copie indiscriminadamente para sync em guild.](/pt-br/cookbook/components/) |
| StringOption, IntegerOption, NumberOption, BooleanOption | [ID, nota, peso e visibilidade; Required, limites, Choices ou Autocomplete.](/pt-br/cookbook/commands/) |
| UserOption, RoleOption, ChannelOption, MentionableOption, AttachmentOption | [Objetos resolvidos; ChannelTypes limita a seleção.](/pt-br/cookbook/laboratory/) |
| Choice, Choices, Autocomplete | [Status fixos e busca dinâmica; escolhas e autocomplete são mutuamente exclusivos na mesma opção.](/pt-br/cookbook/queue/) |
| NewRouter, Use, Group, Command, Component, Modal, Autocomplete | [Registro por feature. Use acrescenta middleware global; Group compartilha prefixo e middleware.](/pt-br/cookbook/architecture/) |
| Handle, Dispatch, NewContext | [Handle adapta o Gateway; Dispatch/NewContext permitem testar com eventos e transporte falso.](/pt-br/cookbook/operations/) |
| OnError, Handler, Middleware, Recovery, Logging, PanicError | [Erros e composição; Recovery é automático no Router. Stack de PanicError pertence ao log protegido, nunca à resposta pública.](/pt-br/cookbook/architecture/) |
| RequireGuild, RequirePermissions | [Servidor e permissão do membro; não verificam permissões do bot no canal.](/pt-br/cookbook/components/) |
| Context, SetContext, Session, Interaction | [Prazo REST e escape hatches nativos; não misture respostas raw e gerenciadas na mesma interação.](/pt-br/cookbook/architecture/) |
| User, Member, GuildID, ChannelID, Locale, IsGuild, IsDM | [Metadados do evento. Este bot recusa DM e qualquer guild fora da configuração.](/pt-br/cookbook/architecture/) |
| Option, FocusedOption, String, Int, Float, Bool | [Presença de opção é separada do valor zero; FocusedOption identifica a busca em edição.](/pt-br/cookbook/queue/) |
| UserOption, MemberOption, Role, Channel, Attachment, Mentionable | [Getters do Context leem resolved; Mentionable.Valid exige exatamente usuário ou cargo.](/pt-br/cookbook/laboratory/) |
| Require… e Must… / Require… and Must… | [Require retorna erro; Must panica. Use Require em entradas externas. Há variantes para string, int, float, bool, entidades e parâmetros.](/pt-br/cookbook/tickets/) |
| CustomID.Validate, Route, Param, Build, MustBuild | [IDs estáveis, codificação de parâmetros e limite de 100 caracteres; não representam autorização.](/pt-br/cookbook/tickets/) |
| Context.Param, RequireParam, MustParam | [Leitura do parâmetro decodificado pelo roteador.](/pt-br/cookbook/tickets/) |
| MessageSpec, InteractionResponseData, WebhookParams, WebhookEdit, MessageEdit | [Um payload, conversões para cada endpoint; File/Attachments/AllowedMentions são campos explícitos.](/pt-br/cookbook/laboratory/) |
| Reply, ReplyText, Ephemeral, EphemeralText | [Primeira resposta e visibilidade; texto é conveniência sobre MessageSpec.](/pt-br/cookbook/commands/) |
| Defer, Edit, DeferUpdate, Update | [Defer antes do disco; Update para navegação imediata; Edit depois do reconhecimento.](/pt-br/cookbook/tickets/) |
| Followup, FollowupEdit, DeleteResponse | [Exportação adicional, revisão da mensagem e ocultação pelo botão.](/pt-br/cookbook/operations/) |
| Context.Autocomplete, ShowModal | [Respostas iniciais específicas; não vêm depois de um defer de mensagem.](/pt-br/cookbook/tickets/) |
| Components, Raw, Text, Row, Container, Separator | [Árvore V2; Container.AccentColor, Spoiler, Separator.Divider/Large alteram apresentação.](/pt-br/cookbook/components/) |
| Button, LinkButton, PremiumButton | [Ações roteadas e links no bot. PremiumButton exige SKU real/monetização; não tem custom ID nem handler de compra fictício.](/pt-br/cookbook/components/) |
| Primary, Secondary, Success, Danger, Emoji, Disabled | [Variantes de botão: estilo não concede permissão. Disabled evita cliques normais, mas o servidor revalida o estado.](/pt-br/cookbook/components/) |
| SelectOption, StringSelect | [Opções, Description, Emoji, Default e Placeholder; valide values no handler.](/pt-br/cookbook/queue/) |
| UserSelect, RoleSelect, MentionableSelect, ChannelSelect | [Seleção de entidades, limites, DefaultValues e ChannelTypes; o laboratório demonstra três e propõe a variante mentionable.](/pt-br/cookbook/laboratory/) |
| Section, Thumbnail, Gallery, MediaItem, File | [Mídia V2, descrições, spoiler e attachment://; seção exige acessório.](/pt-br/cookbook/laboratory/) |
| Form, Fields, Field, TextInput, Modal, ModalOf | [Builder e conversão raw de modal; Value/Placeholder/Paragraph/Required/MinLen/MaxLen configuram inputs.](/pt-br/cookbook/tickets/) |
| FileUpload, FileTypes, MinValues, MaxValues, Required | [Upload no formulário. FileTypes limita a interface; validar tipo/tamanho no servidor seria necessário antes de processar um arquivo.](/pt-br/cookbook/laboratory/) |
| Context.Form, FormData.String, Strings, Files, RequireString, MustString, RequireStrings, RequireFiles | [Texto e valores resolvidos do envio do modal; métodos Require detectam ausência.](/pt-br/cookbook/tickets/) |
| ValidateMessageComponents, ValidateModalComponents | [Validação local, também executada pelos builders/conversores. Não substitui a validação remota.](/pt-br/cookbook/components/) |
| SyncCommands, SyncOptions, DiffCommands, SyncPlan.Empty | [Sync sem exclusão por padrão; prévia via lista remota + DiffCommands. A opção real de remoção se chama Delete.](/pt-br/cookbook/laboratory/) |

## Erros são parte do contrato

Use `errors.Is` com os sentinelas de `errors.go` para classificar erros, e `errors.As` para `PanicError`. Não compare o texto de uma mensagem. `ErrAlreadyResponded` indica que o Context já iniciou uma resposta; não prova que o Discord a recebeu. `ErrNotResponded` indica que você tentou uma operação posterior sem reconhecimento inicial. Erros de validação e rota devem falhar nos testes ou no bootstrap quando forem defeitos de configuração.

## Limites deliberados

Persistência, migrações, filas, permissões de domínio e quotas pertencem à aplicação. O cookbook não adiciona esses serviços ao pacote raiz. Recursos do Discord sem suporte no discordgo fixado pelo `go.mod` não ganham suporte só por estarem embrulhados em `Raw`. Consulte a referência da biblioteca e as notas do upstream antes de experimentar componentes novos.

[Voltar ao roteiro](/pt-br/cookbook/support-bot/).
