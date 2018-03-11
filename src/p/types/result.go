package types

import (
	"github.com/json-iterator/go"
)

type resp struct {
	Code string `json:"code,omitempty"`
	Msg  string `json:"msg,omitempty"`
	Data string `json:"data,omitempty"`
}

func ResultOk() []byte {

	d, _ := jsoniter.MarshalToString(resp{
		Code: "ok",
	})
	return []byte(d + "\n")
}

func ResultError(msg string) []byte {
	d, _ := jsoniter.MarshalToString(resp{
		Code: "error",
		Msg:  msg,
	})
	return []byte(d + "\n")
}
