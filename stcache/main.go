package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
)

type stCached struct {
	hs   *httpServer
	opts *options
	log  *log.Logger
	cm   *cacheManager
	raft *raftNodeInfo
}

type stCachedContext struct {
	st *stCached
}

func main() {
	st := &stCached{
		opts: NewOptions(),
		log:  log.New(os.Stderr, "stCached: ", log.Ldate|log.Ltime),
		cm:   NewCacheManager(),
	}
	ctx := &stCachedContext{st}

	var l net.Listener
	var err error
	l, err = net.Listen("tcp", st.opts.httpAddress)
	if err != nil {
		st.log.Fatal(fmt.Sprintf("listen %s failed: %s", st.opts.httpAddress, err))
	}
	st.log.Printf("http server listen:%s", l.Addr())

	logger := log.New(os.Stderr, "httpserver: ", log.Ldate|log.Ltime)
	httpServer := NewHttpServer(ctx, logger)
	st.hs = httpServer
	go func() {
		http.Serve(l, httpServer.mux)
	}()

	raft, err := newRaftNode(st.opts, ctx)
	if err != nil {
		st.log.Fatal(fmt.Sprintf("new raft node failed:%v", err))
	}
	st.raft = raft

	if st.opts.joinAddress != "" {
		err = joinRaftCluster(st.opts)
		if err != nil {
			st.log.Fatal(fmt.Sprintf("join raft cluster failed:%v", err))
		}
	}

	// monitor leadership
	for {
		select {
		case leader := <-st.raft.leaderNotifyCh:
			if leader {
				st.log.Println("become leader, enable write api")
				st.hs.setWriteFlag(true)
			} else {
				st.log.Println("become follower, close write api")
				st.hs.setWriteFlag(false)
			}
		}
	}
}
