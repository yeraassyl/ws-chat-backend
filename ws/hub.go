package ws

type message struct {
	data     []byte
	receiver string
}

type subscription struct {
	username string
	client   *Client
}

type Hub struct {
	clients map[string]*Client

	broadcast chan message

	register chan subscription

	unregister chan subscription
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan message),
		register:   make(chan subscription),
		unregister: make(chan subscription),
		clients:    make(map[string]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case sub := <-h.register:
			if _, ok := h.clients[sub.username]; !ok {
				h.clients[sub.username] = sub.client
			}
		case sub := <-h.unregister:
			if _, ok := h.clients[sub.username]; ok {
				delete(h.clients, sub.username)
				close(sub.client.send)
			}
		case message := <-h.broadcast:
			receiver := h.clients[message.receiver]
			select {
			case receiver.send <- message.data:
			default:
				close(receiver.send)
				delete(h.clients, message.receiver)
			}
		}
	}
}
