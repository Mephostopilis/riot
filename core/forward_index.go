package core

import (
	"encoding/json"
	"fmt"

	"github.com/go-ego/riot/store"
)

type DocAttri struct {
	fields  interface{}
	docs    bool
	content string
	attri   interface{}
}

type ForwardIndex struct {
	db store.Store
}

func NewForwardIndex(db store.Store) (index *ForwardIndex) {
	index = new(ForwardIndex)
	index.db = db
	return
}

// AddDoc add doc
// 给某个文档添加评分字段
func (index *ForwardIndex) Add(docID string, fields interface{}, content ...interface{}) (err error) {
	attri := &DocAttri{
		fields:  fields,
		docs:    true,
		content: content[0].(string),
		attri:   content[1],
	}
	bytes, err := json.Marshal(attri)
	if err != nil {
		return
	}

	bkey := []byte(fmt.Sprintf("docattri:%s", docID))
	index.db.Set(bkey, bytes)
	return
}

// RemoveDoc 删除某个文档的评分字段
func (index *ForwardIndex) Del(docID string) {
	bkey := []byte(fmt.Sprintf("docattri:%s", docID))
	index.db.Delete(bkey)
}

func (index *ForwardIndex) Has(docID string) (ret bool, err error) {
	bkey := []byte(fmt.Sprintf("docattri:%s", docID))
	ret, err = index.db.Has(bkey)
	return
}
