package store

import (
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
	fieldName string
}

var TypeToIndexMap = make(map[string][]Index)

func PopulateIndexes() {
	for k, t := range model.Types {
		val := reflect.ValueOf(t)
		for i := 0; i < val.Type().NumField(); i++ {
			t := val.Type().Field(i)

			indexTag := t.Tag.Get("index")
			jsonTag := t.Tag.Get("json")
			if indexTag != "" {
				parts := strings.Split(indexTag, ",")
				fieldName := t.Name
				if jsonTag != "" {
					jsonParts := strings.Split(jsonTag, ",")
					fieldName = jsonParts[0]
				}
				for _, part := range parts {
					idx := Index{
						fieldName: fieldName,
					}
					switch part {
					case INMEM:
						idx.indexType = INMEM
						TypeToIndexMap[k] = append(TypeToIndexMap[k], idx)
					case PERSIST:
						idx.indexType = PERSIST
						TypeToIndexMap[k] = append(TypeToIndexMap[k], idx)
					}
				}
			}
		}
	}
}

// func AddIndex(typeName string, field string, indexName string) {
// 	indexesgeneric := MemGet("indexes", typeName)
// 	var indexes []*model.Index
// 	if indexesgeneric == nil {
// 		indexes = make([]*model.Index, 0)
// 	} else {
// 		indexes = indexesgeneric.([]*model.Index)
// 	}
// 	indexes = append(indexes,
// 		&model.Index{
// 			TypeName:  typeName,
// 			FieldName: field,
// 			IndexName: indexName,
// 		})
//
// 	MemSet("indexes", typeName, indexes)
// }
//
// func GetIndexes(typeName string) []*model.Index {
// 	indexptr := MemGet("indexes", typeName)
// 	var indexes []*model.Index
// 	if indexptr != nil {
// 		indexes = indexptr.([]*model.Index)
// 	}
// 	if indexptr == nil {
// 		newMap := make([]*model.Index, 0)
// 		MemSet("indexes", typeName, newMap)
// 		indexes = newMap
// 	}
// 	return indexes
// }
