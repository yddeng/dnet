package json

import (
	"encoding/json"
)

/*
 json 协议解析
*/

type Json struct{}

func (_ Json) Unmarshaler(data []byte, o interface{}) (err error) {
	return json.Unmarshal(data, o)
}

func (_ Json) Marshaler(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}
