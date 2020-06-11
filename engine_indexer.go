// Copyright 2013 Hui Chen
// Copyright 2016 ego authors
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
	"github.com/go-ego/riot/core"
	"github.com/go-ego/riot/types"
)

type indexerAddDocReq struct {
	doc *types.DocIndex

	docId  string
	fields interface{}
	// new
	content string
	// new 属性
	attri interface{}
}

type indexerLookupReq struct {
	countDocsOnly bool
	tokens        []string
	labels        []string

	docIds           map[string]bool
	options          types.RankOpts
	rankerReturnChan chan rankerReturnReq
	orderless        bool
	logic            types.Logic
}

// Indexer initialize the indexer channel
func (engine *Engine) initIndexer(options *types.EngineOpts) {
	// 初始化索引器
	engine.indexers = make([]*core.Indexer, options.NumShards)
	for shard := 0; shard < options.NumShards; shard++ {
		indexer, _ := core.NewIndexer(*options.IndexerOpts)
		engine.indexers[shard] = indexer
	}

	// 初始所有管道
	engine.indexerAddDocChans = make([]chan indexerAddDocReq, options.NumShards)
	engine.indexerLookupChans = make([]chan indexerLookupReq, options.NumShards)
	for shard := 0; shard < options.NumShards; shard++ {
		engine.indexerAddDocChans[shard] = make(chan indexerAddDocReq, options.IndexerBufLen)
		engine.indexerLookupChans[shard] = make(chan indexerLookupReq, options.IndexerBufLen)
	}
}

func (engine *Engine) indexerAddDoc(shard int) {
	for {
		select {
		case request := <-engine.indexerAddDocChans[shard]:
			engine.indexers[shard].AddDocToCache(request.doc)
		}
	}
}

func (engine *Engine) orderLess(request indexerLookupReq, docs []types.IndexedDoc) {
	var outputDocs types.ScoredDocs
	for _, d := range docs {
		ids := types.ScoredID{
			DocId:            d.DocId,
			TokenSnippetLocs: d.TokenSnippetLocs,
			TokenLocs:        d.TokenLocs,
		}

		outputDocs = append(outputDocs, types.ScoredDoc{
			ScoredID: ids,
		})
	}

	request.rankerReturnChan <- rankerReturnReq{
		docs:    outputDocs,
		numDocs: len(outputDocs),
	}
}

func (engine *Engine) indexerLookup(shard int) {
	for {
		select {
		case request := <-engine.indexerLookupChans[shard]:
			docs, numDocs := engine.indexers[shard].Lookup(
				request.tokens,
				request.labels,
				request.docIds,
				request.countDocsOnly,
				request.logic)

			if request.countDocsOnly {
				request.rankerReturnChan <- rankerReturnReq{numDocs: numDocs}
				continue
			}

			if len(docs) == 0 {
				request.rankerReturnChan <- rankerReturnReq{}
				continue
			}

			if request.orderless {
				// var outputDocs interface{}
				engine.orderLess(request, docs)
				continue
			}

			rankerRequest := rankerRankReq{
				countDocsOnly:    request.countDocsOnly,
				docs:             docs,
				options:          request.options,
				rankerReturnChan: request.rankerReturnChan,
			}
			engine.rankerRankChans[shard] <- rankerRequest
		}

	}
}
