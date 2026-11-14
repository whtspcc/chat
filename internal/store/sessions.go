package store

import "time"

func (s *Store) CreateSession(token string, userID int, expiresAt time.Time) error {
	_, err := s.db.Exec(`
	INSERT INTO sessions (token, user_id, expires_at)
	VALUES ($1, $2, $3)`, token, userID, expiresAt)

	return err
}

func (s *Store) SessionUser(token string) (int, string, error) {
	var id int
	var name string

	err := s.db.QueryRow(`
	SELECT u.id, u.name
	FROM sessions s
	JOIN users u ON u.id = s.user_id
	WHERE s.token = $1
	  AND s.expires_at > now()`, token).Scan(&id, &name)

	if err != nil {
		return 0, "", err
	}
	return id, name, nil
}
