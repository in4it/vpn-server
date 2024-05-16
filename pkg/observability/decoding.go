package observability

import (
	"encoding/json"
	"fmt"
	"io"
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
			val, ok := m2["date"]
			if ok {
				fluentBitMessage.Date = val.(float64)
			}
			fluentBitMessage.Data = m2
			result = append(result, fluentBitMessage)
		default:
			return result, fmt.Errorf("invalid type: no map found in array")
		}
	default:
		return result, fmt.Errorf("invalid type: no array found")
	}
	return result, nil
}
