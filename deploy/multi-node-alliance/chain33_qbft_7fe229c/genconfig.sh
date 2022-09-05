#!/bin/bash

function genConfig() {
    echo -e "start auto modify config"

    peerArray=$1
    num=$2

    validatorNodes=""
    seeds=""
    for((i=0;i<$num;i++));
    do
	validatorNodes=${validatorNodes},\""${peerArray[$i]}:33001"\"
        seeds=${seeds},"\"${peerArray[$i]}:13802\""
    done
    validatorNodes=$(echo $validatorNodes|sed '/^,/s///g')
    seeds=$(echo $seeds|sed '/^,/s///g')

    sed -i "s/^validatorNodes=.*/validatorNodes=[${validatorNodes}]/g" chain33.toml
    sed -i "s/^seeds=.*/seeds=[${seeds}]/g" chain33.toml
    for((i=0;i<$num;i++));
    do
	rm -rf ${peerArray[$i]}
        mkdir -p ${peerArray[$i]}
        cp chain33.toml  ${peerArray[$i]}/chain33.toml
    done

}

function prepareChain33() {
    echo -e "start prepare chain33 files"
    
    peerArray=$1
    num=$2

    ./chain33-cli qbft  gen_file -n $num -t bls

    for((i=0;i<$num;i++));
    do
        cp priv_validator_$i.json  ${peerArray[$i]}/priv_validator.json
        cp genesis_file.json  ${peerArray[$i]}/genesis.json
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

