package types

import "github.com/json-iterator/go"

type KMsg struct {
	Event   string `json:"event,omitempty"`
	Account string `json:"account,omitempty"`
	Msg     string `json:"msg,omitempty"`
	Token   string `json:"token,omitempty"`
}

func (t *KMsg) Dumps() []byte {
	d, _ := jsoniter.Marshal(t)
	return d
}
