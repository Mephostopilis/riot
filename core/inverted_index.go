package core

// KeywordIndices 反向索引表的一项，收集了一个搜索键出现的所有文档，按照DocId从小到大排序。
type KeywordIndices struct {
	// 下面的切片是否为空，取决于初始化时IndexType的值
	docID       string  // 全部类型都有
	frequencies float32 // IndexType == FrequenciesIndex
	locations   []int   // IndexType == LocsIndex
}

// 倒排列表
type PostingList struct {
	list []*KeywordIndices
}

type InvertedIndex struct {
	table map[string]PostingList
	// 每个文档的关键词长度
	docTokenLens map[string]float32
	// 所有被索引文本的总关键词数
	totalTokenLen float32
}

func NewInvertedIndex() (index *InvertedIndex) {
	index = new(InvertedIndex)
	index.docTokenLens = make(map[string]float32)
	return
}
