package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

type raftNodeInfo struct {
	raft           *raft.Raft
	fsm            *FSM
	leaderNotifyCh chan bool
}

func newRaftTransport(opts *options) (*raft.NetworkTransport, error) {
	address, err := net.ResolveTCPAddr("tcp", opts.raftTCPAddress)
	if err != nil {
		return nil, err
	}
	transport, err := raft.NewTCPTransport(address.String(), address, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, err
	}
	return transport, nil
}

func newRaftNode(opts *options, ctx *stCachedContext) (*raftNodeInfo, error) {
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(opts.raftTCPAddress)
	raftConfig.Logger = log.New(os.Stderr, "raft: ", log.Ldate|log.Ltime)
	raftConfig.SnapshotInterval = 20 * time.Second
	raftConfig.SnapshotThreshold = 2
	leaderNotifyCh := make(chan bool, 1)
	raftConfig.NotifyCh = leaderNotifyCh

	transport, err := newRaftTransport(opts)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(opts.dataDir, 0700); err != nil {
		return nil, err
	}

	fsm := &FSM{
		ctx: ctx,
		log: log.New(os.Stderr, "FSM: ", log.Ldate|log.Ltime),
	}
	snapshotStore, err := raft.NewFileSnapshotStore(opts.dataDir, 1, os.Stderr)
	if err != nil {
		return nil, err
	}

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(opts.dataDir, "raft-log.bolt"))
	if err != nil {
		return nil, err
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(opts.dataDir, "raft-stable.bolt"))
	if err != nil {
		return nil, err
	}

	raftNode, err := raft.NewRaft(raftConfig, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return nil, err
	}

	if opts.bootstrap {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raftConfig.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		raftNode.BootstrapCluster(configuration)
	}

	return &raftNodeInfo{raft: raftNode, fsm: fsm, leaderNotifyCh: leaderNotifyCh}, nil
}

// joinRaftCluster joins a node to raft cluster
func joinRaftCluster(opts *options) error {
	url := fmt.Sprintf("http://%s/join?peerAddress=%s", opts.joinAddress, opts.raftTCPAddress)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if string(body) != "ok" {
		return errors.New(fmt.Sprintf("Error joining cluster: %s", body))
	}

	return nil
}
