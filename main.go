package main

import (
	"log"
	"net/http"

	"chat/internal/handlers"
	"chat/internal/hub"
	"chat/internal/server"
)

func main() {
	broadcast := make(chan string)
	h := hub.NewHub(broadcast)

	go hub.Run(h)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", handlers.WSHandler(h))
	mux.HandleFunc("/home", handlers.HomeHandler)

	err := server.RunServer(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
