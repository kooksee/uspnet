package types

import (
	"github.com/json-iterator/go"
)

type resp struct {
	Code string `json:"event,omitempty"`
	Msg  string `json:"msg,omitempty"`
	Data string `json:"data,omitempty"`
}

func ResultOk() []byte {

	d, _ := jsoniter.Marshal(resp{
		Code: "ok",
	})
	return d
}

func ResultError(msg string) []byte {
	d, _ := jsoniter.Marshal(resp{
		Code: "error",
		Msg:  msg,
	})
	return d
}
