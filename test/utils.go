package test

import (
	"reflect"

	"github.com/go-ego/riot/types"
)

type ScoringFields struct {
	A, B, C float32
}

var (
	Text1 = "Hello world, 你好世界!"
	Text2 = "在路上, in the way"

	TextJP  = "こんにちは世界"
	TextJP1 = "こんにちは世界, こんにちは"
	ReqText = "World人口"

	ReqG = types.SearchReq{Text: "Google"}
	Req1 = types.SearchReq{Text: ReqText}

	Score1   = ScoringFields{1, 2, 3}
	Score091 = ScoringFields{0, 9, 1}

	InxOpts = &types.IndexerOpts{IndexType: types.LocsIndex}
)

var (
	RankOptsMax1 = RankOptsMax(0, 1)

	RankOptsMax10      = rankOptsOrder(false)
	RankOptsMax10Order = rankOptsOrder(true)

	RankOptsMax3 = RankOptsMax(1, 3)
)

type TestScoringCriteria struct {
}

func (criteria TestScoringCriteria) Score(doc types.IndexedDoc, fields interface{}) []float32 {
	if reflect.TypeOf(fields) != reflect.TypeOf(ScoringFields{}) {
		return []float32{}
	}
	fs := fields.(ScoringFields)
	return []float32{float32(doc.TokenProximity)*fs.A + fs.B*fs.C}
}

var (
	EngOpts = types.EngineOpts{
		Using:   1,
		GseDict: "./testdata/test_dict.txt",
		DefRankOpts: &types.RankOpts{
			ScoringCriteria: TestScoringCriteria{},
		},
		IndexerOpts: InxOpts,
	}
)

func MakeDocIds() map[string]bool {
	docIds := make(map[string]bool)
	docIds["5"] = true
	docIds["3"] = true
	docIds["1"] = true
	docIds["2"] = true

	return docIds
}

type RankByTokenProximity struct {
}

func (rule RankByTokenProximity) Score(
	doc types.IndexedDoc, fields interface{}) []float32 {
	if doc.TokenProximity < 0 {
		return []float32{}
	}
	return []float32{1.0 / (float32(doc.TokenProximity) + 1)}
}

func OrderlessOpts(idOnly bool) types.EngineOpts {
	return types.EngineOpts{
		Using:   1,
		IDOnly:  idOnly,
		GseDict: "./testdata/test_dict.txt",
	}
}

func rankEngineOpts(rankOpts types.RankOpts) types.EngineOpts {
	return types.EngineOpts{
		Using:       1,
		GseDict:     "./testdata/test_dict.txt",
		DefRankOpts: &rankOpts,
		IndexerOpts: InxOpts,
	}
}

func rankOptsOrder(order bool) types.RankOpts {
	return types.RankOpts{
		ReverseOrder:    order,
		OutputOffset:    0,
		MaxOutputs:      10,
		ScoringCriteria: &RankByTokenProximity{},
	}
}

func RankOptsMax(output, max int) types.RankOpts {
	return types.RankOpts{
		ReverseOrder:    true,
		OutputOffset:    output,
		MaxOutputs:      max,
		ScoringCriteria: &RankByTokenProximity{},
	}
}

var (
	TestIndexOpts = rankEngineOpts(RankOptsMax10)

	OrderOpts = rankEngineOpts(RankOptsMax10Order)
)
