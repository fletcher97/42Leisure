package coms

import (
	"log"

	"github.com/gorilla/websocket"
)

func Send(c *websocket.Conn, msg []byte) bool {
	err := c.WriteMessage(websocket.BinaryMessage, msg)
	if err != nil {
		log.Println("write:", err)
		return false
	}
	return true
}
