package util

import (
	"errors"
)

const (
	SEGMENT_BITS     = 0x7F
	CONTINUATION_BIT = 0x80
)

var ErrVarIntTooBig = errors.New("VarInt is too big")

func VarInt(data []byte) (int32, error) {
	var value int32
	var position int32

	for {
		b := data[0]
		value |= int32(b&SEGMENT_BITS) << position

		if b&CONTINUATION_BIT == 0 {
			break
		}

		position += 7
		if position >= 32 {
			return 0, ErrVarIntTooBig
		}
	}

	return value, nil
}

func ConvertToVarint(value int32) (data []byte, err error) {
	for {
		if (value & ^SEGMENT_BITS) == 0 {
			data = append(data, byte(value))
			break
		}

		data = append(data, byte((value&SEGMENT_BITS)|CONTINUATION_BIT))
		value >>= 7
	}

	return data, nil
}
