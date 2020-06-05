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
	"strings"

	"github.com/go-ego/gpy"
	"github.com/go-ego/gse"
	"github.com/go-ego/riot/types"
)

// TMap defines the tokens map type map[string][]int
type TMap map[string][]int

type segmenterReq struct {
	docId string
	hash  uint32
	data  types.DocData
	// data        types.DocumentIndexData
	forceUpdate bool
}

func (engine *Engine) initSegment(options *types.EngineOpts) {

	engine.segmenter = &gse.Segmenter{}
	engine.segmenter.LoadDict(options.SegmenterOpts.GseDict)

	// 初始化停用词
	engine.stopTokens.Init(options.SegmenterOpts.StopTokenFile)

	// 初始化分词器通道
	engine.segmenterChan = make(chan segmenterReq, options.SegmenterOpts.NumGseThreads)
}

func (engine *Engine) splitData(request segmenterReq) (TMap, int) {
	var (
		numTokens int
	)
	tokensMap := make(map[string][]int)
	options := engine.initOptions.SegmenterOpts

	if request.data.Content != "" {
		content := strings.ToLower(request.data.Content)
		if options.Using == 3 {
			// use segmenter
			segments := engine.segmenter.ModeSegment([]byte(content), options.GseMode)
			for _, segment := range segments {
				token := segment.Token().Text()
				if !engine.stopTokens.IsStopToken(token) {
					tokensMap[token] = append(tokensMap[token], segment.Start())
				}
			}
			numTokens += len(segments)
		}
	}

	for _, t := range request.data.Tokens {
		if !engine.stopTokens.IsStopToken(t.Text) {
			tokensMap[t.Text] = t.Locations
		}
	}

	numTokens += len(request.data.Tokens)
	return tokensMap, numTokens
}

func (engine *Engine) segmenterData(request segmenterReq) (TMap, int) {
	tokensMap := make(map[string][]int)
	numTokens := 0
	options := engine.initOptions.SegmenterOpts

	if options.Using == 0 && request.data.Content != "" {
		// Content 分词, 当文档正文不为空时，优先从内容分词中得到关键词
		segments := engine.segmenter.ModeSegment([]byte(request.data.Content), options.GseMode)

		for _, segment := range segments {
			token := segment.Token().Text()
			if !engine.stopTokens.IsStopToken(token) {
				tokensMap[token] = append(tokensMap[token], segment.Start())
			}
		}

		for _, t := range request.data.Tokens {
			if !engine.stopTokens.IsStopToken(t.Text) {
				tokensMap[t.Text] = t.Locations
			}
		}

		numTokens = len(segments) + len(request.data.Tokens)
		return tokensMap, numTokens
	}

	if options.Using == 1 && request.data.Content != "" {
		// Content 分词, 当文档正文不为空时，优先从内容分词中得到关键词
		segments := engine.segmenter.ModeSegment([]byte(request.data.Content),
			options.GseMode)

		for _, segment := range segments {
			token := segment.Token().Text()
			if !engine.stopTokens.IsStopToken(token) {
				tokensMap[token] = append(tokensMap[token], segment.Start())
			}
		}
		numTokens = len(segments)

		return tokensMap, numTokens
	}

	useOpts := options.Using == 1 || options.Using == 3
	contentNil := request.data.Content == ""
	opts := useOpts && contentNil

	if options.Using == 2 || opts {
		for _, t := range request.data.Tokens {
			if !engine.stopTokens.IsStopToken(t.Text) {
				tokensMap[t.Text] = t.Locations
			}
		}

		numTokens = len(request.data.Tokens)

		return tokensMap, numTokens
	}

	tokenMap, lenSplitData := engine.splitData(request)

	return tokenMap, lenSplitData
}

// PinYin get the Chinese alphabet and abbreviation
func (engine *Engine) pinyin(hans string) []string {
	var (
		str      string
		pyStr    string
		strArr   []string
		splitStr string
		// splitArr []string
	)

	splitHans := strings.Split(hans, "")
	for i := 0; i < len(splitHans); i++ {
		if splitHans[i] != "" {
			if !engine.stopTokens.IsStopToken(splitHans[i]) {
				strArr = append(strArr, splitHans[i])
			}
			splitStr += splitHans[i]
		}
		if !engine.stopTokens.IsStopToken(splitStr) {
			strArr = append(strArr, splitStr)
		}
	}

	sehans := engine.Segment(hans)
	for h := 0; h < len(sehans); h++ {
		if !engine.stopTokens.IsStopToken(sehans[h]) {
			strArr = append(strArr, sehans[h])
		}
	}

	py := gpy.LazyConvert(hans, nil)
	for i := 0; i < len(py); i++ {
		// log.Println("py[i]...", py[i])
		pyStr += py[i]
		if !engine.stopTokens.IsStopToken(pyStr) {
			strArr = append(strArr, pyStr)
		}

		if len(py[i]) > 0 {
			str += py[i][0:1]
			if !engine.stopTokens.IsStopToken(str) {
				strArr = append(strArr, str)
			}
		}
	}
	return strArr
}

func (engine *Engine) makeTokensMap(request segmenterReq) (TMap, int) {
	tokensMap := make(map[string][]int)
	numTokens := 0
	options := engine.initOptions.SegmenterOpts
	tokensMap, numTokens = engine.segmenterData(request)

	if options.PinYin {
		strArr := engine.pinyin(request.data.Content)
		count := len(strArr)

		for i := 0; i < count; i++ {
			str := strArr[i]
			if !engine.stopTokens.IsStopToken(str) {
				tokensMap[str] = []int{i}
			}
		}
		numTokens += count
	}

	return tokensMap, numTokens
}

func (engine *Engine) segmenterWorker() {
	for {
		select {
		case request := <-engine.segmenterChan:
			if request.docId == "0" {
				if request.forceUpdate {
					for i := 0; i < engine.initOptions.NumShards; i++ {
						engine.indexerAddDocChans[i] <- indexerAddDocReq{forceUpdate: true}
					}
				}
				continue
			}

			shard := engine.getShard(request.hash)
			tokensMap, numTokens := engine.makeTokensMap(request)

			for _, label := range request.data.Labels {
				if !engine.stopTokens.IsStopToken(label) {
					// 当正文中已存在关键字时，若不判断，位置信息将会丢失
					if _, ok := tokensMap[label]; !ok {
						tokensMap[label] = []int{}
					}
				}
			}

			indexerRequest := indexerAddDocReq{
				doc: &types.DocIndex{
					DocId:    request.docId,
					TokenLen: float32(numTokens),
					Keywords: make([]types.KeywordIndex, len(tokensMap)),
				},
				forceUpdate: request.forceUpdate,
			}
			iTokens := 0
			for k, v := range tokensMap {
				indexerRequest.doc.Keywords[iTokens] = types.KeywordIndex{
					Text: k,
					// 非分词标注的词频设置为0，不参与tf-idf计算
					Frequency: float32(len(v)),
					Starts:    v}
				iTokens++
			}

			engine.indexerAddDocChans[shard] <- indexerRequest
			if request.forceUpdate {
				for i := 0; i < engine.initOptions.NumShards; i++ {
					if i == shard {
						continue
					}
					engine.indexerAddDocChans[i] <- indexerAddDocReq{forceUpdate: true}
				}
			}
			rankerRequest := rankerAddDocReq{
				// docId: request.docId, fields: request.data.Fields}
				docId: request.docId, fields: request.data.Fields,
				content: request.data.Content, attri: request.data.Attri}
			engine.rankerAddDocChans[shard] <- rankerRequest
		}

	}
}
