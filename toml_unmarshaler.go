package config

import (
	"bytes"
	"errors"
	"fmt"

	tomlv2 "github.com/pelletier/go-toml/v2"
)

func TomlUnmarshaler(p []byte, v interface{}) error {
	err := tomlv2.Unmarshal(p, v)
	return decodeErr(err)
}

func TomlMarshalIndent(cfg interface{}) (string, error) {
	buf := bytes.Buffer{}
	enc := tomlv2.NewEncoder(&buf)
	enc.SetIndentTables(true)
	err := enc.Encode(cfg)
	return buf.String(), err
}

func decodeErr(err error) error {
	if err == nil {
		return nil
	}
	decodeErr := &tomlv2.DecodeError{}
	if errors.As(err, &decodeErr) {
		row, column := decodeErr.Position()
		return fmt.Errorf("decode error, key=%v row=%v col=%v err=%w\n--------------------------------\n%v\n--------------------------------",
			decodeErr.Key(), row, column, decodeErr, decodeErr.String())
	}
	return err
}
