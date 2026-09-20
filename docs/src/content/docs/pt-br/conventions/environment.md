---
title: "Variáveis de ambiente"
description: "Leia a configuração uma vez e detecte valores ausentes antes de conectar."
---

O cookbook usa duas variáveis:

| Variável | Valor |
| --- | --- |
| `DISCORD_TOKEN` | Token do bot, obtido no Developer Portal |
| `DISCORD_GUILD_ID` | ID do servidor de teste; não é ID de canal nem de aplicação |

```go
 token := strings.TrimSpace(os.Getenv("DISCORD_TOKEN"))
 if token == "" {
     return fmt.Errorf("DISCORD_TOKEN is required")
 }
```

Importe `strings`, `os` e `fmt`. `os.Getenv` lê o ambiente do processo; ele **não** carrega arquivos `.env`. O [tutorial do primeiro bot](/pt-br/getting-started/first-bot/) mostra como carregá-los no Bash e no PowerShell.

Versione um `.env.example` com valores vazios e ignore `.env` e `.env.*`, exceto `.env.example`. Na hospedagem, configure segredos no gerenciador de processos ou na plataforma. Nunca registre tokens em logs. Redefina um token exposto no Developer Portal.
