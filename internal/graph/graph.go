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

	// graph properties.
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

func (s *GraphStoreService) AddPerson(ctx context.Context, person *internal.Person) error {
	// add to persistent store to generate id.
	if err := s.repo.AddPerson(ctx, person); err != nil {
		return err
	}

	s.mu.Lock()
	s.addPerson(person)
	s.mu.Unlock()
	return nil
}

func (s *GraphStoreService) AddFriendship(ctx context.Context, friendship internal.Friendship) error {
	if len(friendship.With) != 1 {
		return internal.Errorf(internal.ECONFLICT, "provided friendship should only be with one person")
	}

	// possibly skip table scan.
	s.mu.RLock()
	if depth := s.getDepth(friendship.P1.ID, friendship.With[0].ID); depth != -1 {
		return internal.Errorf(internal.ECONFLICT, "%v and %v are already friends", friendship.P1, friendship.With[0])
	}
	s.mu.RUnlock()

	if err := s.repo.AddFriendship(ctx, friendship); err != nil {
		return err
	}

	s.mu.Lock()
	s.addFriendship(friendship.P1, friendship.With[0])
	s.mu.Unlock()
	return nil
}

func (s *GraphStoreService) RemovePerson(ctx context.Context, id int64) error {
	s.mu.RLock()
	if _, err := s.getPerson(id); err != nil {
		return err
	}
	s.mu.RUnlock()

	if err := s.repo.RemovePerson(ctx, id); err != nil {
		return err
	}

	s.mu.Lock()
	friends := s.edges[id]
	for _, friend := range friends {
		s.removeFriendship(id, friend.ID) // unlink everyone linked with current person.
	}
	s.removePerson(id)
	s.mu.Unlock()
	return nil
}

func (s *GraphStoreService) GetPerson(ctx context.Context, id int64) (internal.Friendship, error) {

}

// FindDepth uses bfs to find the depth distance between to people, if not related
// id will be -1.
//
// returns ENOTFOUND if one of the people arent found.
func (s *GraphStoreService) GetDepth(ctx context.Context, first, second int64) (int, error) {

}

func (s *GraphStoreService) GetFriendship(ctx context.Context, id int64) (internal.Friendship, error) {

}

func (s *GraphStoreService) addPerson(p *internal.Person) {
	s.nodes = append(s.nodes, p)
}

func (s *GraphStoreService) addFriendship(p1, p2 *internal.Person) {

}

func (s *GraphStoreService) getDepth(first, second int64) int {

}

func (s *GraphStoreService) getPerson(id int64) (*internal.Person, error) {

}

func (s *GraphStoreService) removeFriendship(p1, p2 int64) {

}

func (s *GraphStoreService) removePerson(id int64) {

}
