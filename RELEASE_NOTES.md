# DiscordKit v0.1.0

A primeira versão do DiscordKit — uma camada ergonômica construída sobre o `discordgo`, sem envolver seus tipos centrais.

## Destaques

- **Router** — dispatch unificado de interações (comandos, componentes, modais e autocomplete) em um único ponto de entrada.
- **Context** — `*Context` tipado com acessores de opções (`RequireStringOption`, `RequireUserOption`, etc.) e ciclo de vida de resposta com máquina de estados (evita double-ack).
- **Command DSL** — builders tipados para chat commands, user/message context menus, opções, subcomandos e grupos, com validação das regras do Discord (limite de 25 opções, obrigatórias antes das opcionais, nomes únicos).
- **Components V2** — DSL completa (`Text`, `Row`, `Button`, `Select`, `Section`, `Thumbnail`, `Gallery`, `File`, `Separator`, `Container`) com normalização e validação da árvore de componentes.
- **Custom ID routing** — construção e parsing seguros de custom IDs com parâmetros nomeados (ex.: `/jobs/:jobID/feedback`).
- **Modal / Forms** — builders de modais com campos envolvidos em `Label`, `TextInput`, `FileUpload` (com extensão `FileTypes`) e `FormData` para leitura tipada das submissões.
- **Command Sync** — sincronização declarativa de comandos via `DiffCommands`/`SyncCommands` (Create/Update/Unchanged/Delete).
- **Middleware** — `Recovery`, `Logging`, `RequireGuild`, `RequirePermissions`.
- **Interop com discordgo** — natureza aditiva com escape hatches (`Raw()`) para os tipos nativos.

## Requisitos

- Go 1.26+
- `discordgo` (master)

## Cobertura de testes

Todos os testes passam, incluindo com o race detector (`go test -race ./...`) e `go vet ./...` limpo.
