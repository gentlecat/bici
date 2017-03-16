package main

import (
	"go.roman.zone/bici/server"
	"go.roman.zone/bici/storage"
	"go.roman.zone/bici/strava/activity"
	"log"
	"net/http"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	check(storage.EstablishConnection())
	go activity.ActivityDetailsRetriever()
	go activity.AthleteRetriever()
	server.StartServer()
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
