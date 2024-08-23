package main

import (
	"log"
	"os"

	"github.com/pelletier/go-toml"
	"proj.dichay.tech/dpb-proxy/modules"
	"proj.dichay.tech/dpb-proxy/modules/discord"
	"proj.dichay.tech/dpb-proxy/server"
)

func main() {
	cfg := readConfig()

	mds := []modules.Module{
		&discord.DiscordModule{},
	}

	srv := server.New(&cfg, mds)

	srv.Listen()
}

func readConfig() server.ServerConfig {
	c := server.ServerConfig{}
	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		f, err := os.Create("config.toml")
		if err != nil {
			log.Fatalf("create config: %v", err)
		}

		c.Servers = append(c.Servers, server.RemoteConfig{
			Name:    "основной",
			Address: ":8080",
		})

		data, err := toml.Marshal(c)
		if err != nil {
			log.Fatalf("encode default config: %v", err)
		}

		if _, err := f.Write(data); err != nil {
			log.Fatalf("write default config: %v", err)
		}
		_ = f.Close()
	}
	data, err := os.ReadFile("config.toml")
	if err != nil {
		log.Fatalf("read config: %v", err)
	}
	if err := toml.Unmarshal(data, &c); err != nil {
		log.Fatalf("decode config: %v", err)
	}
	data, _ = toml.Marshal(c)
	if err := os.WriteFile("config.toml", data, 0644); err != nil {
		log.Fatalf("write config: %v", err)
	}
	return c
}
