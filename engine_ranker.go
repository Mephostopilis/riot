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

type rankerAddDocReq struct {
	docId  string
	fields interface{}
	// new
	content string
	// new 属性
	attri interface{}
}

type rankerRankReq struct {
	docs             []types.IndexedDoc
	options          types.RankOpts
	rankerReturnChan chan rankerReturnReq
	countDocsOnly    bool
}

type rankerReturnReq struct {
	// docs    types.ScoredDocs
	docs    interface{}
	numDocs int
}

type rankerRemoveDocReq struct {
	docId string
}

// Ranker initialize the ranker channel
func (engine *Engine) initRanker(options *types.EngineOpts) {
	// 初始化排序器
	engine.rankers = make([]*core.Ranker, options.NumShards)
	for shard := 0; shard < options.NumShards; shard++ {
		ranker, _ := core.NewRanker()
		engine.rankers[shard] = ranker
	}

	// 初始消息通道
	engine.rankerRankChans = make([]chan rankerRankReq, options.NumShards)
	for shard := 0; shard < options.NumShards; shard++ {
		engine.rankerRankChans[shard] = make(chan rankerRankReq, options.RankerBufLen)
	}
}

// 排序出来
func (engine *Engine) rankerRank(shard int) {
	for {
		select {
		case request := <-engine.rankerRankChans[shard]:
			if request.options.MaxOutputs != 0 {
				request.options.MaxOutputs += request.options.OutputOffset
			}
			request.options.OutputOffset = 0
			outputDocs, numDocs := engine.rankers[shard].Rank(request.docs, request.options, request.countDocsOnly)
			request.rankerReturnChan <- rankerReturnReq{docs: outputDocs, numDocs: numDocs}
		}
	}
}
