---
title: "Hospedando o bot"
description: "Execute a conexão com o Gateway em um processo contínuo."
---

A documentação roda na Vercel; o bot conectado ao Gateway precisa de um processo contínuo. Use uma VM ou hospedagem de containers que suporte conexões persistentes. Compile na plataforma de destino, dentro do projeto do bot:

```bash
go mod tidy
go build -o bot .
./bot
```

Exporte as mesmas variáveis usadas localmente. Fechar o terminal ou suspender a máquina pode parar o bot. No Linux, uma unidade systemd pode supervisioná-lo:

```ini title="/etc/systemd/system/discordkit-bot.service"
[Unit]
Description=DiscordKit support bot
After=network-online.target
Wants=network-online.target

[Service]
User=discordbot
WorkingDirectory=/opt/discordkit-bot
EnvironmentFile=/etc/discordkit-bot.env
ExecStart=/opt/discordkit-bot/bot
Restart=on-failure
RestartSec=5
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target
```

Crie o usuário sem privilégios `discordbot`, copie o binário para o caminho acima e crie o arquivo de ambiente com token e ID da guild. Deixe esse arquivo legível apenas por root (`chmod 600`); systemd o lê antes de trocar de usuário. Depois execute:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now discordkit-bot
sudo journalctl -u discordkit-bot -f
```

Use SIGTERM para encerrar de forma organizada. Comece com uma instância. Para distribuição global, escolha um escopo global explícito e remova duplicatas da guild de teste deliberadamente; mudar o escopo não apaga os comandos antigos.
