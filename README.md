Für client.py:
`pip install websockets`

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





