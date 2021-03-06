package graph

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/Lambels/relationer/internal"
	"github.com/Lambels/relationer/internal/service"
)

// Store is a bi-directional graph ds representing
// relation-ships between people.
type GraphStoreService struct {
	repo  service.Store
	cache service.Cache
	db    *sql.DB

	// graph properties.
	nodes []*internal.Person
	edges map[int64][]int64

	once sync.Once
	mu   sync.RWMutex
}

// New initializes a new store.
func NewGraphStore(db *sql.DB, repo service.Store, cache service.Cache) *GraphStoreService {
	return &GraphStoreService{
		repo:  repo,
		cache: cache,
		db:    db,
		nodes: make([]*internal.Person, 0),
		edges: make(map[int64][]int64),
	}
}

// Load, syncs the store with the database.
//
// should only be used once after initialization.
func (s *GraphStoreService) Load(ctx context.Context) error {
	if s == nil || s.repo == nil {
		return internal.Errorf(internal.EINTERNAL, "store is nil")
	}
	if s.db == nil {
		return internal.Errorf(internal.EINTERNAL, "db is nil")
	}

	// load data from db, full table scan.
	var doErr error
	s.once.Do(func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		// load relationships.
		rows, err := s.db.QueryContext(ctx, `
			SELECT people.id, people.name, people.created_at, friendships.person2_id FROM people
			FULL JOIN friendships ON people.id = friendships.person1_id
			ORDER BY 1`,
		)
		if err != nil {
			doErr = err
			return
		}
		defer rows.Close()

		people := make([]*internal.Person, 1) // skip if check on each itteration
		people[0] = &internal.Person{ID: -1}
		relations := make(map[int64][]int64, 0)
		for rows.Next() {
			var person internal.Person
			var friendID sql.NullInt64

			if err := rows.Scan(
				&person.ID,
				&person.Name,
				&person.CreatedAt,
				&friendID,
			); err != nil {
				doErr = err
				return
			}

			if people[len(people)-1].ID != person.ID {
				people = append(people, &person)
			}

			if friendID.Valid {
				relations[person.ID] = append(relations[person.ID], friendID.Int64)
			}
		}

		s.nodes = people
		s.edges = relations
		doErr = nil
	})
	return doErr
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

	if err := s.repo.AddFriendship(ctx, friendship); err != nil {
		return err
	}

	s.mu.Lock()
	s.addFriendship(friendship.P1.ID, friendship.With[0])
	s.mu.Unlock()
	return nil
}

func (s *GraphStoreService) RemovePerson(ctx context.Context, id int64) error {
	if _, err := s.getPerson(id); err != nil {
		return err
	}

	if err := s.repo.RemovePerson(ctx, id); err != nil {
		return err
	}

	s.mu.Lock()
	friends := s.edges[id]
	for _, id := range friends {
		s.removeFriendship(id, id) // unlink everyone linked with current person.
	}
	s.removePerson(id)
	s.mu.Unlock()
	return nil
}

func (s *GraphStoreService) GetPerson(ctx context.Context, id int64) (*internal.Person, error) {
	return s.getPerson(id)
}

// FindDepth uses bfs to find the depth distance between to people, if not related
// id will be -1.
//
// returns ENOTFOUND if one of the people arent found.
func (s *GraphStoreService) GetDepth(ctx context.Context, first, second int64) (int, error) {
	var res int

	// check cache. (1)
	if err := s.cache.Get(ctx, fmt.Sprintf("D%v%v", first, second), &res); err == nil {
		return res, nil
	}
	// check cache. (2)
	if err := s.cache.Get(ctx, fmt.Sprintf("D%v%v", second, first), &res); err == nil {
		return res, nil
	}

	// fetch depth.
	depth, err := s.getDepth(ctx, first, second)
	if err != nil {
		return depth, err
	}

	if err := s.cache.Set(ctx, fmt.Sprintf("D%v%v", first, second), depth, 5*time.Minute); err != nil {
		return depth, internal.WrapError(err, internal.EINTERNAL, "cache.Set") // wrap error easy to check for cache error.
	}

	return depth, nil
}

func (s *GraphStoreService) GetFriendship(ctx context.Context, id int64) (internal.Friendship, error) {
	var res internal.Friendship

	// search cache.
	if err := s.cache.Get(ctx, fmt.Sprintf("F%v", id), &res); err == nil {
		return res, nil
	}

	pers, err := s.getPerson(id)
	if err != nil {
		return res, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	friends := s.edges[pers.ID]

	res.P1 = pers
	res.With = friends

	// set cache.
	if err := s.cache.Set(ctx, fmt.Sprintf("F%v", id), res, 5*time.Second); err != nil {
		return res, internal.WrapError(err, internal.EINTERNAL, "cache.Set") // wrap error easy to check for cache error.
	}

	return res, nil
}

func (s *GraphStoreService) GetAll(ctx context.Context) ([]internal.Friendship, error) {
	return s.getAll(ctx)
}

func (s *GraphStoreService) addPerson(p *internal.Person) {
	s.nodes = append(s.nodes, p)
}

func (s *GraphStoreService) addFriendship(p1, p2 int64) {
	s.edges[p1] = append(s.edges[p1], p2)
}

func (s *GraphStoreService) getDepth(ctx context.Context, first, target int64) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok1 := s.edges[first]
	_, ok2 := s.edges[target]
	if ok1 != true || ok2 != true {
		return -1, internal.Errorf(internal.ENOTFOUND, "one of the ids provided doesent exist")
	}

	queue := make([]int64, 0)
	visited := make(map[int64]bool)
	queue = append(queue, first) // start from first and look for target.

	var count int
	for {
		select {
		case <-ctx.Done():
			return -1, internal.WrapError(ctx.Err(), internal.EINVALID, "ctx.Err")

		default:
		}

		if len(queue) == 0 {
			return -1, internal.Errorf(internal.ENOTFOUND, "target wasnt found in any relationship connection")
		}

		currID := queue[0]      // seek first item in queue.
		queue = queue[1:]       // dequeue first item in queue.
		visited[currID] = true  // mark id as visited.
		near := s.edges[currID] // get near nodes for enqueueing.

		for i := 0; i < len(near); i++ {
			j := near[i]
			if !visited[j] {
				queue = append(queue, j)
			}
		}

		count++
		if currID == target {
			return count, nil
		}
	}
}

func (s *GraphStoreService) getPerson(id int64) (*internal.Person, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, person := range s.nodes {
		if person.ID == id {
			return person, nil
		}
	}

	return nil, internal.Errorf(internal.ENOTFOUND, "person not found")
}

func (s *GraphStoreService) removeFriendship(p1, p2 int64) {
	friends := s.edges[p1]
	for i, friend := range friends {
		if friend == p2 {
			friends[i] = friends[len(friends)-1]
			friends = friends[:len(friends)-1]
			break
		}
	}

	s.edges[p1] = friends
}

func (s *GraphStoreService) removePerson(id int64) {
	for i, pers := range s.nodes {
		if pers.ID == id {
			s.nodes[i] = s.nodes[len(s.nodes)-1]
			s.nodes = s.nodes[:len(s.nodes)-1]
			break
		}
	}
}

func (s *GraphStoreService) getAll(ctx context.Context) ([]internal.Friendship, error) {
	friendships := make([]internal.Friendship, 0)

	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, person := range s.nodes {
		select {
		case <-ctx.Done():
			return friendships, internal.WrapError(ctx.Err(), internal.EINVALID, "ctx.Err")
		default:
		}

		friends := s.edges[person.ID]
		friendships = append(friendships, internal.Friendship{
			P1:   person,
			With: friends,
		})
	}

	return friendships, nil
}
