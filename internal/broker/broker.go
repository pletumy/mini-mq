package broker

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

const historySize = 10

type topicData struct {
	subs    []chan Message
	history []Message
}

type Message struct {
	ID      string    `json:"id"`
	Topic   string    `json:"topic"`
	Payload string    `json:"payload"`
	TS      time.Time `json:"ts"`
}

type Broker struct {
	mu     sync.RWMutex
	topics map[string]*topicData
}

func NewBroker() *Broker {
	return &Broker{
		topics: make(map[string]*topicData),
	}
}

// tao message moi va gui den toan bo subcribers
func (b *Broker) Publish(topic, payload string) Message {
	msg := Message{
		ID:      uuid.NewString(),
		Topic:   topic,
		Payload: payload,
		TS:      time.Now(),
	}

	b.mu.Lock()
	td, ok := b.topics[topic]
	if !ok {
		td = &topicData{}
		b.topics[topic] = td
	}

	// append vào history
	td.history = append(td.history, msg)
	if len(td.history) > historySize {
		td.history = td.history[len(td.history)-historySize:]
	}

	for _, ch := range td.subs {
		go func(c chan Message) {
			select {
			case c <- msg:
			case <-time.After(time.Second):
			}
		}(ch)
	}
	b.mu.Unlock()

	return msg
}

// sucribers theo topics
func (b *Broker) Subscribe(ctx context.Context, topic string) <-chan Message {
	ch := make(chan Message, 10)

	b.mu.Lock()
	td, ok := b.topics[topic]
	if !ok {
		td = &topicData{}
		b.topics[topic] = td
	}

	// Add subscriber first
	td.subs = append(td.subs, ch)

	// Get history copy to avoid holding lock while sending
	history := make([]Message, len(td.history))
	copy(history, td.history)
	b.mu.Unlock()

	// Send history in a separate goroutine to avoid blocking
	go func() {
		for _, msg := range history {
			select {
			case ch <- msg:
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
				// Skip if can't send within timeout
				continue
			}
		}
	}()

	go func() {
		<-ctx.Done()
		b.removeSubscriber(topic, ch)
		close(ch)
	}()

	return ch
}

func (b *Broker) removeSubscriber(topic string, target chan Message) {
	b.mu.Lock()
	defer b.mu.Unlock()
	td, ok := b.topics[topic]
	if !ok {
		return
	}
	for i, ch := range td.subs {
		if ch == target {
			td.subs = append(td.subs[:i], td.subs[i+1:]...)
			break
		}
	}
}

func (b *Broker) GetTopics() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	topics := make([]string, 0, len(b.topics))
	for t := range b.topics {
		topics = append(topics, t)
	}
	return topics
}

func (b *Broker) GetTopicHistory(topic string) []Message {
	b.mu.RLock()
	defer b.mu.RUnlock()

	td, ok := b.topics[topic]
	if !ok {
		return []Message{}
	}

	// Return a copy to avoid race conditions
	history := make([]Message, len(td.history))
	copy(history, td.history)
	return history
}
