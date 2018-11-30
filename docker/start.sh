#!/usr/bin/env bash

while getopts ":r:c:e:" opt; do
  case ${opt} in
    r ) REPO=$OPTARG && echo "Using source from $REPO" 
      ;;
    c ) CHECKOUT=$OPTARG && echo "Checking out $CHECKOUT"
      ;;
#    n ) NODES=$OPTARG && echo "Starting $NODES Swarm nodes"
#      ;;
    e ) ENS=$OPTARG && echo "Using $ENS for ENS API"
      ;;
    \? ) echo "Usage: devcluster [-r git repo url] [-c branch, tag, or commit to checkout] [-n number of swarm nodes to start] [-e ens-api] [-h help]"
      ;;
  esac
done

cd /app
git clone $REPO
cd go-ethereum
git checkout $CHECKOUT

make geth
make swarm

sudo cp build/bin/geth /app/bin
sudo cp build/bin/swarm /app/bin

DATADIR=/app

if [[ ! -e $DATADIR/keystore ]]; then
    echo "fry-sauce" >> $DATADIR/password
    /app/bin/geth  --datadir $DATADIR account new --password $DATADIR/password
fi

nohup /app/bin/geth --syncmode light \
    --rpc \
    --rpcport 8545 \
    --rpcaddr 0.0.0.0 \
    --rpcapi 'admin,db,eth' \
    --rpcvhosts "*" \
    --bootnodes 'enode://e010178fe6d6bbf280348492ce58bb4d139ad40ad6421365dbad1614f06dd48382d110f191456f637d7afb00cb11a4f287471a7b484ebf031d79223c1c10d8d9@18.219.144.15:30303' &

KEY=$(jq --raw-output '.address' $DATADIR/keystore/*)

/app/bin/swarm \
    --datadir $DATADIR \
    --password $DATADIR/password \
    --verbosity 5 \
    --bzzaccount $KEY \
    --httpaddr 0.0.0.0 \
    --ens-api $ENS \
    --debug \
    --ws \
    --wsaddr 0.0.0.0 \
    --wsorigins "*"

tail -f /dev/null