#!/usr/bin/bash

killall server
for i in {0..1000}
do
    echo "Spiele Spiel Nummer $i"
    ../server/server &
    spid=$!
    echo "Der Server wurde gestartet"
    sleep 0.2
    ./client speku &
    cpid=$!
    ./client smart &
    c2pid=$!
    wait $cpid
    sleep 0.1
    kill $spid
    sleep  0.1
done
