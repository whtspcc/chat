package main

import (
	"log"
	"net/http"
	"os"

	"chat/internal/handlers"
	"chat/internal/hub"
	"chat/internal/server"
	"chat/internal/store"
)

func main() {
	dsn := os.Getenv("CHAT_DB_DSN")
	if dsn == "" {
		log.Fatal("не задана переменная CHAT_DB_DSN")
	}

	st, err := store.Open(dsn)
	if err != nil {
		log.Fatal(err)
	}

	defer st.Close()

	h := hub.NewHub(st)

	go h.Run()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", handlers.WSHandler(h))
	mux.HandleFunc("/home", handlers.HomeHandler)

	err = server.RunServer(":8081", mux)
	if err != nil {
		log.Fatal(err)
	}
}
