package store

import "github.com/fnurk/geom/pkg/model"

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
	indexes := MemGet("indexes", typeName).([]*model.Index)
	if indexes == nil {
		indexes = make([]*model.Index, 0)
	}
	return indexes
}
