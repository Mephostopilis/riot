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

package store

import (
	"os"
	"testing"

	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"github.com/go-kratos/kratos/pkg/log"
	"github.com/stretchr/testify/assert"
)

var TestDBName = "./db_test"

func TestMain(m *testing.M) {
	if err := paladin.Init(); err != nil {
		panic(err)
	}
	log.Init(nil) // debug flag: log.dir={path}
	defer log.Close()
	log.Info("TestGetVer start")
}

func TestBadger(t *testing.T) {
	db, err := OpenBadger(TestDBName)
	assert.Equal(t, "<nil>", err)
	if err != nil {
		log.Error("%v", err)
		return
	}

	log.Error("TestBadger...")
	DBTest(t, db)
	// defer db.Close()
}

func TestLdb(t *testing.T) {
	db, err := OpenLeveldb(TestDBName)
	assert.Equal(t, "<nil>", err)
	if err != nil {
		log.Error("%v", err)
		return
	}

	log.Info("TestLdb...")
	DBTest(t, db)
	// defer db.Close()
}

func DBTest(t *testing.T, db Store) {
	log.Info("db test...")
	os.MkdirAll(TestDBName, 0777)

	err := db.Set([]byte("key1"), []byte("value1"))
	assert.Equal(t, "<nil>", err)

	has, err := db.Has([]byte("key1"))
	assert.Equal(t, nil, err)
	if err == nil {
		assert.Equal(t, true, has)
	}

	buf := make([]byte, 100)
	buf, err = db.Get([]byte("key1"))
	assert.Equal(t, "<nil>", err)
	assert.Equal(t, "value1", string(buf))

	walFile := db.WALName()
	db.Close()
	os.Remove(walFile)
	os.RemoveAll(TestDBName)
}
