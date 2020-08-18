// Copyright 2015 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/branthz/etcd/raft/raftpb"
	"github.com/branthz/utarrow/lib/log"
)

type localConfig struct {
	ID       int
	Listen   string
	Peers    string
	Join     bool
	LogPath  string
	LogLevel string
}

var (
	configPath  string
	LocalConfig localConfig
)

//其实外部交互就两种，控制流和数据流：配置变更，数据提交
//代码里通过channel作为入口
func main() {
	flag.StringVar(&configPath, "c", "./conf.toml", "config path")
	flag.Parse()
	_, err := toml.DecodeFile(configPath, &LocalConfig)
	if err != nil {
		fmt.Println("11111", err)
		os.Exit(-1)
	}
	err = log.Setup(LocalConfig.LogPath, LocalConfig.LogLevel)
	if err != nil {
		fmt.Println("2222", err)
		os.Exit(-1)
	}

	proposeC := make(chan string)
	defer close(proposeC)
	confChangeC := make(chan raftpb.ConfChange)
	defer close(confChangeC)

	// raft provides a commit stream for the proposals from the http api
	var kvs *kvstore
	getSnapshot := func() ([]byte, error) { return kvs.getSnapshot() }
	commitC, errorC, snapshotterReady := newRaftNode(LocalConfig.ID, strings.Split(LocalConfig.Peers, ","), LocalConfig.Join, getSnapshot, proposeC, confChangeC)

	kvs = newKVStore(<-snapshotterReady, proposeC, commitC, errorC)

	// the key-value http handler will propose updates to raft
	serveHttpKVAPI(kvs, LocalConfig.Listen, confChangeC, errorC)
}
