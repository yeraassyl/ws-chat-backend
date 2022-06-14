package main

import (
	"chat/ws"
	"flag"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":8080", "http service address")

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func main() {
	flag.Parse()
	hub := ws.NewHub()
	go hub.Run()
	r := chi.NewRouter()
	r.HandleFunc("/", serveHome)
	r.HandleFunc("/{username}", func(w http.ResponseWriter, r *http.Request) {
		ws.SendMsg(hub, w, r)
	})
	r.HandleFunc("/ws/{username}", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r)
	})
	err := http.ListenAndServe(*addr, r)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
