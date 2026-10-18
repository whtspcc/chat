package hub

import (
	"chat/internal/store"

	"fmt"
	"log"
	"sync"
)

type Hub struct {
	clients       map[*Client]struct{}
	clientsMu     sync.Mutex
	broadcast     chan Message
	confirmations map[*Client]struct{}
	store         *store.Store
}

func NewHub(st *store.Store) *Hub {
	return &Hub{
		clients:       make(map[*Client]struct{}),
		confirmations: make(map[*Client]struct{}),
		broadcast:     make(chan Message),
		store:         st,
	}
}

func (h *Hub) Broadcast(msg Message) {
	h.broadcast <- msg
}

func (h *Hub) AddClient(name string) *Client {
	client := &Client{
		Name: name,
		Send: make(chan string, 256),
	}

	rows, err := h.store.LastMessages(10)

	if err != nil {
		log.Println("не удалось загрузить историю", err)
	} else {
		for _, row := range rows {
			client.Send <- fmt.Sprintf("%s | %s", row.Author, row.Text)
		}
	}

	h.clientsMu.Lock()
	h.clients[client] = struct{}{}
	h.clientsMu.Unlock()

	h.broadcastToAll("CMD:user_joined:" + name)

	return client
}

func (h *Hub) RemoveClient(client *Client) {
	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	delete(h.clients, client)
	delete(h.confirmations, client)

	h.broadcastToAllLocked("CMD:user_left:" + client.Name)

	close(client.Send)
}

func (h *Hub) Run() {
	for msg := range h.broadcast {
		switch msg.Text {

		case "CMD:confirm_yes":
			h.handleConfirmation(msg.Client)

		case "CMD:show_confirm":
			h.broadcastToAll("CMD:show_confirm")

		default:
			row, err := h.store.SaveMessage(msg.Client.Name, msg.Text)
			if err != nil {
				log.Println("не удалось сохранить сообщение:", err)
				continue
			}
			h.broadcastToAll(fmt.Sprintf("%s | %s", row.Author, row.Text))
		}
	}
}

func (h *Hub) broadcastToAll(msg string) {
	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	h.broadcastToAllLocked(msg)
}

func (h *Hub) broadcastToAllLocked(msg string) {
	for client := range h.clients {
		select {
		case client.Send <- msg:
		default:
			log.Printf("клиент %s не успевает получить сообщение", client.Name)
		}
	}
}

func (h *Hub) handleConfirmation(client *Client) {
	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	h.confirmations[client] = struct{}{}

	if len(h.confirmations) != len(h.clients) {
		return
	}

	clear(h.confirmations)

	h.broadcastToAllLocked("CMD:cleared")
}
