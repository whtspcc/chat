package model

import "github.com/gorilla/websocket"

type Client struct {
	Conn *websocket.Conn
	Send chan string
	Name string
	Room string
}
