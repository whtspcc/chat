package handlers

import (
	"log"
	"net/http"
	"strings"
	"unicode/utf8"

	"chat/internal/hub"
	"chat/internal/model"

	"github.com/coder/websocket"
)

const maxNameLen = 20

func validate(name, roomName string) string {
	if utf8.RuneCountInString(name) > maxNameLen {
		return "ник не длиннее 20 символов"
	}
	if strings.Contains(name, ",") {
		return "в нике нельзя использовать запятую"
	}
	if strings.Contains(roomName, ",") {
		return "некорректное имя комнаты"
	}
	return ""
}

func WSHandler(h *hub.Hub) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("кто-то пытается подключиться")

		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			log.Println("ошибка Accept:", err)
			return
		}

		name := strings.TrimSpace(r.URL.Query().Get("name"))
		if name == "" {
			name = "incognito"
		}

		roomName := strings.TrimSpace(r.URL.Query().Get("room"))
		if roomName == "" {
			roomName = "general"
		}

		if msgErr := validate(name, roomName); msgErr != "" {
			log.Println("отклонено:", msgErr)
			conn.Write(r.Context(), websocket.MessageText, []byte("ОШИБКА: "+msgErr))
			conn.Close(websocket.StatusPolicyViolation, msgErr)
			return
		}

		log.Printf("подключение: name=%s room=%s", name, roomName)

		client, err := h.AddClient(name, roomName)
		if err != nil {
			log.Println("не удалось добавить клиента:", err)
			conn.Close(websocket.StatusPolicyViolation, "комната не найдена")
			return
		}
		defer h.RemoveClient(client)

		go func() {
			for msg := range client.Send {
				if err := conn.Write(r.Context(), websocket.MessageText, []byte(msg)); err != nil {
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

			h.Broadcast(model.Message{
				Client: client,
				Text:   string(msg),
			})
		}
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/index.html")
}
