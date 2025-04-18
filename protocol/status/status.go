package status

import "go.minekube.com/common/minecraft/component"

const StatusResponsePacketID byte = 0x00

type Response struct {
	Version            Version      `json:"version"`
	Players            *Players     `json:"players,omitempty"`
	Description        *Description `json:"description,omitempty"`
	Favicon            string       `json:"favicon,omitempty"`
	EnforcesSecureChat bool         `json:"enforcesSecureChat"`
}

type Version struct {
	Name     string `json:"name"`
	Protocol []byte `json:"protocol"`
}

type Players struct {
	Max    uint64 `json:"max"`
	Online uint64 `json:"online"`
	Sample Sample `json:"sample"`
}

type Sample struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type Description struct {
	Text TextComponent `json:"text"`
}

type TextComponent component.Text
