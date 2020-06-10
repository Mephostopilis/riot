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

package types

import (
	"runtime"
)

const (

	// NumNanosecondsInAMillisecond nano-seconds in a milli-second num
	NumNanosecondsInAMillisecond = 1000000
	// StoreFilePrefix persistent store file prefix
	StoreFilePrefix = "riot"

	// DefaultPath default db path
	DefaultPath = "./riot-index"
)

var (
	// EngineOpts 的默认值
	// defaultNumGseThreads default Segmenter threads num
	defaultNumGseThreads = runtime.NumCPU()
	// defaultNumShards                 = 2
	defaultNumShards         = 8
	defaultIndexerBufLen     = runtime.NumCPU()
	defaultNumIndexerThreads = runtime.NumCPU()
	defaultRankerBufLen      = runtime.NumCPU()
	defaultNumRankerThreads  = runtime.NumCPU()
	defaultRankOpts          = RankOpts{
		ScoringCriteria: RankByBM25{},
	}
	defaultIndexerOpts = IndexerOpts{
		IndexType:      FrequenciesIndex,
		BM25Parameters: &defaultBM25Parameters,
	}
	defaultBM25Parameters = BM25Parameters{
		K1: 2.0,
		B:  0.75,
	}
	defaultStoreShards = 8
)

// EngineOpts init engine options
type EngineOpts struct {
	SegmenterOpts *SegmenterOpts

	// 索引器和排序器的 shard 数目
	// 被检索/排序的文档会被均匀分配到各个 shard 中
	NumShards int

	// 索引器的信道缓冲长度
	IndexerBufLen int

	// 索引器每个shard分配的线程数
	NumIndexerThreads int

	// 排序器的信道缓冲长度
	RankerBufLen int

	// 排序器每个 shard 分配的线程数
	NumRankerThreads int

	// 索引器初始化选项
	IndexerOpts *IndexerOpts

	// 默认的搜索选项
	DefRankOpts *RankOpts

	StoreOpts *StoreOpts
}

// Init init engine options
// 初始化 EngineOpts，当用户未设定某个选项的值时用默认值取代
func (options *EngineOpts) Init() {
	// if !options.NotUseGse && options.GseDict == "" {
	// 	log.Fatal("字典文件不能为空")
	//  options.GseDict = "zh"
	// }

	if options.SegmenterOpts.NumGseThreads == 0 {
		options.SegmenterOpts.NumGseThreads = defaultNumGseThreads
	}

	if options.NumShards == 0 {
		options.NumShards = defaultNumShards
	}

	if options.IndexerBufLen == 0 {
		options.IndexerBufLen = defaultIndexerBufLen
	}

	if options.NumIndexerThreads == 0 {
		options.NumIndexerThreads = defaultNumIndexerThreads
	}

	if options.RankerBufLen == 0 {
		options.RankerBufLen = defaultRankerBufLen
	}

	if options.NumRankerThreads == 0 {
		options.NumRankerThreads = defaultNumRankerThreads
	}

	if options.IndexerOpts == nil {
		options.IndexerOpts = &defaultIndexerOpts
	}

	if options.IndexerOpts.BM25Parameters == nil {
		options.IndexerOpts.BM25Parameters = &defaultBM25Parameters
	}

	if options.DefRankOpts == nil {
		options.DefRankOpts = &defaultRankOpts
	}

	if options.DefRankOpts.ScoringCriteria == nil {
		options.DefRankOpts.ScoringCriteria = defaultRankOpts.ScoringCriteria
	}

}
