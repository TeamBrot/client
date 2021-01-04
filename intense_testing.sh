#!/bin/bash
baseerror="error"
baselog="log/nightlyTests"
baseoutput="output"
probabilities=( "0.2" "0.3" "0.4" "0.5" "0.6" "0.7" "0.8" "0.9" "1.0" "1.2" "1.3" "1.4" "1.6" "1.7" "1.8" "1.9" "2.0")
activations=("0.005" "0.001" "0.0005" "0.00001" "0.000005")
clients=("smart" "minimax" "rollouts" "probability")
numPlayers=("2" "3" "4" "5" "6")
lengths=("10" "15" "25" "35" "40" "50" "70" "100")
offsets=("10" "5" "1" "15")
deadlines=("1" "3" "5")

cd server
go build .
echo "build server..."
cd ../client
go build .
echo "build client..."
cd ..
while [ ! -f ../stop ]
do
    echo "start new"
    pid=()
    cd server
    players=${numPlayers[$RANDOM % ${#numPlayers[@]} ]}
    echo "using players $players"
    height=${lengths[$RANDOM % ${#lengths[@]} ]}
    echo "height $height"
    width=${lengths[$RANDOM % ${#lengths[@]} ]}
    echo "width $width"
    deadline=${deadlines[$RANDOM % ${#deadlines[@]} ]}
    echo "minimal deadline $deadline"
    offset=${offsets[$RANDOM % ${#offsets[@]} ]}
    echo "offset $offset"
    ./server -p "$players" -h "$height" -w "$width" -d "$deadline" -o "$offset" &> "/dev/null" &
    sleep 0.2
    cd ../client
    for((i=1; i<=$players;i++))
    do
        if (($i == 1))
        then 
            client="combi"
            probability=${probabilities[$RANDOM % ${#probabilities[@]} ]}
            minimax=${activations[$RANDOM % ${#activations[@]} ]}
            errordir="$baseerror/$client/$probability/$minimax"
            logdir="$baselog/$client/$probability/$minimax"
            outputdir="$baseoutput/$client"
            mkdir -p "$logdir"
            mkdir -p "$errordir"
            mkdir -p "$outputdir"
            now=$( date +%s)
            echo "starting first client"
            ./client -client "$client" -log "$logdir" -probability "$probability" -activation "$minimax"   >> "$outputdir/outputdir$now.txt" 2>> "$errordir/error$now.txt" &
            pids[${i}]=$!
        else
            client=${clients[$RANDOM % ${#clients[@]}]}
            if [ "$client" = "combi" ];
            then
                probability=${probabilities[$RANDOM % ${#probabilities[@]} ]}
                minimax=${activations[$RANDOM % ${#activations[@]} ]}
                errordir="error/$client/$probability/$minimax"
                logdir="$baselog/$client/$probability/$minimax"
            else
                probability="0"
                minimax="0"
                logdir="$baselog/$client"
                errordir="$baseerror/$client"
            fi
            now=$( date +%s)
            outputdir="$baseoutput/$client"
            mkdir -p "$outputdir"
            mkdir -p "$logdir"
            mkdir -p "$errordir"
            ./client -client "$client" -log "$logdir" -probability "$probability" -activation "$minimax"  >> "$outputdir/outputdir$now.txt"  2>> "$errordir/error$now.txt" &
            pids[${i}]=$!
        fi
    done 
    for pid in ${pids[*]}
    do
        echo "Waiting for $pid"
        wait $pid
    done
    killall server
    cd ..
done