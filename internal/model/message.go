package model

type Message struct {
	ID      string `json:"id"`
	Topic   string `json:"topic"`
	Payload string `json:"payload"`
	TS      string `json:"ts"`
}
