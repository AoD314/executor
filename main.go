package main

import (
	"executor/api"
	"log"

	_ "github.com/mattn/go-sqlite3"
	_ "net/http/pprof" // важно: подключаем pprof
)

func main() {
	server := api.NewServer("localhost:13666")
	log.Fatal(server.Run())
}
