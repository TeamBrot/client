#!/bin/bash
client="combi"
logdir="log/api/$client"

cd "$(dirname "$0")/../client"
mkdir -p "$logdir"
go build .


while [ ! -f ../stop ]
do
    now=$(date +%s)
    ./client -client "$client" -log "$logdir" >> "$logdir/$now-output.txt" 2>> "$logdir/$now-error.txt"
done
