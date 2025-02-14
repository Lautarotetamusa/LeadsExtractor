package whatsapp

import (
	"encoding/json"
	"fmt"
)

type MessageType int

const (
	TextMessage MessageType = iota
	InteractiveMessage
	ButtonMessage
	OrderMessage
)

var (
	messageTypeName = map[MessageType]string{
		InteractiveMessage: "interactive",
		TextMessage:        "text",
		ButtonMessage:      "button",
		OrderMessage:       "order", //TODO
	}
	messageTypeValue = map[string]MessageType{
		"text":        0,
		"interactive": 1,
		"button":      2,
		"order":       3,
	}
)

func ParseMessageType(s string) (MessageType, error) {
	value, ok := messageTypeValue[s]
	if !ok {
		return MessageType(0), fmt.Errorf("%q is not a valid message type", s)
	}
	return MessageType(value), nil
}

func (s MessageType) String() string {
	return messageTypeName[s]
}

func (s *MessageType) UnmarshalJSON(data []byte) (err error) {
	var status string
	if err := json.Unmarshal(data, &status); err != nil {
		return err
	}
	if *s, err = ParseMessageType(status); err != nil {
		return err
	}
	return nil
}

func (s MessageType) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}
