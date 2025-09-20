package main

import (
	"log"
	"net/http"

	"github.com/pletumy/mini-mq/internal/api"
	"github.com/pletumy/mini-mq/internal/broker"
	"github.com/pletumy/mini-mq/internal/config"
	"github.com/pletumy/mini-mq/internal/store"
)

func main() {
	cfg := config.Load()
	db, err := store.NewDB(cfg.DBUrl)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	defer db.Close()

	var messageStore store.MessageStore

	b := broker.NewBroker(messageStore)

	// init api
	mux := http.NewServeMux()
	api.RegisterRoutes(mux, b)

	log.Printf("Server listening on :%s", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, mux))
}
