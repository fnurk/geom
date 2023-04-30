package model

import "encoding/json"

var Types = map[string]DataType{}

type DataType struct {
	Template interface{}
}

type Index struct {
	TypeName  string
	FieldName string
	IndexName string
}

func RegisterType(name string, template interface{}) {
	Types[name] = DataType{Template: template}
}

func Decode(t string, data []byte) (interface{}, error) {
	decoded := Types[t].Template
	err := json.Unmarshal(data, &decoded)

	if err != nil {
		return nil, err
	}
	return decoded, nil
}
