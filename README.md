## Docker

Bauen:
`docker build . -t spe_ed`

Ausführen:
`docker run -e URL="wss://msoll.de/spe_ed" -e KEY="<key>" spe_ed`

## Python

Für client.py:
`pip install websockets`


## Run on remote Server
For Running on a remote Server you have to set the Environment variables URL, TIME_URL and KEY 
For the official api URL has to be wss://msoll.de/spe_ed and TIME_URL has to be https://msoll.de/spe_ed_time
## Go Clients

Für go clients:
`go build .` ausführen
anschließend können drei bots ausgeführt werden
- minimax
- left
- right
- smart
- speku

im Format `./clients <bot strategy>`

**TODOS**
- [ ] Eine util.go anlegen in der aller shared Code liegt
- [ ] Code duplikate entfernen und auslagern
- [ ] speku.go fertigstellen
    - [ ] Berechnung des Fields nutzen
    - [ ] Berechnung des Fields optimieren (Geschwindigkeit der Berechnung, Gleichzeitigkeit der Züge, Ausgabe sobald ein Turn abgeschlossen ist)
    - [ ] optimale parameter finden
    - [ ] Test schreiben, um zu überprüfen was passiert
    - [ ] Verschiedene Parameter in Abhängigkeit der Eingabe nutzen
- [ ] Dokumentation aufteilen und schreiben
- [ ] Spiele auf der offiziellen api sichtbar machen
- [ ] Logging entwickeln um erfolg auswerten zu können
- [ ] Docker image erstellen





