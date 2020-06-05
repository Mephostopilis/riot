package types

type StoreOpts struct {
	// 是否使用持久数据库，以及数据库文件保存的目录和裂分数目
	StoreOnly bool `toml:"store_only"`
	UseStore  bool `toml:"use_store"`

	StoreFolder     string `toml:"store_folder"`
	StoreShards     int    `toml:"store_shards"`
	StoreEngine     string `toml:"store_engine"`
	StoreFilePrefix string `toml:"riot_"`
}
