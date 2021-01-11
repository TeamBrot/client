#!/bin/bash
client="combi"
logdir="log/api/$client"

cd "$(dirname "$0")/../client"
touch error.txt
mkdir -p "$logdir"
go build .

while [ ! -f ../stop ]
do
    ./client -client "$client" -log "$logdir" 2>> error.txt
done
