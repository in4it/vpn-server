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
		switch m2 := m1[0].(type) {
		case map[string]interface{}:
			var fluentBitMessage FluentBitMessage
			fluentBitMessage.Data = make(map[string]string)
			val, ok := m2["date"]
			if ok {
				fluentBitMessage.Date = val.(float64)
			}
			for key, value := range m2 {
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
			result = append(result, fluentBitMessage)
		default:
			return result, fmt.Errorf("invalid type: no map found in array")
		}
	default:
		return result, fmt.Errorf("invalid type: no array found")
	}
	return result, nil
}

func decodeMessage(msgs []byte) []FluentBitMessage {
	res := []FluentBitMessage{}
	recordOffset := 0
	for k := 0; k < len(msgs); k++ {
		if k > recordOffset+8 && msgs[k] == 0xff && msgs[k-1] == 0xff {
			bits := binary.LittleEndian.Uint64(msgs[recordOffset : recordOffset+8])
			msg := FluentBitMessage{
				Date: math.Float64frombits(bits),
				Data: map[string]string{},
			}
			isKey := false
			key := ""
			start := 8 + recordOffset
			for kk := 8 + recordOffset; kk < k; kk++ {
				if msgs[kk] == 0xff {
					if isKey {
						isKey = false
						msg.Data[key] = string(msgs[recordOffset+start+1 : recordOffset+kk])
						start = kk + 1
					} else {
						isKey = true
						key = string(msgs[recordOffset+start : recordOffset+kk])
						start = kk
					}
				}
			}
			res = append(res, msg)
			recordOffset = k + 1
		}
	}
	return res
}
