package config

import "os"

type ServerConfig struct {
	Version    string // 1.25.1
	MaxPlayers uint64
}

var (
	FAVICON_PATH = os.Getenv("FAVICON_PATH")
)
