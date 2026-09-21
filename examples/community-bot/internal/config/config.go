package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Token, GuildID, DataFile string
	Stage                    int
}

func Load() (Config, error) {
	c := Config{Token: strings.TrimSpace(os.Getenv("DISCORD_TOKEN")), GuildID: strings.TrimSpace(os.Getenv("DISCORD_GUILD_ID")), DataFile: os.Getenv("BOT_DATA_FILE"), Stage: 6}
	if c.Token == "" || c.GuildID == "" {
		return c, fmt.Errorf("set DISCORD_TOKEN and DISCORD_GUILD_ID")
	}
	if c.DataFile == "" {
		c.DataFile = "data/community.json"
	}
	if v := os.Getenv("COOKBOOK_STAGE"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 6 {
			return c, fmt.Errorf("COOKBOOK_STAGE must be 1..6")
		}
		c.Stage = n
	}
	return c, nil
}
