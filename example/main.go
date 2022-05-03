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

package example

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"go.etcd.io/etcd/raft/v3/raftpb"
)

func main() {
	baseName := os.Getenv("node_base_name")
	nodesCnt, _ := strconv.Atoi(os.Getenv("nodes"))
	clusterArr := make([]string, nodesCnt)
	for i := 0; i < nodesCnt; i++ {
		clusterArr[i] = fmt.Sprintf("http://%s-%d:9021", baseName, i+1)
	}
	cluster := strings.Join(clusterArr, ",")
	id, _ := strconv.Atoi(os.Getenv("node_id"))
	kvport, _ := strconv.Atoi(os.Getenv("port"))
	//kvport := flag.Int("port", 8080, "key-value server port")
	join := flag.Bool("join", false, "join an existing cluster")
	flag.Parse()

	proposeC := make(chan string)
	defer close(proposeC)
	confChangeC := make(chan raftpb.ConfChange)
	defer close(confChangeC)

	// raft provides a commit stream for the proposals from the http api
	var kvs *kvstore
	getSnapshot := func() ([]byte, error) { return kvs.getSnapshot() }
	commitC, errorC, snapshotterReady := newRaftNode(id, strings.Split(cluster, ","), *join, getSnapshot, proposeC, confChangeC)

	kvs = newKVStore(<-snapshotterReady, proposeC, commitC, errorC)

	// the key-value http handler will propose updates to raft
	serveHttpKVAPI(kvs, kvport, confChangeC, errorC)
}
