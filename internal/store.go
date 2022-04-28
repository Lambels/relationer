package internal

import (
	"database/sql"
	"errors"
	"sync"
)

type Store struct {
	nodes []*Person
	//edges

	once sync.Once
	mu   sync.RWMutex
}

func New() *Store {
	return &Store{}
}

func (s *Store) Load(db *sql.DB) error {
	if s == nil {
		return errors.New("store isnt initialized")
	}

	// load data from db, full table scan.
	s.once.Do(func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		// load many to many relationship from db.
	})
	return nil
}
