#!/bin/bash

function genConfig() {
    echo -e "start auto modify config"

    peerArray=$1
    num=$2

    peersURL=""
    seeds=""
    for((i=0;i<$num;i++));
    do
	peersURL=${peersURL},"http:\/\/${peerArray[$i]}:9021"
        seeds=${seeds},"\"${peerArray[$i]}:13802\""
    done
    peersURL=$(echo $peersURL|sed '/^,/s///g')
    seeds=$(echo $seeds|sed '/^,/s///g')

    sed -i "s/^peersURL=.*/peersURL=\"${peersURL}\"/g" chain33.raft.toml
    sed -i "s/^seeds=.*/seeds=[${seeds}]/g" chain33.raft.toml
    sed -i 's/^isSeed=.*/isSeed=true/g' chain33.raft.toml
    for((i=0;i<$num;i++));
    do
	rm -rf ${peerArray[$i]}
        mkdir -p ${peerArray[$i]}
	sed -i "s/^nodeID=.*/nodeID=$[$i+1]/g" chain33.raft.toml
        cp chain33.raft.toml  ${peerArray[$i]}/chain33.raft.toml
    done

}

function prepareChain33() {
    echo -e "start prepare chain33 files"
    
    peerArray=$1
    num=$2

    for((i=0;i<$num;i++));
    do
	cp chain33 ${peerArray[$i]}/chain33
	cp chain33-cli ${peerArray[$i]}/chain33-cli
    done

}

function main() {
    peers=$1

    peerArray=(${peers//,/ })

    peerNum=${#peerArray[*]}

    genConfig $peerArray $peerNum

    prepareChain33 $peerArray $peerNum
}

main $1

