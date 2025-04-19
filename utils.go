package main

import (
	"encoding/base64"
	"os"
)

func encodeImageToFavicon(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(data)

	return "data:image/png;base64," + encoded, nil
}
