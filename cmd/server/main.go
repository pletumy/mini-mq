package main

import (
	"log"
	"net/http"

	"github.com/pletumy/mini-mq/internal/api"
	"github.com/pletumy/mini-mq/internal/broker"
)

func main() {
	b := broker.NewBroker()

	// init api
	mux := http.NewServeMux()
	api.RegisterRoutes(mux, b)

	log.Println("miniMQ running on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
