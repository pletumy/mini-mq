package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pletumy/mini-mq/internal/broker"
)

func RegisterRoutes(mux *http.ServeMux, b *broker.Broker) {
	mux.Handle("/", http.FileServer(http.Dir("./web")))

	mux.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Topic   string `json:"topic"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		msg := b.Publish(req.Topic, req.Message)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(msg)
	})

	mux.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
		topic := r.URL.Query().Get("topic")
		if topic == "" {
			http.Error(w, "topic required", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		msgCh := b.Subscribe(ctx, topic)

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		for msg := range msgCh {
			fmt.Fprintf(w, "data: %s\n\n", toJSON(msg))
			flusher.Flush()
		}
	})

	mux.HandleFunc("/topics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusBadRequest)
			return
		}

		topics := b.GetTopics()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(topics)
	})

	// Get message history for a specific topic
	mux.HandleFunc("/topics/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusBadRequest)
			return
		}

		// Extract topic from URL path like /topics/orders/messages
		path := r.URL.Path
		if len(path) < 8 { // minimum "/topics/"
			http.Error(w, "invalid topic path", http.StatusBadRequest)
			return
		}

		// Parse topic from path
		topicPath := path[8:] // remove "/topics/"
		if topicPath == "" {
			http.Error(w, "topic required", http.StatusBadRequest)
			return
		}

		// Check if it's a messages request
		if len(topicPath) > 9 && topicPath[len(topicPath)-9:] == "/messages" {
			topic := topicPath[:len(topicPath)-9] // remove "/messages"
			history := b.GetTopicHistory(topic)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(history)
			return
		}

		// For now, just return topic info
		http.Error(w, "endpoint not found", http.StatusNotFound)
	})

	// health
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
}

func toJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
