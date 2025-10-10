package tobytes

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
)

func ToBytes(value any) ([]byte, error) {
	switch v := value.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		var buf bytes.Buffer
		err := binary.Write(&buf, binary.BigEndian, v)
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	case bool:
		if v {
			return []byte{1}, nil
		}
		return []byte{0}, nil
	default:
		return json.Marshal(value)
	}
}
