#!/usr/bin/bash

killall server
for i in {0..1000}
do
    echo "Spiele Spiel Nummer $i"
    ../server/server &
    pid=$!
    echo "Der Server wurde gestartet"
    sleep 1
    ./client smart &
    ./client smart 
    sleep 1
    kill $pid
    sleep  1
done
