package observability

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
)

func Decode(r io.Reader) ([]FluentBitMessage, error) {
	var result []FluentBitMessage
	var msg interface{}

	err := json.NewDecoder(r).Decode(&msg)
	if err != nil {
		return result, err
	}
	switch m1 := msg.(type) {
	case []interface{}:
		if len(m1) == 0 {
			return result, fmt.Errorf("empty array")
		}
		for _, m1Element := range m1 {
			switch m2 := m1Element.(type) {
			case map[string]interface{}:
				var fluentBitMessage FluentBitMessage
				fluentBitMessage.Data = make(map[string]string)
				val, ok := m2["date"]
				if ok {
					fluentBitMessage.Date = val.(float64)
				}
				for key, value := range m2 {
					if key != "date" {
						switch valueTyped := value.(type) {
						case string:
							fluentBitMessage.Data[key] = valueTyped
						case float64:
							fluentBitMessage.Data[key] = strconv.FormatFloat(valueTyped, 'f', -1, 64)
						case []byte:
							fluentBitMessage.Data[key] = string(valueTyped)
						default:
							fmt.Printf("no hit on type: %s", reflect.TypeOf(valueTyped))
						}
					}
				}
				result = append(result, fluentBitMessage)
			default:
				return result, fmt.Errorf("invalid type: no map found in array")
			}
		}
	default:
		return result, fmt.Errorf("invalid type: no array found")
	}
	return result, nil
}

func decodeMessages(msgs []byte) []FluentBitMessage {
	res := []FluentBitMessage{}
	recordOffset := 0
	for k := 0; k < len(msgs); k++ {
		if k > recordOffset+8 && msgs[k] == 0xff && msgs[k-1] == 0xff {
			res = append(res, decodeMessage(msgs[recordOffset:k]))
			recordOffset = k + 1
		}
	}
	return res
}
func decodeMessage(data []byte) FluentBitMessage {
	bits := binary.LittleEndian.Uint64(data[0:8])
	msg := FluentBitMessage{
		Date: math.Float64frombits(bits),
		Data: map[string]string{},
	}
	isKey := false
	key := ""
	start := 8
	for kk := start; kk < len(data); kk++ {
		if data[kk] == 0xff {
			if isKey {
				isKey = false
				msg.Data[key] = string(data[start+1 : kk])
				start = kk + 1
			} else {
				isKey = true
				key = string(data[start:kk])
				start = kk
			}
		}
	}
	// if last record didn't end with the terminator
	if data[len(data)-1] != 0xff {
		msg.Data[key] = string(data[start+1:])
	}
	return msg
}

func scanMessage(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	for i := 0; i < len(data); i++ {
		if data[i] == 0xff && data[i-1] == 0xff {
			return i + 1, data[0 : i-1], nil
		}
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		if len(data) > 1 && data[len(data)-1] == 0xff && data[len(data)-2] == 0xff {
			return len(data[0 : len(data)-2]), data, nil
		} else {
			return len(data), data, nil
		}
	}
	// Request more data.
	return 0, nil, nil
}
