package main

import (
	"bufio"
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func send(c *websocket.Conn, msg []byte) bool {
	err := c.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		log.Println("write:", err)
		return false
	}
	return true
}

func awaitMsgLoop(c *websocket.Conn, resp chan []byte) {
	defer close(resp)
	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("AwaitMsgLoop-ReadMessage:", err)
			return
		}
		resp <- msg
	}
}

func terminate(c *websocket.Conn, serverMsg chan []byte, input chan string, stopInput chan bool) {
	close(stopInput)
	// Cleanly close the connection by sending a close message and then
	// waiting (with timeout) for the server to close the connection.
	err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close:", err)
		return
	}
}

func readInput(input chan string, stopInput chan bool) {
	defer close(input)
	defer func() { println("closed input") }()
	reader := bufio.NewReader(os.Stdin)
	for {
		select {
		case <-stopInput:
			println("clean stop")
			return
		default:
			txt, err := reader.ReadString('\n')
			if err != nil {
				log.Println("input:", err)
				return
			}
			input <- txt
		}
	}

}

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	serverMsg := make(chan []byte, 1)
	input := make(chan string, 1)
	stopInput := make(chan bool)

	go awaitMsgLoop(c, serverMsg)

	go readInput(input, stopInput)

	for {
		select {
		case msg, ok := <-input:
			if !ok {
				println("input closed")
				terminate(c, serverMsg, input, stopInput)
				return
			}
			send(c, []byte(msg))
		case msg, ok := <-serverMsg:
			if !ok {
				println("server closed")
				close(stopInput)
				return
			}
			println(string(msg), ok)
		case <-interrupt:
			log.Println("interrupt")
			terminate(c, serverMsg, input, stopInput)
			return
		}
	}
}
