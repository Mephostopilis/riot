// Copyright 2017 ego authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package riot

import (
	"bytes"
	"log"

	"encoding/gob"

	"github.com/go-ego/murmur"
	"github.com/go-ego/riot/types"
)

// func (engine *Engine) IsDocExist(docId uint64) bool {
// 	return core.IsDocExist(docId)
// }

// HasDoc if the document is exist return true
func (engine *Engine) HasDoc(docId string) bool {
	for shard := 0; shard < engine.initOptions.NumShards; shard++ {
		// engine.indexers = append(engine.indexers, core.Indexer{})

		has := engine.indexers[shard].HasDoc(docId)

		if has {
			return true
		}
	}

	return false
}

// HasDocDB if the document is exist in the database
// return true
func (engine *Engine) HasDocDB(docId string) bool {
	shard := murmur.Sum32(docId) % uint32(engine.initOptions.StoreShards)

	has, err := engine.dbs[shard].Has([]byte(docId))
	if err != nil {
		log.Println("engine.dbs[shard].Has(b[0:length]): ", err)
	}

	return has
}

// GetDBAllIds get all the DocId from the storage database
// and return
// 从数据库遍历所有的 DocId, 并返回
func (engine *Engine) GetDBAllIds() []string {
	docsId := make([]string, 0)

	for i := range engine.dbs {
		engine.dbs[i].ForEach(func(k, v []byte) error {
			// fmt.Println(k, v)
			docsId = append(docsId, string(k))
			return nil
		})
	}

	return docsId
}

// GetDBAllDocs get the db all docs
func (engine *Engine) GetDBAllDocs() (docsId []string, docsData []types.DocData) {
	for i := range engine.dbs {
		engine.dbs[i].ForEach(func(key, val []byte) error {
			// fmt.Println(k, v)
			docsId = append(docsId, string(key))

			buf := bytes.NewReader(val)
			dec := gob.NewDecoder(buf)

			var data types.DocData
			err := dec.Decode(&data)
			if err != nil {
				log.Println("dec.decode: ", err)
			}

			docsData = append(docsData, data)

			return nil
		})
	}

	return docsId, docsData
}

// GetAllDocIds get all the DocId from the storage database
// and return
// 从数据库遍历所有的 DocId, 并返回
func (engine *Engine) GetAllDocIds() []string {
	return engine.GetDBAllIds()
}

// Try handler(err)
func Try(fun func(), handler func(interface{})) {
	defer func() {
		if err := recover(); err != nil {
			handler(err)
		}
	}()
	fun()
}

// SearchDoc find the document that satisfies the search criteria.
// This function is thread safe, return not IDonly
func (engine *Engine) SearchDoc(request types.SearchReq) (output types.SearchDoc) {
	resp := engine.Search(request)
	return types.SearchDoc{
		BaseResp: resp.BaseResp,
		Docs:     resp.Docs.(types.ScoredDocs),
	}
}

// SearchID find the document that satisfies the search criteria.
// This function is thread safe, return IDonly
func (engine *Engine) SearchID(request types.SearchReq) (output types.SearchID) {
	// return types.SearchID(engine.Search(request))
	resp := engine.Search(request)
	return types.SearchID{
		BaseResp: resp.BaseResp,
		Docs:     resp.Docs.(types.ScoredIDs),
	}
}

// Search find the document that satisfies the search criteria.
// This function is thread safe
// 查找满足搜索条件的文档，此函数线程安全
func (engine *Engine) Search(request types.SearchReq) (output types.SearchResp) {

	tokens := engine.Tokens(request)

	var rankOpts types.RankOpts
	if request.RankOpts == nil {
		rankOpts = *engine.initOptions.DefRankOpts
	} else {
		rankOpts = *request.RankOpts
	}

	if rankOpts.ScoringCriteria == nil {
		rankOpts.ScoringCriteria = engine.initOptions.DefRankOpts.ScoringCriteria
	}

	// 建立排序器返回的通信通道
	rankerReturnChan := make(
		chan rankerReturnReq, engine.initOptions.NumShards)

	// 生成查找请求
	lookupRequest := indexerLookupReq{
		countDocsOnly:    request.CountDocsOnly,
		tokens:           tokens,
		labels:           request.Labels,
		docIds:           request.DocIds,
		options:          rankOpts,
		rankerReturnChan: rankerReturnChan,
		orderless:        request.Orderless,
		logic:            request.Logic,
	}

	// 向索引器发送查找请求
	for shard := 0; shard < engine.initOptions.NumShards; shard++ {
		engine.indexerLookupChans[shard] <- lookupRequest
	}

	if engine.initOptions.IDOnly {
		output = engine.RankID(request, rankOpts, tokens, rankerReturnChan)
		return
	}

	output = engine.Ranks(request, rankOpts, tokens, rankerReturnChan)
	return
}
