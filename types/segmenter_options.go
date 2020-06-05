package types

const (
	SegmenterUsingTy1 = 1
)

type SegmenterOpts struct {
	// new, 分词规则
	Using int `toml:"using"`

	// 半角逗号 "," 分隔的字典文件，具体用法见
	// gse.Segmenter.LoadDict 函数的注释
	GseDict string `toml:"gse_dict"`
	PinYin  bool   `toml:"pinyin"`

	// 停用词文件
	StopTokenFile string `toml:"stop_file"`

	// Gse search mode
	GseMode bool   `toml:"gse_mode"`
	Hmm     bool   `toml:"hmm"`
	Model   string `toml:"model"`

	// 分词器线程数
	NumGseThreads int
}
