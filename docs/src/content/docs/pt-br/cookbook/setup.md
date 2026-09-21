---
title: "01 · Prepare e conecte o bot"
description: "Do Developer Portal ao primeiro checkpoint executável."
---

**Objetivo:** executar `/ping` em um servidor de teste. Ainda não vamos publicar um painel ou gravar chamados.

## 1. Prepare a aplicação Discord

Abra o [Developer Portal](https://discord.com/developers/applications), crie uma aplicação e obtenha o token na seção **Bot**. Guarde-o fora do Git. Em **Installation**, escolha instalação em servidor e os escopos `bot` e `applications.commands`. Para este exemplo, dê ao bot **View Channels**, **Send Messages**, **Embed Links** e **Attach Files** no canal de teste. Não conceda Administrador.

Instale a aplicação no servidor. Ative o modo de desenvolvedor no cliente Discord e copie o ID do servidor. Os nomes e a localização das telas do portal podem variar; os valores que precisamos são o token do bot e o ID do servidor.

Deixe **Interactions Endpoint URL** vazio: receberemos eventos pelo Gateway, usando uma conexão persistente do discordgo. Só usamos `IntentsGuilds`. Não precisamos de Message Content, Server Members ou Presence Intent para slash commands e componentes. Publicar um painel exige que o **usuário** tenha Gerenciar Mensagens; isso é diferente das permissões concedidas ao **bot**.

## 2. Obtenha o projeto

Instale Go **1.26 ou superior**, Git e um editor. O módulo do repositório declara o mínimo necessário; use `go version` para conferir.

```bash
git clone https://github.com/freitaseric/discordkit.git
cd discordkit/examples/community-bot
go version
go test ./...
```

Não execute `go mod init` nesta pasta: ela já pertence ao módulo na raiz do repositório. Go encontra o `go.mod` subindo pelos diretórios. Os imports `.../examples/community-bot/internal/...` são os pacotes locais deste exemplo.

## 3. Configure o ambiente

No Bash/Zsh, leia o token sem mostrá-lo na tela nem colocá-lo literalmente no histórico:

```bash
read -rs -p 'Token do bot: ' DISCORD_TOKEN; echo
export DISCORD_TOKEN
export DISCORD_GUILD_ID='ID_DO_SEU_SERVIDOR'
export BOT_DATA_FILE='data/community.json'
export COOKBOOK_STAGE=1
go run ./cmd/bot
```

No PowerShell 7:

```powershell
$env:DISCORD_TOKEN = Read-Host 'Token do bot' -MaskInput
$env:DISCORD_GUILD_ID = 'ID_DO_SEU_SERVIDOR'
$env:BOT_DATA_FILE = 'data/community.json'
$env:COOKBOOK_STAGE = '1'
go run ./cmd/bot
```

`.env.example` documenta as variáveis; **o programa não carrega `.env` automaticamente**. Um arquivo não vira variável de ambiente sem um carregador. O caminho relativo de dados é resolvido a partir do diretório em que você iniciou o processo. Use um caminho absoluto na hospedagem.

## 4. Confira o checkpoint

O terminal deve mostrar `ready` com `stage=1`. No servidor configurado, execute `/ping`: a resposta `Pong! DiscordKit está conectado.` deve aparecer somente para você. Encerre com Ctrl+C. O arquivo JSON só é criado quando houver a primeira gravação de chamado.

Se não houver comando, confirme o ID do servidor, os escopos de instalação e o erro de sincronização no terminal. Se aparecer “application did not respond”, confira se o processo ainda está rodando. Não inicie duas cópias para tentar resolver: elas disputam eventos e o arquivo local.

**Exercício:** mude a resposta do handler `ping` em `internal/bot/bot.go`, reinicie e confirme a alteração. Não precisa mudar o builder: o texto da resposta não faz parte da definição registrada no Discord.

Próximo: [como o processo e os pacotes se conectam](/pt-br/cookbook/architecture/).

---

[← Anterior: O projeto: uma central de atendimento](/pt-br/cookbook/support-bot/) · [Próximo →: 02 · Organize a aplicação Go](/pt-br/cookbook/architecture/)
