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
	"strconv"

	"github.com/go-ego/riot/store"
	"github.com/go-ego/riot/types"
	"github.com/go-kratos/kratos/pkg/log"
)

// InitStore initialize the persistent store channel
func (engine *Engine) initStore(options *types.EngineOpts) (err error) {
	// 打开或者创建数据库
	engine.dbs = make([]store.Store, 1)
	for shard := 0; shard < 1; shard++ {
		dbPath := options.StoreOpts.StoreFolder + "/" + options.StoreOpts.StoreFilePrefix + "." + strconv.Itoa(shard)
		db, xerr := store.OpenStore(dbPath, options.StoreOpts.StoreEngine)
		if db == nil || xerr != nil {
			log.Error("Unable to open database %s:%v ", dbPath, err)
			err = xerr
			return
		}
		engine.dbs[shard] = db
	}
	return
}
