package riot

import (
	"encoding/gob"
	"fmt"
	"os"
	"testing"

	"github.com/go-ego/riot/test"
	"github.com/go-ego/riot/types"
	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"github.com/go-kratos/kratos/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestEngineIndexWithNewStore(t *testing.T) {
	gob.Register(test.ScoringFields{})
	engine := New(
		types.EngineOpts{
			SegmenterOpts: &types.SegmenterOpts{
				GseDict: "zh,./configs/dict/dictionary.txt,testdata/test_dict.txt",
			},
			StoreOpts: &types.StoreOpts{
				StoreFolder: "./riot.new",
			},
			NumShards: 8})

	log.Info("new engine start...")

	AddDocs(engine)

	engine.RemoveDoc("5", true)
	engine.Flush()

	engine.Close()
	// os.RemoveAll("riot.new")

	var (
		engineOpts types.EngineOpts
		ct         paladin.TOML
	)
	if err := paladin.Get("riot.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err := ct.Get("Engine").UnmarshalTOML(&engineOpts); err != nil {
		return
	}

	var engine1 = New(engineOpts)
	log.Info("test...")
	engine1.Flush()
	log.Info("new engine1 start...")

	outputs := engine1.Search(types.SearchReq{Text: test.ReqText})
	assert.Equal(t, "2", len(outputs.Tokens))
	assert.Equal(t, "world", outputs.Tokens[0])
	assert.Equal(t, "人口", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	assert.Equal(t, "2", len(outDocs))

	// assert.Equal(t, "2", outDocs[0].DocId)
	assert.Equal(t, "2500", int(outDocs[0].Scores[0]*1000))
	assert.Equal(t, "[]", outDocs[0].TokenSnippetLocs)

	// assert.Equal(t, "1", outDocs[1].DocId)
	assert.Equal(t, "2215", int(outDocs[1].Scores[0]*1000))
	assert.Equal(t, "[]", outDocs[1].TokenSnippetLocs)

	engine1.Close()
	os.RemoveAll("riot.new")
	// os.RemoveAll("riot-index")
}

var (
	rankTestOpts = test.RankOptsMax(0, 1)
)

func testRankOpt(idOnly bool) types.EngineOpts {
	return types.EngineOpts{
		SegmenterOpts: &types.SegmenterOpts{
			Using:   1,
			GseDict: "zh,./testdata/test_dict.txt",
		},
		DefRankOpts: &rankTestOpts,
		IndexerOpts: test.InxOpts,
	}
}

func lookupReq(engine *Engine) (types.SearchReq, []string, chan rankerReturnReq) {
	request := types.SearchReq{
		Text:   test.ReqText,
		DocIds: test.MakeDocIds(),
	}

	tokens := engine.Tokens(request)
	// 建立排序器返回的通信通道
	rankerReturnChan := make(
		chan rankerReturnReq, engine.initOptions.NumShards)

	// 生成查找请求
	lookupRequest := indexerLookupReq{
		countDocsOnly:    request.CountDocsOnly,
		tokens:           tokens,
		labels:           request.Labels,
		docIds:           request.DocIds,
		options:          rankTestOpts,
		rankerReturnChan: rankerReturnChan,
		orderless:        request.Orderless,
		logic:            request.Logic,
	}

	// 向索引器发送查找请求
	for shard := 0; shard < engine.initOptions.NumShards; shard++ {
		engine.indexerLookupChans[shard] <- lookupRequest
	}

	return request, tokens, rankerReturnChan
}

func TestDocRankID(t *testing.T) {

	engine := New(testRankOpt(true))
	AddDocs(engine)

	engine.RemoveDoc("5")
	engine.Flush()

	request, tokens, rankerReturnChan := lookupReq(engine)
	outputs := engine.RankID(request, rankTestOpts, tokens, rankerReturnChan)

	if outputs.Docs != nil {
		outDocs := outputs.Docs.(types.ScoredIDs)
		assert.Equal(t, "1", len(outDocs))
	}
	assert.Equal(t, "2", len(outputs.Tokens))
	assert.Equal(t, "2", outputs.NumDocs)

	engine.Close()
}

func TestDocRanks(t *testing.T) {

	engine := New(testRankOpt(false))
	AddDocs(engine)

	engine.RemoveDoc("5")
	engine.Flush()

	request, tokens, rankerReturnChan := lookupReq(engine)
	outputs := engine.Ranks(request, rankTestOpts, tokens, rankerReturnChan)

	if outputs.Docs != nil {
		outDocs := outputs.Docs.(types.ScoredDocs)
		assert.Equal(t, "1", len(outDocs))
	}
	assert.Equal(t, "2", len(outputs.Tokens))
	assert.Equal(t, "2", outputs.NumDocs)

	// test search
	outputs1 := engine.Search(types.SearchReq{
		Text:    test.ReqText,
		Timeout: 1000,
		DocIds:  test.MakeDocIds()})

	if outputs1.Docs != nil {
		outDocs1 := outputs.Docs.(types.ScoredDocs)
		assert.Equal(t, "1", len(outDocs1))
	}
	assert.Equal(t, "2", len(outputs1.Tokens))
	assert.Equal(t, "2", outputs1.NumDocs)

	engine.Close()
}

func TestDocGetAllDocAndID(t *testing.T) {
	gob.Register(test.ScoringFields{})

	opts := types.EngineOpts{
		SegmenterOpts: &types.SegmenterOpts{
			Using:   1,
			GseDict: "zh,./testdata/test_dict.txt",
		},
		NumShards:   5,
		DefRankOpts: &rankTestOpts,
		IndexerOpts: test.InxOpts,
		StoreOpts: &types.StoreOpts{
			UseStore: true,
		},
	}
	engine := New(opts)
	engine.Startup()

	AddDocs(engine)

	engine.RemoveDoc("5")
	engine.Flush()

	allIds := engine.GetDBAllIds()
	fmt.Println("all id", allIds)
	assert.Equal(t, "5", len(allIds))
	assert.Equal(t, "[3 4 1 6 2]", allIds)

	allIds = engine.GetAllDocIds()
	fmt.Println("all doc id", allIds)
	assert.Equal(t, "5", len(allIds))
	assert.Equal(t, "[3 4 1 6 2]", allIds)

	ids, docs := engine.GetDBAllDocs()
	fmt.Println("all id and doc", allIds, docs)
	assert.Equal(t, "5", len(ids))
	assert.Equal(t, "5", len(docs))
	assert.Equal(t, "[3 4 1 6 2]", ids)
	allDoc := `[{The world <nil> [] [] <nil>} {有人口 <nil> [] [] {2 3 1}} {The world, 有七十亿人口人口 <nil> [] [] {1 2 3}} {有七十亿人口 <nil> [] [] {2 3 3}} {The world, 人口 <nil> [] [] <nil>}]`
	assert.Equal(t, allDoc, docs)

	has := engine.HasDoc("5")
	assert.Equal(t, "false", has)

	has = engine.HasDoc("2")
	assert.Equal(t, true, has)
	has = engine.HasDoc("3")
	assert.Equal(t, true, has)
	has = engine.HasDoc("4")
	assert.Equal(t, "true", has)

	dbhas := engine.HasDocDB("5")
	assert.Equal(t, "false", dbhas)

	dbhas = engine.HasDocDB("2")
	assert.Equal(t, true, dbhas)
	dbhas = engine.HasDocDB("3")
	assert.Equal(t, true, dbhas)
	dbhas = engine.HasDocDB("4")
	assert.Equal(t, "true", dbhas)

	docIds := make(map[string]bool)
	docIds["5"] = true
	docIds["1"] = true

	outputs := engine.Search(types.SearchReq{
		Text:   test.ReqText,
		DocIds: docIds})

	if outputs.Docs != nil {
		outDocs := outputs.Docs.(types.ScoredIDs)
		fmt.Println("output docs: ", outputs)
		assert.Equal(t, "1", len(outDocs))
	}
	assert.Equal(t, "2", len(outputs.Tokens))
	assert.Equal(t, "1", outputs.NumDocs)

	engine.Close()
	os.RemoveAll("riot.id")
}

func testOpts(use int, store string, args ...bool) types.EngineOpts {
	var pinyin bool
	if len(args) > 0 {
		pinyin = args[0]
	}

	return types.EngineOpts{
		// Using:      1,
		SegmenterOpts: &types.SegmenterOpts{
			Using:   use,
			PinYin:  pinyin,
			GseDict: "zh,./testdata/test_dict.txt",
		},
		StoreOpts: &types.StoreOpts{
			UseStore:    true,
			StoreFolder: store,
		},
	}
}

func TestDocPinYin(t *testing.T) {

	engine := New(testOpts(0, "riot.py"))
	pinyinOpt := New(testOpts(0, "riot.py.opt", true))

	tokens := engine.pinyin(test.Text2)
	fmt.Println("tokens...", tokens)
	assert.Equal(t, "52", len(tokens))

	var tokenDatas []types.TokenData
	// tokens := []string{"z", "zl"}
	for i := 0; i < len(tokens); i++ {
		tokenData := types.TokenData{Text: tokens[i]}
		tokenDatas = append(tokenDatas, tokenData)
	}

	index1 := types.DocData{Tokens: tokenDatas, Fields: "在路上"}
	index2 := types.DocData{Content: test.Text2, Tokens: tokenDatas}

	engine.Index("10", index1)
	engine.Index("11", index2)
	engine.Flush()

	data := types.DocData{Content: test.Text2}
	pinyinOpt.Index("10", data)
	pinyinOpt.Index("11", data)
	pinyinOpt.Flush()

	docIds := make(map[string]bool)
	docIds["5"] = true
	docIds["10"] = true
	docIds["11"] = true

	pyOutputs := pinyinOpt.SearchID(types.SearchReq{
		Text:   "zl",
		DocIds: docIds,
	})

	assert.Equal(t, "2", len(pyOutputs.Docs))
	assert.Equal(t, "1", len(pyOutputs.Tokens))
	assert.Equal(t, "2", pyOutputs.NumDocs)

	outputs := engine.Search(types.SearchReq{
		Text:   "zl",
		DocIds: docIds,
	})

	fmt.Println("outputs", outputs.Docs)
	if outputs.Docs != nil {
		outDocs := outputs.Docs.(types.ScoredIDs)
		assert.Equal(t, "2", len(outDocs))
		// assert.Equal(t, "11", outDocs[0].DocId)
		// assert.Equal(t, "10", outDocs[1].DocId)
	}
	assert.Equal(t, "1", len(outputs.Tokens))
	assert.Equal(t, "2", outputs.NumDocs)

	engine.Close()
	pinyinOpt.Close()
	os.RemoveAll("riot.py")
	os.RemoveAll("riot.py.opt")
}

func TestForSplitData(t *testing.T) {

	engine := New(testOpts(4, "riot.data"))
	engine.Startup()

	AddDocs(engine)

	engine.RemoveDoc("5")
	engine.Flush()

	// TODO:
	// 测试数据
	// tokenDatas := engine.pinyin(test.Text2)
	// tokens, num := engine.ForSplitData(tokenDatas, 52)
	// assert.Equal(t, "93", len(tokens))
	// assert.Equal(t, "104", num)

	// index1 := types.DocData{Content: "在路上"}
	// engine.Index("10", index1, true)

	// docIds := make(map[string]bool)
	// docIds["5"] = true
	// docIds["1"] = true
	// outputs := engine.Search(types.SearchReq{
	// 	Text:   test.ReqText,
	// 	DocIds: docIds})

	// if outputs.Docs != nil {
	// 	outDocs := outputs.Docs.(types.ScoredIDs)
	// 	assert.Equal(t, "0", len(outDocs))
	// }
	// assert.Equal(t, "2", len(outputs.Tokens))
	// assert.Equal(t, "0", outputs.NumDocs)

	// engine.Close()
	// os.RemoveAll("riot.data")
}

func testNum(t *testing.T, numAdd, numInx, numRm uint64) {
	assert.Equal(t, "26", numAdd)
	assert.Equal(t, "6", numInx)
	assert.Equal(t, "8", numRm)
}

func TestDocCounters(t *testing.T) {

	engine := New(testOpts(1, "riot.doc"))

	AddDocs(engine)
	engine.RemoveDoc("5")
	engine.Flush()

	numAdd := engine.NumTokenAdded()
	numInx := engine.NumIndexed()
	numRm := engine.NumRemoved()
	testNum(t, numAdd, numInx, numRm)

	numAdd = engine.NumTokenIndexAdded()
	numInx = engine.NumDocsIndexed()
	numRm = engine.NumDocsRemoved()
	testNum(t, numAdd, numInx, numRm)

	docIds := make(map[string]bool)
	docIds["5"] = true
	docIds["1"] = true

	outputs := engine.Search(types.SearchReq{
		Text:   test.ReqText,
		DocIds: docIds})

	if outputs.Docs != nil {
		outDocs := outputs.Docs.(types.ScoredIDs)
		assert.Equal(t, "1", len(outDocs))
	}
	assert.Equal(t, "2", len(outputs.Tokens))
	assert.Equal(t, "1", outputs.NumDocs)

	engine.Close()
	os.RemoveAll("riot.doc")
}
