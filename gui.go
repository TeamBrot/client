package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Gui struct {
	conn *websocket.Conn
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func StartGui(logger *log.Logger) *Gui {
	gui := &Gui{nil}
	http.HandleFunc("/spe_ed/gui", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("could not accept gui connection:", err)
			return
		}
		logger.Println("gui connected")
		gui.conn = c
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "gui.html")
	})

	go func() {
		logger.Println(http.ListenAndServe(guiURL, nil))
	}()
	return gui
}

func (g *Gui) WriteStatus(status *JSONStatus) error {
	if g.conn != nil {
		return g.conn.WriteJSON(status)
	}
	return nil
}

func (g *Gui) Close() {
	if g.conn != nil {
		g.conn.Close()
	}
}
