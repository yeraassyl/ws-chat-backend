package ws

import (
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = (pongWait * 9) / 10
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	hub *Hub

	conn *websocket.Conn

	send chan []byte
}

func (s subscription) writePump() {
	ticker := time.NewTicker(pingPeriod)
	conn := s.client.conn

	defer func() {
		ticker.Stop()
		s.client.hub.unregister <- s
		conn.Close()
	}()

	for {
		select {
		case message, ok := <-s.client.send:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(s.client.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-s.client.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func SendMsg(hub *Hub, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	receiver := chi.URLParam(r, "username")
	if _, ok := hub.clients[receiver]; !ok {
		http.Error(w, http.StatusText(404), http.StatusNotFound)
		return
	}

	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Can't read body", http.StatusBadRequest)
		return
	}
	msg := message{
		data:     b,
		receiver: receiver,
	}
	hub.broadcast <- msg
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub, conn, make(chan []byte, 256)}
	sub := subscription{username, client}
	client.hub.register <- sub

	go sub.writePump()
}
