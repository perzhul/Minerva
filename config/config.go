package config

import (
	"os"

	"github.com/perzhul/Minerva/types"
)

type ServerConfig struct {
	Version         string // 1.25.1
	MaxPlayers      uint64
	ProtocolVersion types.Varint
}

var (
	FAVICON_PATH = os.Getenv("FAVICON_PATH")
)
