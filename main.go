package main

import (
	"executor/api"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	server := api.NewServer("localhost:13666")
	log.Fatal(server.Run())
}
