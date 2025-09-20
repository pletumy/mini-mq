package store

import (
	"database/sql"
	"fmt"

	"github.com/pletumy/mini-mq/internal/broker"
)

type PGMessageStore struct {
	db         *sql.DB
	topicStore *PGTopicStore
}

func NewPGMessageStore(db *sql.DB, ts *PGTopicStore) *PGMessageStore {
	return &PGMessageStore{
		db:         db,
		topicStore: ts,
	}
}

func (s *PGMessageStore) SaveMessage(topicID int, msg broker.Message) error {
	if topicID == 0 {
		var err error
		topicID, err = s.topicStore.EnsureTopic(msg.Topic)
		if err != nil {
			return fmt.Errorf("ensure topic: %w", err)
		}
	}

	_, err := s.db.Exec(
		`INSERT INTO messages (id, topic_id, payload, ts)
		VALUES ($1, $2, $3, $4)`,
		msg.ID, topicID, msg.Payload, msg.TS,
	)
	if err != nil {
		return fmt.Errorf("insert message failed: %w", err)
	}
	return nil
}

func (s *PGMessageStore) GetHistory(topic string, limit int) ([]broker.Message, error) {
	topicID, err := s.topicStore.EnsureTopic(topic)
	if err != nil {
		return nil, fmt.Errorf("ensure topic: %w", err)
	}

	rows, err := s.db.Query(
		`SELECT id, payload, ts
		FROM messages
		WHERE topic_id = $1
		ORDER BY ts DESC
		LIMIT $2`, topicID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query history failed: %w", err)
	}

	defer rows.Close()

	msgs := []broker.Message{}
	for rows.Next() {
		var m broker.Message
		m.Topic = topic
		if err := rows.Scan(&m.ID, &m.Payload, &m.TS); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}

	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}

	return msgs, nil
}
