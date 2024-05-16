package observability

type FluentBitMessage struct {
	Date float64        `json:"date"`
	Data map[string]any `json:"data"`
}
