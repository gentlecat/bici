package main

import (
	"go.roman.zone/cazador/server"
	"go.roman.zone/cazador/storage"
	"go.roman.zone/cazador/strava/activity"
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
