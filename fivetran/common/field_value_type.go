package common

import (
	"bytes"
	"encoding/json"
)

type FieldValueType int64

const (
	String FieldValueType = iota
	Integer
	Boolean
	StringList
	ObjectList
	Object

	Unknown
)

var typeMap = map[string]FieldValueType{
	"string":      String,
	"integer":     Integer,
	"boolean":     Boolean,
	"string_list": StringList,
	"object_list": ObjectList,
	"object":      Object,
}

func (lang FieldValueType) String() string {
	return [...]string{
		"string",
		"integer",
		"boolean",
		"string_list",
		"object_list",
		"object",
	}[lang]
}

func Type(typeName string) FieldValueType {
	l, ok := typeMap[typeName]
	if !ok {
		return Unknown
	}
	return l
}

func (s FieldValueType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(s.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (s *FieldValueType) UnmarshalJSON(b []byte) error {
	var j string
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}
	*s = Type(j)
	return nil
}
