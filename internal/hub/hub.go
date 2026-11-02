package hub

import (
	"chat/internal/model"
	"chat/internal/store"
	"strings"

	"fmt"
	"log"
	"sync"
)

type Hub struct {
	clientsMu sync.Mutex

	broadcast     chan model.Message
	confirmations map[*model.Client]struct{}
	store         *store.Store

	rooms map[string]*Room
}

type Room struct {
	id      int
	clients map[*model.Client]struct{}
}

func NewHub(st *store.Store) *Hub {
	return &Hub{
		rooms:         make(map[string]*Room),
		confirmations: make(map[*model.Client]struct{}),
		broadcast:     make(chan model.Message),
		store:         st,
	}
}

func (h *Hub) Broadcast(msg model.Message) {
	h.broadcast <- msg
}

func (h *Hub) AddClient(name, roomName string) (*model.Client, error) {
	client := &model.Client{
		Name: name,
		Send: make(chan string, 256),
	}

	roomID, err := h.store.RoomID(roomName)
	if err != nil {
		return nil, fmt.Errorf("комната %q не найдена: %w", roomName, err)
	}

	h.clientsMu.Lock()
	room, ok := h.rooms[roomName]
	if !ok {
		room = &Room{id: roomID, clients: make(map[*model.Client]struct{})}
		h.rooms[roomName] = room
	}
	room.clients[client] = struct{}{}
	client.Room = roomName
	names := h.onlineNamesLocked(room)
	h.clientsMu.Unlock()

	rows, err := h.store.LastMessages(roomID, 10)

	if err != nil {
		log.Println("не удалось загрузить историю", err)
	} else {
		for _, row := range rows {
			client.Send <- fmt.Sprintf("%s | %s", row.Author, row.Text)
		}
	}
	h.broadcastToRoom(room, "CMD:user_joined:"+name)
	h.broadcastToRoom(room, "CMD:online:"+strings.Join(names, ","))

	return client, nil
}

func (h *Hub) RemoveClient(client *model.Client) {
	h.clientsMu.Lock()

	room, ok := h.rooms[client.Room]
	if ok {
		delete(room.clients, client)
	}
	delete(h.confirmations, client)

	var online []string
	if ok {
		online = h.onlineNamesLocked(room)
	}
	h.clientsMu.Unlock()

	if ok {
		h.broadcastToRoom(room, "CMD:user_left:"+client.Name)
		h.broadcastToRoom(room, "CMD:online:"+strings.Join(online, ","))
	}

	close(client.Send)
}

func (h *Hub) Run() {
	for msg := range h.broadcast {
		switch msg.Text {

		case "CMD:confirm_yes":
			h.handleConfirmation(msg.Client)

		case "CMD:show_confirm":
			room, ok := h.rooms[msg.Client.Room]
			if ok {
				h.broadcastToRoom(room, "CMD:show_confirm")
			}

		default:
			room, ok := h.rooms[msg.Client.Room]
			if !ok {
				log.Println("комната не найдена:", msg.Client.Room)
				continue
			}
			row, err := h.store.SaveMessage(msg.Client.Name, msg.Text, room.id)
			if err != nil {
				log.Println("не удалось сохранить сообщение: ", err)
				continue
			}
			h.broadcastToRoom(room, fmt.Sprintf("%s | %s", row.Author, row.Text))
		}
	}
}

func (h *Hub) broadcastToRoom(room *Room, msg string) {
	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	for client := range room.clients {
		select {
		case client.Send <- msg:
		default:
			log.Printf("клиент %s не успевает получить сообщение", client.Name)
		}
	}
}

func (h *Hub) handleConfirmation(client *model.Client) {
	h.clientsMu.Lock()

	room, ok := h.rooms[client.Room]
	if !ok {
		h.clientsMu.Unlock()
		return
	}

	h.confirmations[client] = struct{}{}

	if len(h.confirmations) != len(room.clients) {
		h.clientsMu.Unlock()
		return
	}

	clear(h.confirmations)
	h.clientsMu.Unlock()

	h.broadcastToRoom(room, "CMD:cleared")
}

func (h *Hub) onlineNamesLocked(room *Room) []string {
	names := []string{}
	for c := range room.clients {
		names = append(names, c.Name)
	}
	return names
}
