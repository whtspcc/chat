package main

import (
	"log"
	"net/http"

	"chat/internal/handlers"
	"chat/internal/hub"
	"chat/internal/server"
)

func main() {
	h := hub.NewHub()

	go h.Run()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", handlers.WSHandler(h))
	mux.HandleFunc("/home", handlers.HomeHandler)

	err := server.RunServer(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
