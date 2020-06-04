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

// Config search config options
type Config struct {
	Engine EngineConfig

	Url []string
}

// Engine search engine options
type EngineConfig struct {
	Mode  string
	Using int

	StoreShards int    `toml:"store_shards"`
	StoreEngine string `toml:"store_engine"`
	StoreFolder string `toml:"store_folder"`

	NumShards    int `toml:"num_shards"`
	OutputOffset int `toml:"output_offset"`
	MaxOutputs   int `toml:"max_outputs"`

	GseDict       string `toml:"gse_dict"`
	GseMode       string `toml:"gse_mode"`
	StopTokenFile string `toml:"stop_token_file"`

	Relation string
	Time     string
	Ts       int64
}
