package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"chat/internal/hub"

	"github.com/coder/websocket"
)

func WSHandler(h *hub.Hub) func(w http.ResponseWriter, r *http.Request) {
	var userCount = 0

	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("кто-то пытается подключиться")
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			log.Println("ошибка Accept:", err)
			return
		}

		send := make(chan string, 5)

		h.ClientsMu.Lock()
		h.Clients = append(h.Clients, send)
		userCount++
		name := fmt.Sprintf("user%d", userCount)
		h.ClientsMu.Unlock()

		go func() {
			for msg := range send {
				err := conn.Write(r.Context(), websocket.MessageText, []byte(msg))
				if err != nil {
					log.Println("ошибка Write:", err)
				}

			}
		}()

		for {
			_, msg, err := conn.Read(r.Context())
			if err != nil {
				log.Println("прочитал из сокета:", string(msg))
				h.ClientsMu.Lock()
				for i, c := range h.Clients {
					if c == send {
						h.Clients = append(h.Clients[:i], h.Clients[i+1:]...)
						break
					}
				}
				h.ClientsMu.Unlock()
				close(send)
				return
			}

			text := string(msg)
			if strings.HasPrefix(text, "CMD:") {
				h.Broadcast <- text
			} else {
				h.Broadcast <- fmt.Sprintf("%s | %s", name, text)
			}
		}
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/index.html")
}
