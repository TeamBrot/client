#!/bin/bash
client="combi"
logdir="log/api/$client"

cd client
touch error.txt
mkdir -p "$logdir"
go build .

while [ ! -f ../stop ]
do
    ./client -client "$client" -log "$logdir" 2>> error.txt
done
