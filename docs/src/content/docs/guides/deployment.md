---
title: "Hosting the bot"
description: "Run the Gateway connection as a long-lived process."
---

The documentation runs on Vercel; the Gateway bot needs a continuously running process. Use a VM or a container host that supports persistent connections. Build on the target platform, from your bot project:

```bash
go mod tidy
go build -o bot .
./bot
```

Export the same environment variables used locally. Closing the terminal or suspending the machine can stop the bot. On Linux, a systemd unit can supervise it:

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

Create the unprivileged `discordbot` user, copy the binary to the path above, and create the environment file with your token and guild ID. Keep that file readable only by root (`chmod 600`); systemd reads it before switching users. Then run:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now discordkit-bot
sudo journalctl -u discordkit-bot -f
```

Send SIGTERM for graceful shutdown. Start with one instance. For global distribution, choose an explicit global sync scope and remove test-guild duplicates deliberately; changing scope does not delete the old commands.
