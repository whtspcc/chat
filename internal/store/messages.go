package store

import (
	"time"
)

type MessageRow struct {
	ID        int
	Author    string
	Text      string
	CreatedAt time.Time
}

func (s *Store) SaveMessage(author, text string) (MessageRow, error) {
	var m MessageRow

	err := s.db.QueryRow(`
	INSERT INTO messages (author, text)
	VALUES ($1, $2)
	RETURNING id, author, text, created_at`, author, text).Scan(&m.ID, &m.Author, &m.Text, &m.CreatedAt)

	return m, err
}

func (s *Store) LastMessages(limit int) ([]MessageRow, error) {
	rows, err := s.db.Query(`
		SELECT id, author, text, created_at
		FROM messages
		ORDER BY id DESC
		LIMIT $1
		`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []MessageRow
	for rows.Next() {
		var m MessageRow
		if err := rows.Scan(&m.ID, &m.Author, &m.Text, &m.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, m)
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result, rows.Err()
}
