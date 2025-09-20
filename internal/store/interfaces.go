package store

import (
	"github.com/pletumy/mini-mq/internal/model"
)

type TopicStore interface {
	EnsureTopic(name string) (int, error)
	ListTopics() ([]string, error)
}

type MessageStore interface {
	SaveMessage(topicID int, msg model.Message) error
	GetHistory(topic string, limit int) ([]model.Message, error)
}
