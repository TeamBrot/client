#!/bin/bash
client="combi"
logdir="log/api/$client"

cd client
mkdir -p "$logdir"
go build .

while [ ! -f ../stop ]
do
    ./client -client "$client" -log "$logdir"
done
