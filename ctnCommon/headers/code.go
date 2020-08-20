package headers

import (
	"bytes"
	"encoding/gob"
)

type mem struct {
	Usage    float64 `json:"usage"`
	MaxUsage float64 `json:"max_usage"`
	Limite   float64 `json:"limit"`
}

func Encode(data interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Decode(data []byte, to interface{}) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(to)
}
