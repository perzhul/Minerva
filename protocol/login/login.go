package login

import (
	"bufio"
	"bytes"
	"io"

	"github.com/google/uuid"
	"github.com/perzhul/Minerva/protocol"
)

type LoginPacket struct {
	Name string
	UUID uuid.UUID
}

func ParseLoginPacket(data []byte) (*LoginPacket, error) {
	r := bufio.NewReader(bytes.NewReader(data))
	_, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	name, err := protocol.String(r)
	if err != nil {
		return nil, err
	}

	uuidData, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	uuid, err := uuid.FromBytes(uuidData)
	if err != nil {
		return nil, err
	}

	return &LoginPacket{
		Name: name,
		UUID: uuid,
	}, nil
}
