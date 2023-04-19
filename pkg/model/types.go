package model

import (
	"encoding/json"
	"time"
)

var Types = map[string]DataType{
	"note": {
		Template: Note{},
	},
	"thing": {
		Template: Thing{},
	},
}

type Document interface {
	Note | Thing
}

type MetaFields struct {
	CreatedBy    string    `json:"createdBy"`
	Created      time.Time `json:"created"`
	LastModified time.Time `json:"lastModified"`
	SharedWith   []string  `json:"sharedWith"`
}

type Note struct {
	MetaFields
	Body string `json:"body" validate:"required`
}

type Thing struct {
	MetaFields
	Id string `json:"id" validate:"required"` //to be indexed
}

// GENERIC CODE BELOW

type Index struct {
	TypeName  string
	FieldName string
	IndexName string
}

func Decode(t string, data []byte) (interface{}, error) {
	decoded := Types[t].Template
	err := json.Unmarshal(data, &decoded)

	if err != nil {
		return nil, err
	}
	return decoded, nil
}

type DataType struct {
	Template interface{}
}
