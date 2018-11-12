#!/usr/bin/env bash

while getopts ":r:c:n:e:" opt; do
  case ${opt} in
    r ) REPO=$OPTARG && echo "Using source from $REPO" 
      ;;
    c ) CHECKOUT=$OPTARG && echo "Checking out $CHECKOUT"
      ;;
    n ) NODES=$OPTARG && echo "Starting $NODES Swarm nodes"
      ;;
    e ) ENS=$OPTARG && echo "Using $ENS for ENS API"
      ;;
    \? ) echo "Usage: devcluster [-r git repo url] [-c branch, tag, or commit to checkout] [-n number of swarm nodes to start] [-e ens-api] [-h help]"
      ;;
  esac
done

cd /app
git clone $REPO
git checkout $CHECKOUT

cd go-ethereum

make geth
make swarm

sudo cp build/bin/geth /app/bin
sudo cp build/bin/swarm /app/bin

DATADIR=/app

if [[ ! -e $DATADIR/keystore ]]; then
    echo "fry-sauce" >> $DATADIR/password
    /app/bin/geth  --datadir $DATADIR account new --password $DATADIR/password
fi

nohup /app/bin/geth --syncmode light --rpc --rpcport 8545 --rpccorsdomain "*" --rpcvhosts "*" --rpcaddr 0.0.0.0 --rpcapi ws,db,eth,net,web3,personal,admin &

KEY=$(jq --raw-output '.address' $DATADIR/keystore/*)

/app/bin/swarm \
    --datadir $DATADIR \
    --password $DATADIR/password \
    --verbosity 5 \
    --bzzaccount $KEY \
    --httpaddr 0.0.0.0 \
    --nodiscover \
    --ens-api "" \
    --debug

