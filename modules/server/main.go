package main

import (
	"42Leisure/server/db"
	"42Leisure/server/ttt"
	"flag"
	"log"
	"net/http"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	db.InitDB()

	ttt.LoadGames()

	http.HandleFunc("/ttt", ttt.TTT)

	log.Fatal(http.ListenAndServe(*addr, nil))
}
