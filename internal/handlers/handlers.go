package handlers

import (
	"log"
	"net/http"

	"chat/internal/hub"

	"github.com/coder/websocket"
)

func WSHandler(h *hub.Hub) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("кто-то пытается подключиться")

		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			log.Println("ошибка Accept:", err)
			return
		}

		name := r.URL.Query().Get("name")
		if name == "" {
			name = "incognito"
		}

		client := h.AddClient(name)
		defer h.RemoveClient(client)

		go func() {
			for msg := range client.Send {
				err := conn.Write(r.Context(), websocket.MessageText, []byte(msg))
				if err != nil {
					log.Println("ошибка Write:", err)
				}

			}
		}()

		for {
			_, msg, err := conn.Read(r.Context())
			if err != nil {
				log.Println("соединение закрыто", err)
				return
			}

			text := string(msg)

			h.Broadcast(hub.Message{
				Client: client,
				Text:   text,
			})
		}
	}
}
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/index.html")
}
