#!/bin/bash
#Defines the base Error directory for logging purposes
#Hier bitte euren Namen eingeben!!
name=""
baseerror="error"
baselog="log/intense-testing/"
baseoutput="output"

#When starting a combi Client it takes one of those values as combis
probabilities=("1.0" "1.2" "1.4" "1.6" "1.8" "2.0" "2.2" "2.4")
activations=("0.05" "0.01" "0.005" "0.001" "0.0005" "0.00001")
clients=("combi")
numPlayers=("2")
lengths=("10" "15" "25" "35" "40" "50" "70")
offsets=("4")
deadlines=("2")
filterValues=("0.55" "0.6" "0.65" "0.7" "0.75" "0.8" "0.85")

cd server
go build .
echo "build server..."
cd ../client
go build .
echo "build client..."
cd ..
counter=0

while [ -d "client/$baselog/$name-game-$counter" ]
do
    counter=$((counter+1))
done
while [ ! -f ./stop ]
do
    echo "starting game $counter"
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
    logdir="$baselog/$name-game-$counter"
    errordir="$logdir/$baseerror"
    outputdir="$logdir/$baseoutput"
    
    mkdir -p "$logdir"
    mkdir -p "$errordir"
    mkdir -p "$outputdir"
    for((i=1; i<=$players;i++))
    do
        if (($i == 1))
        then
            client="combi"
            probability=${probabilities[$RANDOM % ${#probabilities[@]} ]}
            minimax=${activations[$RANDOM % ${#activations[@]} ]}
            filterValue=${filterValues[$RANDOM % ${#filterValues[@]}]}
            now=$( date +%s)
            echo "starting first client"
            echo -e "Game Info \n players: $players \n width: $width \n height: $height \n mindead: $deadline \n off: $offset \n $i $client \n $probability \n $minimax \n $filterValue\n" >> "$logdir/gameInfo.txt"
            ./client -client "$client" -log "$logdir" -filter "$filterValue" -probability "$probability" -activation "$minimax"   >> "$outputdir/$i-$client-output-$now.txt" 2>> "$errordir/$i-$client-error-$now.txt" &
            pids[${i}]=$!
        else
            client=${clients[$RANDOM % ${#clients[@]}]}
            if [ "$client" = "combi" ];
            then
                probability=${probabilities[$RANDOM % ${#probabilities[@]} ]}
                minimax=${activations[$RANDOM % ${#activations[@]} ]}
            else
                probability="0"
                minimax="0"
            fi
            filterValue=${filterValues[$RANDOM % ${#filterValues[@]}]}
            now=$( date +%s)
            echo -e "$i $client \n $probability \n $minimax \n $filterValue\n" >> "$logdir/gameInfo.txt"
            ./client -client "$client" -log "$logdir" -filter "$filterValue" -probability "$probability" -activation "$minimax"   >> "$outputdir/$i-$client-output-$now.txt" 2>> "$errordir/$i-$client-error-$now.txt" &
            pids[${i}]=$!
        fi
    done
    for pid in ${pids[*]}
    do
        echo "Waiting for $pid"
        wait $pid
    done
    counter=$((counter+1))
    killall server
    cd ..
done