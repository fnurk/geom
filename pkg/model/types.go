package model

import (
	"encoding/json"
	"time"
)

var Types = map[string]DataType{
	"note": {
		Template: Note{},
		IdType:   UUID,
	},
	"thing": {
		Template: Thing{},
		IdType:   AutoIncr,
	},
}

type Note struct {
	Body         string    `json:"body"`
	Created      time.Time `json:"created"`
	LastModified time.Time `json:"lastModified"`
}

type Thing struct {
	Id           string    `json:"id"`
	Created      time.Time `json:"created"`
	LastModified time.Time `json:"lastModified"`
}

// GENERIC CODE BELOW

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
	IdType   IdType
}

type IdType int64

const (
	UUID     IdType = 0
	AutoIncr        = 1
)
