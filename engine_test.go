package riot

import (
	"encoding/gob"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"testing"

	"github.com/go-ego/gse"
	"github.com/go-ego/riot/test"
	"github.com/go-ego/riot/types"
	"github.com/stretchr/testify/assert"

	"github.com/go-kratos/kratos/pkg/log"
)

func AddDocs(engine *Engine) {
	engine.Index("1", types.DocData{
		Content: "The world, 有七十亿人口人口",
		Fields:  test.Score1,
	})

	engine.IndexDoc("2", types.DocIndexData{
		Content: "The world, 人口",
		Fields:  nil,
	})

	engine.Index("3", types.DocData{
		Content: "The world",
		Fields:  nil,
	})

	engine.Index("4", types.DocData{
		Content: "有人口",
		Fields:  test.ScoringFields{2, 3, 1},
	})

	engine.Index("5", types.DocData{
		Content: "The world, 七十亿人口",
		Fields:  test.Score091,
	})

	engine.Index("6", types.DocData{
		Content: "有七十亿人口",
		Fields:  test.ScoringFields{2, 3, 3},
	})

	engine.Flush()
}

func AddDocsWithLabels(engine *Engine) {
	engine.Index("1", types.DocData{
		Content: "《复仇者联盟3：无限战争》是全片使用IMAX摄影机拍摄",
		Labels:  []string{"复仇者", "战争"},
	})
	log.Info("engine.Segment(): %v", engine.Segment("《复仇者联盟3：无限战争》是全片使用IMAX摄影机拍摄"))

	// docId++
	engine.Index("2", types.DocData{
		Content: "在IMAX影院放映时",
		Labels:  []string{"影院"},
	})

	engine.Index("3", types.DocData{
		Content: " Google 是世界最大搜索引擎, baidu 是最大中文的搜索引擎",
		Labels:  []string{"Google"},
	})

	engine.Index("4", types.DocData{
		Content: "Google 在研制无人汽车",
		Labels:  []string{"Google"},
	})

	engine.Index("5", types.DocData{
		Content: " GAMAF 世界五大互联网巨头, BAT 是中国互联网三巨头",
		Labels:  []string{"互联网"},
	})

	engine.Flush()
}

func TestMain(m *testing.M) {
	// if err := paladin.Init(); err != nil {
	// 	panic(err)
	// }
	log.Init(nil) // debug flag: log.dir={path}
	defer log.Close()
	log.Info("test start")
	m.Run()
}

func TestGetVer(t *testing.T) {
	fmt.Println("go version: ", runtime.Version())
	ver := GetVersion()
	assert.Equal(t, Version, ver)
}

func TestTry(t *testing.T) {
	var arr []int
	Try(func() {
		fmt.Println(arr[2])
	}, func(err interface{}) {
		log.Info("err = %v", err)
		assert.Equal(t, "runtime error: index out of range [2] with length 0", err)
	})
}

func TestEngineIndexDoc(t *testing.T) {

	engine := New(test.TestIndexOpts)
	engine.Startup()

	AddDocs(engine)

	outputs := engine.Search(test.Req1)
	assert.Equal(t, "2", len(outputs.Tokens))
	assert.Equal(t, "world", outputs.Tokens[0])
	assert.Equal(t, "人口", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	assert.Equal(t, "3", len(outDocs))

	log.Info("TestEngineIndexDoc: %v", outDocs)
	assert.Equal(t, "2", outDocs[0].DocId)
	assert.Equal(t, "333", int(outDocs[0].Scores[0]*1000))
	assert.Equal(t, "[4 11]", outDocs[0].TokenSnippetLocs)

	assert.Equal(t, "5", outDocs[1].DocId)
	assert.Equal(t, "83", int(outDocs[1].Scores[0]*1000))
	assert.Equal(t, "[4 20]", outDocs[1].TokenSnippetLocs)

	assert.Equal(t, "1", outDocs[2].DocId)
	assert.Equal(t, "66", int(outDocs[2].Scores[0]*1000))
	assert.Equal(t, "[4 23]", outDocs[2].TokenSnippetLocs)

	engine.Close()
}

func TestReverseOrder(t *testing.T) {
	engine := New(test.OrderOpts)
	engine.Startup()

	AddDocs(engine)

	outputs := engine.Search(test.Req1)

	outDocs := outputs.Docs.(types.ScoredDocs)
	assert.Equal(t, "3", len(outDocs))

	assert.Equal(t, "1", outDocs[0].DocId)
	assert.Equal(t, "5", outDocs[1].DocId)
	assert.Equal(t, "2", outDocs[2].DocId)

	engine.Close()
}

func TestOffsetAndMaxOutputs(t *testing.T) {
	engine := New(types.EngineOpts{
		SegmenterOpts: &types.SegmenterOpts{
			Using:   1,
			GseDict: "zh,./testdata/test_dict.txt",
		},
		DefRankOpts: &test.RankOptsMax3,
		IndexerOpts: test.InxOpts,
	})
	engine.Startup()

	AddDocs(engine)

	outputs := engine.Search(test.Req1)

	outDocs := outputs.Docs.(types.ScoredDocs)
	assert.Equal(t, "2", len(outDocs))
	assert.Equal(t, "5", outDocs[0].DocId)
	assert.Equal(t, "2", outDocs[1].DocId)

	engine.Close()
}

func TestSearchWithCriteria(t *testing.T) {
	engine := New(test.EngOpts)
	engine.Startup()

	AddDocs(engine)

	outputs := engine.Search(test.Req1)

	outDocs := outputs.Docs.(types.ScoredDocs)
	assert.Equal(t, "2", len(outDocs))

	log.Info("%v", outDocs)
	assert.Equal(t, "1", outDocs[0].DocId)
	assert.Equal(t, "20000", int(outDocs[0].Scores[0]*1000))

	assert.Equal(t, "5", outDocs[1].DocId)
	assert.Equal(t, "9000", int(outDocs[1].Scores[0]*1000))

	engine.Close()
}

func TestCompactIndex(t *testing.T) {
	engine := New(useOpts)

	AddDocs(engine)

	outputs := engine.Search(test.Req1)

	outDocs := outputs.Docs.(types.ScoredDocs)
	assert.Equal(t, "2", len(outDocs))

	assert.Equal(t, "5", outDocs[0].DocId)
	assert.Equal(t, "9000", int(outDocs[0].Scores[0]*1000))

	assert.Equal(t, "1", outDocs[1].DocId)
	assert.Equal(t, "6000", int(outDocs[1].Scores[0]*1000))

	engine.Close()
}

type BM25ScoringCriteria struct {
}

func (criteria BM25ScoringCriteria) Score(doc types.IndexedDoc, fields interface{}) []float32 {
	if reflect.TypeOf(fields) != reflect.TypeOf(test.ScoringFields{}) {
		return []float32{}
	}
	return []float32{doc.BM25}
}

func TestFrequenciesIndex(t *testing.T) {
	var engine *Engine = New(types.EngineOpts{
		SegmenterOpts: &types.SegmenterOpts{
			Using:   1,
			GseDict: "zh,./testdata/test_dict.txt",
		},
		DefRankOpts: &types.RankOpts{
			ScoringCriteria: BM25ScoringCriteria{},
		},
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.FrequenciesIndex,
		},
	})

	AddDocs(engine)

	outputs := engine.Search(test.Req1)

	outDocs := outputs.Docs.(types.ScoredDocs)
	assert.Equal(t, "2", len(outDocs))

	assert.Equal(t, "1", outDocs[0].DocId)
	assert.Equal(t, "2374", int(outDocs[0].Scores[0]*1000))

	assert.Equal(t, "5", outDocs[1].DocId)
	assert.Equal(t, "2133", int(outDocs[1].Scores[0]*1000))

	engine.Close()
}

var (
	useOpts = types.EngineOpts{
		SegmenterOpts: &types.SegmenterOpts{
			Using:   1,
			GseDict: "./testdata/test_dict.txt",
		},
		DefRankOpts: &types.RankOpts{
			ScoringCriteria: test.TestScoringCriteria{},
		},
	}
)

func TestRemoveDoc(t *testing.T) {
	engine := New(useOpts)

	AddDocs(engine)

	engine.RemoveDoc("5")
	engine.RemoveDoc("6")
	engine.Flush()
	engine.Index("6", types.DocData{
		Content: "World, 人口有七十亿",
		Fields:  test.Score091,
	})
	engine.Flush()

	outputs := engine.Search(test.Req1)

	outDocs := outputs.Docs.(types.ScoredDocs)
	assert.Equal(t, "2", len(outDocs))

	assert.Equal(t, "6", outDocs[0].DocId)
	assert.Equal(t, "9000", int(outDocs[0].Scores[0]*1000))
	assert.Equal(t, "1", outDocs[1].DocId)
	assert.Equal(t, "6000", int(outDocs[1].Scores[0]*1000))

	engine.Close()
}

func TestEngineIndexWithTokens(t *testing.T) {
	engine := New(test.TestIndexOpts)

	data1 := types.TokenData{
		Text:      "world",
		Locations: []int{0},
	}
	// docId := uint64(1)
	engine.Index("1", types.DocData{
		Content: "",
		Tokens: []types.TokenData{
			data1,
			{"人口", []int{18, 24}},
		},
		Fields: test.Score1,
	})

	// docId++
	engine.Index("2", types.DocData{
		Content: "",
		Tokens: []types.TokenData{
			data1,
			{"人口", []int{6}},
		},
		Fields: test.Score1,
	})

	engine.Index("3", types.DocData{
		Content: "The world, 七十亿人口",
		Fields:  test.Score091,
	})
	engine.FlushIndex()

	outputs := engine.Search(test.Req1)
	log.Info("TestEngineIndexWithTokens: %v", outputs)
	assert.Equal(t, "2", len(outputs.Tokens))
	assert.Equal(t, "world", outputs.Tokens[0])
	assert.Equal(t, "人口", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	assert.Equal(t, "3", len(outDocs))

	assert.Equal(t, "2", outDocs[0].DocId)
	assert.Equal(t, "500", int(outDocs[0].Scores[0]*1000))
	assert.Equal(t, "[0 6]", outDocs[0].TokenSnippetLocs)

	assert.Equal(t, "3", outDocs[1].DocId)
	assert.Equal(t, "83", int(outDocs[1].Scores[0]*1000))
	assert.Equal(t, "[4 20]", outDocs[1].TokenSnippetLocs)

	assert.Equal(t, "1", outDocs[2].DocId)
	assert.Equal(t, "71", int(outDocs[2].Scores[0]*1000))
	assert.Equal(t, "[0 18]", outDocs[2].TokenSnippetLocs)

	engine.Close()
}

func testLabelsOpts(indexType int) types.EngineOpts {
	return types.EngineOpts{
		SegmenterOpts: &types.SegmenterOpts{
			GseDict: "./data/dict/dictionary.txt",
		},
		IndexerOpts: &types.IndexerOpts{
			IndexType: indexType,
		},
	}
}

func TestEngineIndexWithContentAndLabels(t *testing.T) {

	engine1 := New(testLabelsOpts(types.LocsIndex))
	engine2 := New(testLabelsOpts(types.DocIdsIndex))

	AddDocsWithLabels(engine1)
	AddDocsWithLabels(engine2)

	outputs1 := engine1.Search(test.ReqG)
	outputs2 := engine2.Search(test.ReqG)
	assert.Equal(t, "1", len(outputs1.Tokens))
	assert.Equal(t, "1", len(outputs2.Tokens))
	assert.Equal(t, "google", outputs1.Tokens[0])
	assert.Equal(t, "google", outputs2.Tokens[0])

	outDocs := outputs1.Docs.(types.ScoredDocs)
	assert.Equal(t, "2", len(outDocs))
	assert.Equal(t, "2", len(outputs2.Docs.(types.ScoredDocs)))

	engine1.Close()
	engine2.Close()
}

func TestIndexWithLabelsStopTokenFile(t *testing.T) {
	engine1 := New(types.EngineOpts{
		SegmenterOpts: &types.SegmenterOpts{
			GseDict:       "./data/dict/dictionary.txt",
			StopTokenFile: "zh,./testdata/test_stop_dict.txt",
		},
		IndexerOpts: test.InxOpts,
	})

	AddDocsWithLabels(engine1)

	outputs1 := engine1.Search(test.ReqG)
	outputsDoc := engine1.SearchDoc(test.ReqG)
	assert.Equal(t, "1", len(outputs1.Tokens))
	// assert.Equal(t, "Google", outputs1.Tokens[0])

	outDocs := outputs1.Docs.(types.ScoredDocs)
	assert.Equal(t, "2", len(outDocs))
	assert.Equal(t, "2", len(outputsDoc.Docs))
}

func TestEngineIndexWithStore(t *testing.T) {
	gob.Register(test.ScoringFields{})

	var opts = types.EngineOpts{
		SegmenterOpts: &types.SegmenterOpts{
			Using:   1,
			GseDict: "zh,./testdata/test_dict.txt",
		},

		DefRankOpts: &test.RankOptsMax10,
		IndexerOpts: test.InxOpts,
		StoreOpts: &types.StoreOpts{
			UseStore:    true,
			StoreFolder: "riot.persistent",
			StoreShards: 2,
		},
	}

	engine := New(opts)

	AddDocs(engine)

	engine.RemoveDoc("5", true)
	engine.Flush()

	engine.Close()

	engine1 := New(opts)
	engine1.Flush()

	outputs := engine1.Search(test.Req1)
	assert.Equal(t, "2", len(outputs.Tokens))
	assert.Equal(t, "world", outputs.Tokens[0])
	assert.Equal(t, "人口", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	assert.Equal(t, "2", len(outDocs))

	assert.Equal(t, "2", outDocs[0].DocId)
	assert.Equal(t, "333", int(outDocs[0].Scores[0]*1000))
	assert.Equal(t, "[4 11]", outDocs[0].TokenSnippetLocs)

	assert.Equal(t, "1", outDocs[1].DocId)
	assert.Equal(t, "66", int(outDocs[1].Scores[0]*1000))
	assert.Equal(t, "[4 23]", outDocs[1].TokenSnippetLocs)

	engine1.Close()
	os.RemoveAll("riot.persistent")
}

func TestCountDocsOnly(t *testing.T) {
	var engine *Engine = New(types.EngineOpts{
		SegmenterOpts: &types.SegmenterOpts{
			Using:   1,
			GseDict: "zh,./testdata/test_dict.txt",
		},
		DefRankOpts: &test.RankOptsMax1,
		IndexerOpts: test.InxOpts,
	})

	AddDocs(engine)

	engine.RemoveDoc("5")
	engine.Flush()

	outputs := engine.Search(
		types.SearchReq{Text: test.ReqText, CountDocsOnly: true},
	)
	// assert.Equal(t, "0", len(outputs.Docs))
	if outputs.Docs == nil {
		assert.Equal(t, "0", 0)
	}
	assert.Equal(t, "2", len(outputs.Tokens))
	assert.Equal(t, "2", outputs.NumDocs)

	engine.Close()
}

func TestDocOrderless(t *testing.T) {

	engine := New(test.OrderlessOpts(false))

	AddDocs(engine)

	engine.RemoveDoc("5")
	engine.Flush()

	orderReq := types.SearchReq{Text: test.ReqText, Orderless: true}
	outputs := engine.Search(orderReq)
	// assert.Equal(t, "0", len(outputs.Docs))
	if outputs.Docs == nil {
		assert.Equal(t, "0", 0)
	}
	assert.Equal(t, "2", len(outputs.Tokens))
	assert.Equal(t, "2", outputs.NumDocs)

	engine1 := New(test.OrderlessOpts(true))

	AddDocs(engine1)

	engine1.RemoveDoc("5")
	engine1.Flush()

	outputs1 := engine1.Search(orderReq)
	if outputs1.Docs == nil {
		assert.Equal(t, "0", 0)
	}

	assert.Equal(t, "2", len(outputs1.Tokens))
	assert.Equal(t, "2", outputs1.NumDocs)

	engine.Close()
}

var (
	testIDInlyOpts = types.EngineOpts{
		// Using:       1,
		SegmenterOpts: &types.SegmenterOpts{
			GseDict: "./testdata/test_dict.txt",
		},
		DefRankOpts: &test.RankOptsMax1,
		IndexerOpts: test.InxOpts,
		IDOnly:      true,
	}
)

func TestDocOnlyID(t *testing.T) {
	engine := New(testIDInlyOpts)
	AddDocs(engine)

	engine.RemoveDoc("5")
	engine.Flush()

	req := types.SearchReq{
		Text:   test.ReqText,
		DocIds: test.MakeDocIds(),
	}
	outputs := engine.Search(req)
	outputsID := engine.SearchID(req)
	assert.Equal(t, "1", len(outputsID.Docs))

	if outputs.Docs != nil {
		outDocs := outputs.Docs.(types.ScoredIDs)
		assert.Equal(t, "1", len(outDocs))
	}
	assert.Equal(t, "2", len(outputs.Tokens))
	assert.Equal(t, "2", outputs.NumDocs)

	outputs1 := engine.Search(types.SearchReq{
		Text:    test.ReqText,
		Timeout: 10,
		DocIds:  test.MakeDocIds(),
	})

	if outputs1.Docs != nil {
		outDocs1 := outputs.Docs.(types.ScoredIDs)
		assert.Equal(t, "1", len(outDocs1))
	}
	assert.Equal(t, "2", len(outputs1.Tokens))
	assert.Equal(t, "2", outputs1.NumDocs)

	engine.Close()
}

func TestSearchWithin(t *testing.T) {
	engine := New(test.OrderOpts)

	AddDocs(engine)

	docIds := make(map[string]bool)
	docIds["5"] = true
	docIds["1"] = true

	outputs := engine.Search(types.SearchReq{
		Text:   test.ReqText,
		DocIds: docIds,
	})
	assert.Equal(t, "2", len(outputs.Tokens))
	assert.Equal(t, "world", outputs.Tokens[0])
	assert.Equal(t, "人口", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	assert.Equal(t, "2", len(outDocs))

	assert.Equal(t, "1", outDocs[0].DocId)
	assert.Equal(t, "66", int(outDocs[0].Scores[0]*1000))
	assert.Equal(t, "[4 23]", outDocs[0].TokenSnippetLocs)

	assert.Equal(t, "5", outDocs[1].DocId)
	assert.Equal(t, "83", int(outDocs[1].Scores[0]*1000))
	assert.Equal(t, "[4 20]", outDocs[1].TokenSnippetLocs)

	engine.Close()
}

func testJPOpts(use int) types.EngineOpts {
	return types.EngineOpts{
		// Using:           1,
		SegmenterOpts: &types.SegmenterOpts{
			Using:   use,
			GseDict: "zh,./testdata/test_dict_jp.txt",
		},
		DefRankOpts: &test.RankOptsMax10Order,
		IndexerOpts: test.InxOpts,
	}
}

func TestSearchJp(t *testing.T) {

	engine := New(testJPOpts(1))

	AddDocs(engine)

	engine.Index("7", types.DocData{
		Content: test.TextJP1,
		Fields:  test.Score1,
	})
	engine.Flush()

	docIds := make(map[string]bool)
	docIds["5"] = true
	docIds["1"] = true
	docIds["7"] = true

	outputs := engine.Search(types.SearchReq{
		Text:   test.TextJP,
		DocIds: docIds,
	})

	assert.Equal(t, "2", len(outputs.Tokens))
	assert.Equal(t, "こんにちは", outputs.Tokens[0])
	assert.Equal(t, "世界", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	log.Info("outputs docs... %v", outDocs)
	assert.Equal(t, "1", len(outDocs))

	assert.Equal(t, "7", outDocs[0].DocId)
	assert.Equal(t, "1000", int(outDocs[0].Scores[0]*1000))
	assert.Equal(t, "[0 15]", outDocs[0].TokenSnippetLocs)

	engine.Close()
}

func makeGseDocIds() map[string]bool {
	docIds := make(map[string]bool)
	docIds["5"] = true
	docIds["1"] = true
	docIds["6"] = true
	docIds["7"] = true
	docIds["8"] = true

	return docIds
}

func tokenData() []types.TokenData {
	tokenData := types.TokenData{
		Text:      "こんにちは",
		Locations: []int{10, 20},
	}

	return []types.TokenData{tokenData}
}

func TestSearchGse(t *testing.T) {
	log.Info("Test search gse ...")

	engine := New(testJPOpts(0))

	AddDocs(engine)

	engine.Index("7", types.DocData{
		Content: test.TextJP1,
		Fields:  test.Score1,
	})

	tokenDatas := tokenData()
	engine.Index("8", types.DocData{
		Content: test.Text1,
		Tokens:  tokenDatas,
		Fields:  test.ScoringFields{4, 5, 6},
	})
	engine.Flush()

	docIds := makeGseDocIds()
	outputs := engine.Search(types.SearchReq{
		Text:   test.TextJP,
		DocIds: docIds,
	})

	assert.Equal(t, "2", len(outputs.Tokens))
	assert.Equal(t, "こんにちは", outputs.Tokens[0])
	assert.Equal(t, "世界", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	log.Info("outputs docs... %v", outDocs)
	assert.Equal(t, "2", len(outDocs))

	assert.Equal(t, "8", outDocs[0].DocId)
	assert.Equal(t, "142", int(outDocs[0].Scores[0]*1000))
	assert.Equal(t, "[10 19]", outDocs[0].TokenSnippetLocs)

	assert.Equal(t, "7", outDocs[1].DocId)
	assert.Equal(t, "1000", int(outDocs[1].Scores[0]*1000))
	assert.Equal(t, "[0 15]", outDocs[1].TokenSnippetLocs)

	engine.Close()
}

func TestSearchNotUseGse(t *testing.T) {

	engine := New(types.EngineOpts{
		SegmenterOpts: &types.SegmenterOpts{
			Using: 4,
		},
	})

	engine1 := New(types.EngineOpts{
		IDOnly: true,
	})

	AddDocs(engine)
	AddDocs(engine1)

	data := types.DocData{
		Content: "Google Is Experimenting With Virtual Reality Advertising",
		Fields:  test.Score1,
		Tokens:  []types.TokenData{{Text: "test"}},
	}

	engine.Index("7", data)
	engine.Index("8", data)

	engine1.Index("7", data, true)
	engine1.Index("8", data, true)

	engine.Flush()
	engine1.Flush()

	docIds := makeGseDocIds()

	outputs := engine.Search(types.SearchReq{
		Text:   "google is",
		DocIds: docIds,
	})

	assert.Equal(t, "2", len(outputs.Tokens))
	assert.Equal(t, "google", outputs.Tokens[0])
	assert.Equal(t, "is", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	log.Info("outputs docs... %v", outDocs)
	assert.Equal(t, "2", len(outDocs))

	assert.Equal(t, "8", outDocs[0].DocId)
	assert.Equal(t, "3736", int(outDocs[0].Scores[0]*1000))
	assert.Equal(t, "[]", outDocs[0].TokenSnippetLocs)

	outputs1 := engine1.Search(types.SearchReq{
		Text:   "google",
		DocIds: docIds})
	assert.Equal(t, "1", len(outputs1.Tokens))
	assert.Equal(t, "2", outputs1.NumDocs)

	engine.Close()
	engine1.Close()
}

func TestSearchWithGse(t *testing.T) {
	seg := gse.Segmenter{}
	seg.LoadDict("zh") // ./data/dict/dictionary.txt

	searcher2 := New(types.EngineOpts{
		SegmenterOpts: &types.SegmenterOpts{
			Using: 1,
		},
	})
	defer searcher2.Close()

	engine1 := New(types.EngineOpts{
		IndexerOpts: test.InxOpts,
	})
	engine1.WithGse(&seg)
	engine1.Startup()

	engine2 := New(types.EngineOpts{
		// GseDict: "./data/dict/dictionary.txt",
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.DocIdsIndex,
		},
	})
	engine2.WithGse(&seg)

	AddDocsWithLabels(engine1)
	AddDocsWithLabels(engine2)

	outputs1 := engine1.Search(test.ReqG)
	outputs2 := engine2.Search(test.ReqG)
	assert.Equal(t, "1", len(outputs1.Tokens))
	assert.Equal(t, "1", len(outputs2.Tokens))
	assert.Equal(t, "google", outputs1.Tokens[0])
	assert.Equal(t, "google", outputs2.Tokens[0])

	outDocs := outputs1.Docs.(types.ScoredDocs)
	assert.Equal(t, "2", len(outDocs))
	assert.Equal(t, "2", len(outputs2.Docs.(types.ScoredDocs)))

	engine1.Close()
	engine2.Close()
}

func TestRiotGse(t *testing.T) {
	engine := New(types.EngineOpts{
		SegmenterOpts: &types.SegmenterOpts{
			Using: 1,
		},
	})

	AddDocs(engine)

	engine1 := New(types.EngineOpts{
		SegmenterOpts: &types.SegmenterOpts{
			Using:   1,
			GseMode: true,
		},
	})

	AddDocs(engine1)
	assert.Equal(t, "[《 复仇者 联盟 3 ： 无限 战争 》 是 全片 使用 imax 摄影机 拍摄]",
		engine.Segment("《复仇者联盟3：无限战争》是全片使用IMAX摄影机拍摄"))
	assert.Equal(t, "[此次 google 收购 将 成 世界 联网 互联网 最大 并购]",
		engine1.Segment("此次Google收购将成世界互联网最大并购"))

	engine.Close()
	engine1.Close()
}

func TestSearchLogic(t *testing.T) {
	var engine *Engine = New(testJPOpts(0))

	AddDocs(engine)

	engine.Index("7", types.DocData{
		Content: test.TextJP1,
		Fields:  test.Score1,
	})

	tokenDatas := tokenData()
	engine.Index("8", types.DocData{
		Content: test.Text1,
		Tokens:  tokenDatas,
		Fields:  test.Score1,
	})

	engine.Index("9", types.DocData{
		Content: test.Text1,
		Fields:  test.Score1,
	})

	engine.Index("10", types.DocData{
		Content: "Hello, 你好世界!",
		Tokens: []types.TokenData{{
			"世界",
			[]int{2, 3},
		}},
		Fields: test.Score091,
	})

	engine.Flush()

	docIds := make(map[string]bool)
	for index := 0; index < 10; index++ {
		docIds[strconv.Itoa(index)] = true
	}

	strArr := []string{"こんにちは"}
	logic := types.Logic{
		Should: true,
		Expr: types.Expr{
			NotIn: strArr,
		},
	}

	outputs := engine.Search(types.SearchReq{
		Text:   test.TextJP,
		DocIds: docIds,
		Logic:  logic,
	})

	assert.Equal(t, "2", len(outputs.Tokens))
	assert.Equal(t, "こんにちは", outputs.Tokens[0])
	assert.Equal(t, "世界", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	log.Info("outputs docs... %v", outDocs)
	assert.Equal(t, "2", len(outDocs))

	assert.Equal(t, "10", outDocs[0].DocId)
	assert.Equal(t, "1000", int(outDocs[0].Scores[0]*1000))
	assert.Equal(t, "[]", outDocs[0].TokenSnippetLocs)

	assert.Equal(t, "9", outDocs[1].DocId)
	assert.Equal(t, "1000", int(outDocs[1].Scores[0]*1000))
	assert.Equal(t, "[]", outDocs[1].TokenSnippetLocs)

	engine.Close()
}
