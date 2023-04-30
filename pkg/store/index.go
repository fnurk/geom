package store

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fnurk/geom/pkg/model"
)

type IndexType string

const (
	INMEM   = "inmem"
	PERSIST = "persist"
)

type Index struct {
	indexType IndexType
	field     reflect.StructField
}

var indexes = make(map[reflect.Type][]Index)

func CheckIndexes() {
	for _, t := range model.Types {
		val := reflect.ValueOf(t)
		for i := 0; i < val.Type().NumField(); i++ {
			t := val.Type().Field(i)

			indexTag := t.Tag.Get("index")
			if indexTag != "" {
				parts := strings.Split(indexTag, ",")
				for _, part := range parts {
					idx := Index{
						field: t,
					}

					switch part {
					case INMEM:
						idx.indexType = INMEM
						indexes[val.Type()] = append(indexes[val.Type()], idx)
					case PERSIST:
						idx.indexType = PERSIST
						indexes[val.Type()] = append(indexes[val.Type()], idx)
					}

				}
			}
		}
	}
	fmt.Printf("INDEXES: %+v\n", indexes)
}

func AddIndex(typeName string, field string, indexName string) {
	indexesgeneric := MemGet("indexes", typeName)
	var indexes []*model.Index
	if indexesgeneric == nil {
		indexes = make([]*model.Index, 0)
	} else {
		indexes = indexesgeneric.([]*model.Index)
	}
	indexes = append(indexes,
		&model.Index{
			TypeName:  typeName,
			FieldName: field,
			IndexName: indexName,
		})

	MemSet("indexes", typeName, indexes)
}

func GetIndexes(typeName string) []*model.Index {
	indexptr := MemGet("indexes", typeName)
	var indexes []*model.Index
	if indexptr != nil {
		indexes = indexptr.([]*model.Index)
	}
	if indexptr == nil {
		newMap := make([]*model.Index, 0)
		MemSet("indexes", typeName, newMap)
		indexes = newMap
	}
	return indexes
}
