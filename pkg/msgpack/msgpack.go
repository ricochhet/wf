package data

import (
	"encoding/json"

	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/fsutil"
	"github.com/tidwall/gjson"
	"github.com/vmihailenco/msgpack/v5"
)

// Encode encodes a file in the msgpack format.
func Encode(value any) ([]byte, error) {
	b, err := msgpack.Marshal(value)
	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	return b, nil
}

// Decode decodes a file in the msgpack format.
func Decode(data []byte) (any, error) {
	var result any

	err := msgpack.Unmarshal(data, &result)
	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	return result, nil
}

// EncodeFile encodes a file in the msgpack format.
func EncodeFile(name string) ([]byte, error) {
	b, err := fsutil.ReadBytes(name)
	if err != nil {
		return nil, errutil.New("fsutil.ReadBytes", err)
	}

	enc, err := Encode(gjson.ParseBytes(b).Value())
	if err != nil {
		return nil, errutil.New("Encode", err)
	}

	return enc, nil
}

// DecodeFile decodes a file in the msgpack format.
func DecodeFile(name string) (string, error) {
	b, err := fsutil.ReadBytes(name)
	if err != nil {
		return "", errutil.New("fsutil.ReadBytes", err)
	}

	dec, err := Decode(b)
	if err != nil {
		return "", errutil.New("Decode", err)
	}

	result, err := json.MarshalIndent(dec, "", " ")
	if err != nil {
		return "", errutil.New("json.MarshalIndent", err)
	}

	return string(result), nil
}
