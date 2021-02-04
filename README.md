# stcache

A simple cache server showing how to use hashicorp/raft

# doc

https://zhuanlan.zhihu.com/p/58048906

# build

```bash
make build
```

# start

## start leader node1

```bash
./run 1 1
```

## start follower node2

```bash
./run 2 0
```

## start follower node3

```bash
./run 3 0
```