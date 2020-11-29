#!/usr/bin/bash

killall server
for i in {0..100000}
do
    echo "Spiele Spiel Nummer $i"
    ../server/server  &> /dev/null &
    spid=$!
    echo "Der Server wurde gestartet"
    ./qlearningclient.py &> /dev/null & 
    cpid=$!
    ./client smart &> /dev/null &
    c2pid=$!
    wait $cpid
    killall server
done
