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

/*

Package riot is riot engine

*/

package riot

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"sync/atomic"

	"github.com/go-ego/gse"
	"github.com/go-ego/murmur"
	"github.com/go-ego/riot/core"
	"github.com/go-ego/riot/store"
	"github.com/go-ego/riot/types"
	"github.com/go-ego/riot/utils"
)

const (

	// NumNanosecondsInAMillisecond nano-seconds in a milli-second num
	NumNanosecondsInAMillisecond = 1000000
)

// Engine initialize the engine
type Engine struct {
	loc sync.RWMutex

	// 计数器，用来统计有多少文档被索引等信息
	numDocsIndexed      uint64
	numDocsRemoved      uint64
	numDocsForceUpdated uint64

	numIndexingReqs      uint64
	numRemovingReqs      uint64
	numForceUpdatingReqs uint64
	numTokenIndexAdded   uint64
	numDocsStored        uint64

	// 记录初始化参数
	initOptions types.EngineOpts

	// 所有
	segmenter  *gse.Segmenter
	stopTokens StopTokens
	indexers   []*core.Indexer
	rankers    []*core.Ranker
	dbs        []store.Store

	segmenterChan chan segmenterReq

	// 建立索引器使用的通信通道
	indexerAddDocChans []chan indexerAddDocReq
	indexerLookupChans []chan indexerLookupReq

	// 建立排序器使用的通信通道
	rankerRankChans []chan rankerRankReq
}

func Default() (engine *Engine) {
	var (
		numShards   = 10
		segmentDict string
	)
	opts := types.EngineOpts{

		SegmenterOpts: &types.SegmenterOpts{
			// Using:         using,
			GseDict: segmentDict,
			// StopTokenFile: stopTokenFile,
		},
		NumShards: numShards,
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.DocIdsIndex,
		},
		StoreOpts: &types.StoreOpts{},
	}

	engine = New(opts)

	return
}

// New initialize the engine
func New(options types.EngineOpts) (engine *Engine) {
	// 将线程数设置为CPU数
	// runtime.GOMAXPROCS(runtime.NumCPU())
	// runtime.GOMAXPROCS(128)
	options.Init()
	engine = new(Engine)
	engine.initOptions = options

	// 初始化持久化存储通道
	engine.initStore(&options)

	// 初始化分词器
	engine.initSegment(&options)

	// 初始化索引器通道
	engine.initIndexer(&options)

	// 初始化排序器通道
	engine.initRanker(&options)

	engine.CheckMem()

	atomic.AddUint64(&engine.numDocsStored, engine.numIndexingReqs)
	return
}

func (engine *Engine) WithGse(seg *gse.Segmenter) {
	engine.segmenter = seg
}

// Startup
func (engine *Engine) Startup() {
	options := engine.initOptions
	// 启动分词器
	for iThread := 0; iThread < options.SegmenterOpts.NumGseThreads; iThread++ {
		go engine.segmenterWorker()
	}

	// 启动索引器和排序器
	for shard := 0; shard < options.NumShards; shard++ {
		go engine.indexerAddDoc(shard)

		for i := 0; i < options.NumIndexerThreads; i++ {
			go engine.indexerLookup(shard)
		}
		for i := 0; i < options.NumRankerThreads; i++ {
			go engine.rankerRank(shard)
		}
	}
}

// CheckMem check the memory when the memory is larger
// than 99.99% using the store
func (engine *Engine) CheckMem() {
	// Todo test
	// if !engine.initOptions.StoreOpts.UseStore {
	// 	log.Info("Check virtualMemory...")
	// 	vmem, _ := mem.VirtualMemory()

	// 	log.Info("Total: %v, Free: %v, UsedPercent: %f%%\n", vmem.Total, vmem.Free, vmem.UsedPercent)

	// 	useMem := fmt.Sprintf("%.2f", vmem.UsedPercent)
	// 	if useMem == "99.99" {
	// 		engine.initOptions.StoreOpts.UseStore = true
	// 		// engine.initOptions.StoreFolder = DefaultPath
	// 		// os.MkdirAll(DefaultPath, 0777)
	// 	}
	// }
}

// IndexDoc add the document to the index
// 将文档加入索引
//
// 输入参数：
//  docId	      标识文档编号，必须唯一，docId == 0 表示非法文档（用于强制刷新索引），[1, +oo) 表示合法文档
//  data	      见 DocIndexData 注释
//  forceUpdate   是否强制刷新 cache，如果设为 true，则尽快添加到索引，否则等待 cache 满之后一次全量添加
//
// 注意：
//      1. 这个函数是线程安全的，请尽可能并发调用以提高索引速度
//      2. 这个函数调用是非同步的，也就是说在函数返回时有可能文档还没有加入索引中，因此
//         如果立刻调用Search可能无法查询到这个文档。强制刷新索引请调用FlushIndex函数。
func (engine *Engine) IndexDoc(docId string, data types.DocData) (err error) {
	if docId == "0" {
		return
	}
	hash := murmur.Sum32(fmt.Sprintf("%s%s", docId, data.Content))
	engine.segmenterChan <- segmenterReq{docId: docId, hash: hash, data: data}
	bytes, err := json.Marshal(data)
	if err != nil {
		return
	}
	engine.dbs[0].Set([]byte(fmt.Sprintf("doc:%s", docId)), bytes)
	return
}

// RemoveDoc remove the document from the index
// 将文档从索引中删除
//
// 输入参数：
//  docId	      标识文档编号，必须唯一，docId == 0 表示非法文档（用于强制刷新索引），[1, +oo) 表示合法文档
//  forceUpdate 是否强制刷新 cache，如果设为 true，则尽快删除索引，否则等待 cache 满之后一次全量删除
//
// 注意：
//      1. 这个函数是线程安全的，请尽可能并发调用以提高索引速度
//      2. 这个函数调用是非同步的，也就是说在函数返回时有可能文档还没有加入索引中，因此
//         如果立刻调用 Search 可能无法查询到这个文档。强制刷新索引请调用 FlushIndex 函数。
func (engine *Engine) RemoveDoc(docId string) (err error) {
	if docId == "0" {
		return
	}
	atomic.AddUint64(&engine.numRemovingReqs, 1)
	engine.dbs[0].Delete([]byte(fmt.Sprintf("doc:%s", docId)))
	return
}

// // 获取文本的分词结果
// func (engine *Engine) Tokens(text []byte) (tokens []string) {
// 	querySegments := engine.segmenter.Segment(text)
// 	for _, s := range querySegments {
// 		token := s.Token().Text()
// 		if !engine.stopTokens.IsStopToken(token) {
// 			tokens = append(tokens, token)
// 		}
// 	}
// 	return tokens
// }

// Segment get the word segmentation result of the text
// 获取文本的分词结果, 只分词与过滤弃用词
func (engine *Engine) Segment(content string) (keywords []string) {
	var segments []string
	hmm := engine.initOptions.SegmenterOpts.Hmm

	if engine.initOptions.SegmenterOpts.GseMode {
		segments = engine.segmenter.CutSearch(content, hmm)
	} else {
		segments = engine.segmenter.Cut(content, hmm)
	}
	for _, token := range segments {
		if !engine.stopTokens.IsStopToken(token) {
			keywords = append(keywords, token)
		}
	}
	return
}

// Tokens get the engine tokens
func (engine *Engine) Tokens(request types.SearchReq) (tokens []string) {
	// 收集关键词
	// tokens := []string{}
	if request.Text != "" {
		reqText := strings.ToLower(request.Text)
		tokens = engine.Segment(reqText)

		// 叠加 tokens
		// TODO: 去掉外界分词功能，不留接口
		for _, t := range request.Tokens {
			tokens = append(tokens, t)
		}
		return
	}

	for _, t := range request.Tokens {
		tokens = append(tokens, t)
	}
	return
}

func maxRankOutput(rankOpts types.RankOpts, rankLen int) (int, int) {
	var start, end int
	if rankOpts.MaxOutputs == 0 {
		start = utils.MinInt(rankOpts.OutputOffset, rankLen)
		end = rankLen
		return start, end
	}

	start = utils.MinInt(rankOpts.OutputOffset, rankLen)
	end = utils.MinInt(start+rankOpts.MaxOutputs, rankLen)
	return start, end
}

func (engine *Engine) rankOutID(rankerOutput rankerReturnReq, rankOutArr types.ScoredIDs) types.ScoredIDs {
	for _, doc := range rankerOutput.docs.(types.ScoredIDs) {
		rankOutArr = append(rankOutArr, doc)
	}
	return rankOutArr
}

func (engine *Engine) rankOutDocs(rankerOutput rankerReturnReq, rankOutArr types.ScoredDocs) types.ScoredDocs {
	for _, doc := range rankerOutput.docs.(types.ScoredDocs) {
		rankOutArr = append(rankOutArr, doc)
	}
	return rankOutArr
}

// NotTimeOut not set engine timeout
func (engine *Engine) NotTimeOut(request types.SearchReq, rankerReturnChan chan rankerReturnReq) (
	rankOutArr interface{}, numDocs int) {

	var (
		rankOutDoc types.ScoredDocs
	)

	for shard := 0; shard < engine.initOptions.NumShards; shard++ {
		rankerOutput := <-rankerReturnChan
		if !request.CountDocsOnly {
			if rankerOutput.docs != nil {

				rankOutDoc = engine.rankOutDocs(rankerOutput, rankOutDoc)

			}
		}
		numDocs += rankerOutput.numDocs
	}
	rankOutArr = rankOutDoc
	return
}

// TimeOut set engine timeout
func (engine *Engine) TimeOut(request types.SearchReq, rankerReturnChan chan rankerReturnReq) (
	rankOutArr interface{}, numDocs int, isTimeout bool) {

	deadline := time.Now().Add(time.Nanosecond * time.Duration(NumNanosecondsInAMillisecond*request.Timeout))

	var (
		rankOutDoc types.ScoredDocs
	)

	for shard := 0; shard < engine.initOptions.NumShards; shard++ {
		select {
		case rankerOutput := <-rankerReturnChan:
			if !request.CountDocsOnly {
				if rankerOutput.docs != nil {

					rankOutDoc = engine.rankOutDocs(rankerOutput, rankOutDoc)

				}
			}
			numDocs += rankerOutput.numDocs
		case <-time.After(deadline.Sub(time.Now())):
			isTimeout = true
			break
		}
	}

	rankOutArr = rankOutDoc
	return
}

// RankID rank docs by types.ScoredIDs
func (engine *Engine) RankID(request types.SearchReq, rankOpts types.RankOpts, tokens []string, rankerReturnChan chan rankerReturnReq) (output types.SearchResp) {
	// 从通信通道读取排序器的输出
	numDocs := 0
	rankOutput := types.ScoredIDs{}

	//**********/ begin
	timeout := request.Timeout
	isTimeout := false
	if timeout <= 0 {
		// 不设置超时
		rankOutArr, num := engine.NotTimeOut(request, rankerReturnChan)
		rankOutput = rankOutArr.(types.ScoredIDs)
		numDocs += num
	} else {
		// 设置超时
		rankOutArr, num, timeout := engine.TimeOut(request, rankerReturnChan)
		rankOutput = rankOutArr.(types.ScoredIDs)
		numDocs += num
		isTimeout = timeout
	}

	// 再排序
	if !request.CountDocsOnly && !request.Orderless {
		if rankOpts.ReverseOrder {
			sort.Sort(sort.Reverse(rankOutput))
		} else {
			sort.Sort(rankOutput)
		}
	}

	// 准备输出
	output.Tokens = tokens
	// 仅当 CountDocsOnly 为 false 时才充填 output.Docs
	if !request.CountDocsOnly {
		if request.Orderless {
			// 无序状态无需对 Offset 截断
			output.Docs = rankOutput
		} else {
			rankOutLen := len(rankOutput)
			start, end := maxRankOutput(rankOpts, rankOutLen)

			output.Docs = rankOutput[start:end]
		}
	}

	output.NumDocs = numDocs
	output.Timeout = isTimeout

	return
}

// Ranks rank docs by types.ScoredDocs
func (engine *Engine) Ranks(request types.SearchReq, rankOpts types.RankOpts, tokens []string, rankerReturnChan chan rankerReturnReq) (output types.SearchResp) {
	// 从通信通道读取排序器的输出
	numDocs := 0
	rankOutput := types.ScoredDocs{}

	//**********/ begin
	timeout := request.Timeout
	isTimeout := false
	if timeout <= 0 {
		// 不设置超时
		rankOutArr, num := engine.NotTimeOut(request, rankerReturnChan)
		rankOutput = rankOutArr.(types.ScoredDocs)
		numDocs += num
	} else {
		// 设置超时
		rankOutArr, num, timeout := engine.TimeOut(request, rankerReturnChan)
		rankOutput = rankOutArr.(types.ScoredDocs)
		numDocs += num
		isTimeout = timeout
	}

	// 再排序
	if !request.CountDocsOnly && !request.Orderless {
		if rankOpts.ReverseOrder {
			sort.Sort(sort.Reverse(rankOutput))
		} else {
			sort.Sort(rankOutput)
		}
	}

	// 准备输出
	output.Tokens = tokens
	// 仅当 CountDocsOnly 为 false 时才充填 output.Docs
	if !request.CountDocsOnly {
		if request.Orderless {
			// 无序状态无需对 Offset 截断
			output.Docs = rankOutput
		} else {
			rankOutLen := len(rankOutput)
			start, end := maxRankOutput(rankOpts, rankOutLen)

			output.Docs = rankOutput[start:end]
		}
	}

	output.NumDocs = numDocs
	output.Timeout = isTimeout

	return
}

// Flush block wait until all indexes are added
// 阻塞等待直到所有索引添加完毕
func (engine *Engine) Flush() {
	for {
		runtime.Gosched()

		inxd := engine.numIndexingReqs == atomic.LoadUint64(&engine.numDocsIndexed)
		numRm := engine.numRemovingReqs * uint64(engine.initOptions.NumShards)
		rmd := numRm == atomic.LoadUint64(&engine.numDocsRemoved)

		if inxd && rmd {
			// 保证 CHANNEL 中 REQUESTS 全部被执行完
			break
		}
	}

	// 强制更新，保证其为最后的请求
	engine.IndexDoc("0", types.DocData{})
	for {
		runtime.Gosched()

		numf := engine.numForceUpdatingReqs * uint64(engine.initOptions.NumShards)
		forced := numf == atomic.LoadUint64(&engine.numDocsForceUpdated)

		if forced {
			return
		}

	}
}

// FlushIndex block wait until all indexes are added
// 阻塞等待直到所有索引添加完毕
func (engine *Engine) FlushIndex() {
	engine.Flush()
}

// Close close the engine
// 关闭引擎
func (engine *Engine) Close() {
	engine.Flush()
	for _, db := range engine.dbs {
		db.Close()
	}
}

// 从文本hash得到要分配到的 shard
func (engine *Engine) getShard(hash uint32) int {
	n := engine.initOptions.NumShards
	return int(hash - hash/uint32(n)*uint32(n))
}
