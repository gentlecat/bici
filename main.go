package main

import (
	"go.roman.zone/turbo-parakeet/server"
	"go.roman.zone/turbo-parakeet/storage"
	"log"
	"net/http"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	check(storage.EstablishConnection())
	server.StartServer()
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
