package main

import (
	"flag"
)

type options struct {
	dataDir        string // data directory
	httpAddress    string // http server address
	raftTCPAddress string // construct Raft Address
	bootstrap      bool   // start as master or not
	joinAddress    string // peer address to join
}

func NewOptions() *options {
	opts := &options{}

	var httpAddress = flag.String("http", "127.0.0.1:6000", "Http address")
	var raftTCPAddress = flag.String("raft", "127.0.0.1:7000", "raft tcp address")
	var node = flag.String("node", "node1", "raft node name")
	var bootstrap = flag.Bool("bootstrap", false, "start as raft cluster")
	var joinAddress = flag.String("join", "", "join address for raft cluster")
	flag.Parse()

	opts.dataDir = "./" + *node
	opts.httpAddress = *httpAddress
	opts.bootstrap = *bootstrap
	opts.raftTCPAddress = *raftTCPAddress
	opts.joinAddress = *joinAddress
	return opts
}
