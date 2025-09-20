package store

import (
	"database/sql"
	"fmt"
)

type PGTopicStore struct {
	db *sql.DB
}

func NewPGTopicStore(db *sql.DB) *PGTopicStore {
	return &PGTopicStore{db: db}
}

// EnsureTopic: nếu topic chưa có thì insert, return topicID
func (s *PGTopicStore) EnsureTopic(name string) (int, error) {
	var id int
	err := s.db.QueryRow(`SELECT id FROM topics WHERE name = $1`, name).Scan(&id)
	if err == sql.ErrNoRows {
		err = s.db.QueryRow(
			`INSERT INTO topics (names) VALUES ($1) RETURNING id`, name,
		).Scan(&id)
		if err != nil {
			return 0, fmt.Errorf("insert topic failed: %w", err)
		}
		return id, nil
	} else if err != nil {
		return 0, fmt.Errorf("query topic failed: %w", err)
	}
	return id, nil
}

func (s *PGTopicStore) ListTopics() ([]string, error) {
	rows, err := s.db.Query(`SELECT name FROM topics ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list topics failed: %w", err)
	}
	defer rows.Close()

	topics := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		topics = append(topics, name)
	}
	return topics, nil
}
