package models

type Keyer interface {
	Key() string
}

type Message struct {
	UID string `json:"UID"`
	Message string `json:"Message"`
}

func (m Message) Key() string {
	return "message:" + m.UID
}