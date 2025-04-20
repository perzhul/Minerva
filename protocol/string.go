package protocol

import (
	"bufio"

	"github.com/multiformats/go-varint"
)

func String(r *bufio.Reader) (string, error) {
	//TODO: use proper read var int compatible with
	// minecraft format
	length, err := varint.ReadUvarint(r) // it already read those bytes
	if err != nil {
		return "", err
	}

	strBytes := make([]byte, length)
	_, err = r.Read(strBytes)
	if err != nil {
		return "", err
	}

	return string(strBytes), nil
}
