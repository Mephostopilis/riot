package types

type StoreOpts struct {
	StoreFolder     string `toml:"store_folder"`
	StoreEngine     string `toml:"store_engine"`
	StoreFilePrefix string `toml:"store_prefix"`
}
