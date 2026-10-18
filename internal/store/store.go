package store

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Store struct {
	db *sql.DB
}

func (s *Store) Close() error {
	return s.db.Close()
}

func Open(dsn string) (*Store, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	var dbName, schema string
	if err := db.QueryRow("SELECT current_database(), current_schema()").Scan(&dbName, &schema); err != nil {
		log.Println("не удалось узнать базу:", err)
	} else {
		log.Println("подключён к:", dbName, "схема:", schema)
	}
	return &Store{db: db}, nil
}
