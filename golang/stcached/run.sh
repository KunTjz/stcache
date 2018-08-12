#!/usr/bin/env bash

node_number=$1
is_master=$2

function usage() {
    echo "Usage:"
    echo "   run.sh nodeNumber(1-5)"
}

if ! [[ "${node_number}" =~ ^[1-5]$ ]] ; then
    usage
    exit 1
fi

node_raft=$((6999 + $node_number))
node_http=$((5999 + $node_number))
node="./node${node_number}"

args="${args} --http=127.0.0.1:${node_http}"
args="${args} --raft=127.0.0.1:${node_raft}"
args="${args} --node=${node}"

if [ "${is_master}" == 1 ] ; then
    args="${args} --bootstrap true"
else
    args="${args} --join=127.0.0.1:6000"
fi

echo${args}
./stcached ${args}

