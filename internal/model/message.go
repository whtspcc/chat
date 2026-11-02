package model

import "time"

type Message struct {
	ID        int
	Author    string
	Text      string
	Timestamp time.Time
	Client    *Client
}
