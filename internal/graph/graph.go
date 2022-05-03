package graph

import (
	"context"
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
func (s *GraphStoreService) Load() error {
	if s == nil || s.repo == nil {
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

	s.addPerson(person)
	return nil
}

func (s *GraphStoreService) AddFriendship(ctx context.Context, friendship internal.Friendship) error {
	if len(friendship.With) != 1 {
		return internal.Errorf(internal.ECONFLICT, "provided friendship should only be with one person")
	}

	// possibly skip table scan.
	if depth := s.getDepth(friendship.P1.ID, friendship.With[0].ID); depth != -1 {
		return internal.Errorf(internal.ECONFLICT, "%v and %v are already friends", friendship.P1, friendship.With[0])
	}

	if err := s.repo.AddFriendship(ctx, friendship); err != nil {
		return err
	}

	s.addFriendship(friendship.P1, friendship.With[0])
	return nil
}

func (s *GraphStoreService) RemovePerson(ctx context.Context, id int64) error {
	if _, err := s.getPerson(id); err != nil {
		return err
	}

	if err := s.repo.RemovePerson(ctx, id); err != nil {
		return err
	}

	s.mu.RLock()
	friends := s.edges[id]
	s.mu.RUnlock()
	for _, friend := range friends {
		s.removeFriendship(id, friend.ID) // unlink everyone linked with current person.
	}
	s.removePerson(id)
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
	s.mu.Lock()
	s.nodes = append(s.nodes, p)
	s.mu.Unlock()
}

func (s *GraphStoreService) addFriendship(p1, p2 *internal.Person) {
	s.mu.Lock()
	s.mu.Unlock()
}

func (s *GraphStoreService) getDepth(first, target int64) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok1 := s.edges[first]
	_, ok2 := s.edges[target]
	if ok1 != true || ok2 != true {
		return -1
	}

	queue := make([]int64, 0)
	visited := make(map[int64]bool)
	queue = append(queue, first) // start from first and look for target.

	var count int
	for {
		if len(queue) == 0 {
			return -1
		}

		currID := queue[0]      // seek first item in queue.
		queue = queue[1:]       // dequeue first item in queue.
		visited[currID] = true  // mark id as visited.
		near := s.edges[currID] // get near nodes for enqueueing.

		for i := 0; i < len(near); i++ {
			j := near[i]
			if !visited[j.ID] {
				queue = append(queue, j.ID)
			}
		}

		count++
		if currID == target {
			return count
		}
	}
}

func (s *GraphStoreService) getPerson(id int64) (*internal.Person, error) {

}

func (s *GraphStoreService) removeFriendship(p1, p2 int64) {
	s.mu.Lock()
	s.mu.Unlock()

}

func (s *GraphStoreService) removePerson(id int64) {
	s.mu.RLock()
	s.mu.RUnlock()
}
