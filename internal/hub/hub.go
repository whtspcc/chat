package hub

import (
	"fmt"
	"log"
	"strings"
	"sync"
)

type Hub struct {
	Clients      []chan string
	ClientsMu    sync.Mutex
	Broadcast    chan string
	ConfirmCount int
}

func NewHub(br chan string) *Hub {
	return &Hub{
		Broadcast: br,
	}
}

func Run(h *Hub) {
	for msg := range h.Broadcast {
		log.Println("Run получил:", msg)
		switch {
		case msg == "CMD:confirm_yes":
			h.ClientsMu.Lock()
			h.ConfirmCount++
			ready := h.ConfirmCount == len(h.Clients)
			if ready {
				h.ConfirmCount = 0
			}
			h.ClientsMu.Unlock()

			if ready {
				h.ClientsMu.Lock()
				for i, v := range h.Clients {
					select {
					case v <- "CMD:cleared":
						log.Print("команда cleared отправлена: чат удаляется")
					default:
						log.Printf("клиент %d не успевает, пропускаю\n", i+1)
					}
				}
				h.ClientsMu.Unlock()
			}

		case strings.HasPrefix(msg, "CMD:"):
			h.ClientsMu.Lock()
			for i, v := range h.Clients {
				select {
				case v <- msg:
					log.Print("пришла команда с префиксом CMD")
				default:
					log.Printf("клиент %d не успевает, пропускаю\n", i+1)
				}
			}
			h.ClientsMu.Unlock()

		default:
			h.ClientsMu.Lock()
			for i, v := range h.Clients {
				select {
				case v <- msg:
					fmt.Println("отправили")
				default:
					fmt.Printf("клиент %d не успевает, пропускаю\n", i+1)
				}
			}
			h.ClientsMu.Unlock()
		}
	}
}
