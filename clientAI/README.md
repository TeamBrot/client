Um gegen die KI zu spielen:

    `pip install websocket-client`
    `pip install tesorflow`
    evtl. noch andere Sachen...

Nach Spielende wird automatisch neu verbunden.
Am besten läuft der Server schon vor Ausführung.
`for ((i = 0 ; i < 10 ; i++)); do
        ./client minimax
done`