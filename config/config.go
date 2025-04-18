package config

import "os"

var (
	FAVICON_PATH = os.Getenv("FAVICON_PATH")
)
