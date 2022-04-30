package graph

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"github.com/Lambels/relationer/internal"
	"github.com/Lambels/relationer/internal/service"
)

// Store is a bi-directional graph ds representing
// relation-ships between people.
type GraphStoreService struct {
	repo service.Store

	// graph properties
	nodes []*internal.Person
	edges map[int64][]*internal.Person

	once sync.Once
	mu   sync.RWMutex
}

// New initializes a new store.
func NewGraphStore() *GraphStoreService {
	return &GraphStoreService{
		nodes: make([]*internal.Person, 0),
		edges: make(map[int64][]*internal.Person),
	}
}

// Load, syncs the store with the database.
//
// should only be used once after initialization.
func (s *GraphStoreService) Load(db *sql.DB) error {
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

func (s *GraphStoreService) AddPerson(ctx context.Context, in *internal.Person) error {

}

func (s *GraphStoreService) AddFriendship(ctx context.Context, in internal.Friendship) error {

}

func (s *GraphStoreService) RemovePerson(ctx context.Context, in int64) error {

}

func (s *GraphStoreService) GetPerson(ctx context.Context, id int64) (internal.Friendship, error) {

}

// FindDepth uses bfs to find the depth distance between to people, if not related
// id will be -1.
//
// returns ENOTFOUND if one of the people arent found.
func (s *GraphStoreService) GetDepth(ctx context.Context, first, second int64) (int, error) {

}
