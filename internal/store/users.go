package store

func (s *Store) CreateUser(name, passwordHash string) (int, error) {
	var id int

	err := s.db.QueryRow(`
	INSERT INTO users (name, password_hash)
	VALUES ($1, $2)
	RETURNING id`, name, passwordHash).Scan(&id)

	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Store) UserByName(name string) (int, string, error) {
	var id int
	var hash string

	err := s.db.QueryRow(`
	SELECT id, password_hash
	FROM users
	WHERE name = $1`, name).Scan(&id, &hash)

	if err != nil {
		return 0, "", err
	}
	return id, hash, nil
}
