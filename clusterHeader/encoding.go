package header

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"unsafe"
)

func Str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func Bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
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

func Decode(bytestream []byte, data interface{}) error {
	buf := bytes.NewBuffer(bytestream)
	dec := gob.NewDecoder(buf)
	return dec.Decode(data)
}

func JsonString(i interface{}) string {
	s, _ := json.Marshal(i)
	return string(s)
}

func JsonByteArray(i interface{}) []byte {
	s, _ := json.Marshal(i)
	return s
}
