package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"chat/internal/auth"
	"chat/internal/hub"
	"chat/internal/model"
	"chat/internal/store"

	"github.com/coder/websocket"
)

const maxNameLen = 20

func validateName(name string) string {
	if name == "" {
		return "ник не может быть пустым"
	}
	if utf8.RuneCountInString(name) > maxNameLen {
		return "ник не длиннее 20 символов"
	}
	if strings.Contains(name, ",") {
		return "в нике нельзя использовать запятую"
	}
	return ""
}

func validateRoom(roomName string) string {
	if strings.Contains(roomName, ",") {
		return "некорректное имя комнаты"
	}
	return ""
}

func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func WSHandler(h *hub.Hub, st *store.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("кто-то пытается подключиться")

		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "нужен вход", http.StatusUnauthorized)
			return
		}

		_, name, err := st.SessionUser(cookie.Value)
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "нужен вход", http.StatusUnauthorized)
			return
		}
		if err != nil {
			log.Println("ошибка сессии:", err)
			http.Error(w, "внутренняя ошибка", http.StatusInternalServerError)
			return
		}

		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			log.Println("ошибка Accept:", err)
			return
		}

		roomName := strings.TrimSpace(r.URL.Query().Get("room"))
		if roomName == "" {
			roomName = "general"
		}

		if msgErr := validateName(name); msgErr != "" {
			log.Println("отклонено:", msgErr)
			conn.Write(r.Context(), websocket.MessageText, []byte("ОШИБКА: "+msgErr))
			conn.Close(websocket.StatusPolicyViolation, msgErr)
			return
		}

		if msgErr := validateRoom(roomName); msgErr != "" {
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

func RegisterHandler(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		password := r.FormValue("password")

		if msgErr := validateName(name); msgErr != "" {
			http.Error(w, msgErr, http.StatusBadRequest)
			return
		}
		if len(password) < 6 {
			http.Error(w, "пароль должен быть не короче 6 символов", http.StatusBadRequest)
			return
		}

		hash, err := auth.HashPassword(password)
		if err != nil {
			log.Println("ошибка хэширования:", err)
			http.Error(w, "внутренняя ошибка", http.StatusInternalServerError)
			return
		}

		_, err = st.CreateUser(name, hash)
		if err != nil {
			if strings.Contains(err.Error(), "unique") {
				http.Error(w, "ник уже занят", http.StatusConflict)
				return
			}
			log.Println("ошибка создания пользователя:", err)
			http.Error(w, "внутренняя ошибка", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func LoginHandler(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		password := r.FormValue("password")

		id, hash, err := st.UserByName(name)
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "неверный ник или пароль", http.StatusUnauthorized)
			return
		}
		if err != nil {
			log.Println("ошибка при входе:", err)
			http.Error(w, "внутренняя ошибка", http.StatusInternalServerError)
			return
		}
		if !auth.CheckPassword(hash, password) {
			http.Error(w, "неверный ник или пароль", http.StatusUnauthorized)
			return
		}

		token, err := newToken()
		if err != nil {
			log.Println("ошибка генерации токена:", err)
			http.Error(w, "внутренняя ошибка", http.StatusInternalServerError)
			return
		}

		expiresAt := time.Now().Add(24 * time.Hour)
		if err := st.CreateSession(token, id, expiresAt); err != nil {
			log.Println("ошибка создания сессии:", err)
			http.Error(w, "внутренняя ошибка", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Expires:  expiresAt,
		})
		w.WriteHeader(http.StatusOK)
	}
}
