package segmenter

import (
	"strings"

	"github.com/go-ego/riot"
	"github.com/go-ego/riot/types"
)

type Engine struct {
	// 记录初始化参数
	initOptions types.EngineOpts
	stopTokens  riot.StopTokens
}

type TMap map[string][]int

func (engine *Engine) splitData(request segmenterReq) (TMap, int) {
	var (
		num       int
		numTokens int
	)
	tokensMap := make(map[string][]int)
	options := engine.initOptions.SegmenterOpts

	if request.data.Content != "" {
		content := strings.ToLower(request.data.Content)

		if options.Using == 4 {
			tokensMap, numTokens = engine.defaultTokens(content)
		}

		if options.Using != 4 {
			strData := strings.Split(content, "")
			num = len(strData)
			tokenMap, numToken := engine.ForSplitData(strData, num)
			numTokens += numToken
			for key, val := range tokenMap {
				tokensMap[key] = val
			}
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

// ForSplitData for split segment's data, segspl
func (engine *Engine) ForSplitData(strData []string, num int) (TMap, int) {

	var (
		numTokens int
		splitStr  string
	)
	tokensMap := make(map[string][]int)
	options := engine.initOptions.SegmenterOpts

	for i := 0; i < num; i++ {
		if strData[i] != "" {
			if !engine.stopTokens.IsStopToken(strData[i]) {
				numTokens++
				tokensMap[strData[i]] = append(tokensMap[strData[i]], numTokens)
			}
			splitStr += strData[i]
			if !engine.stopTokens.IsStopToken(splitStr) {
				numTokens++
				tokensMap[splitStr] = append(tokensMap[splitStr], numTokens)
			}

			if options.Using == 6 {
				// more combination
				var splitsStr string
				for s := i + 1; s < len(strData); s++ {
					splitsStr += strData[s]

					if !engine.stopTokens.IsStopToken(splitsStr) {
						numTokens++
						tokensMap[splitsStr] = append(tokensMap[splitsStr], numTokens)
					}
				}
			}
		}
	}

	return tokensMap, numTokens
}

func (engine *Engine) defaultTokens(content string) (tokensMap TMap, numTokens int) {
	// use segmenter
	tokensMap = make(map[string][]int)
	strData := strings.Split(content, " ")
	num := len(strData)

	if num > 0 {
		tokenMap, numToken := engine.ForSplitData(strData, num)
		numTokens += numToken

		for key, val := range tokenMap {
			tokensMap[key] = val
		}
	}

	return
}
