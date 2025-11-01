package types

type DemoEvent struct {
	MessageID string `json:"messageId"`
	Type      string `json:"type"`
	Version   int    `json:"version"`
	Payload   any    `json:"payload"`
}
