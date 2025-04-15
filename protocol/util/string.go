package util

func String(data []byte) ([]byte, error) {
	length, err := VarInt(data)
	if err != nil {
		return nil, err
	}

	val := make([]byte, length)

	return val, nil
}

func WriteString(val string) (data []byte, err error) {
	l, err := ConvertToVarint(int32(len(val)))
	if err != nil {
		return nil, err
	}

	data = append(data, l...)

	varIntVal, err := ConvertToVarint(int32(len(val)))
	if err != nil {
		return nil, err
	}

	data = append(data, varIntVal...)

	return data, nil
}
