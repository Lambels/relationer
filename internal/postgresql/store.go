package postgresql

import (
	"context"
	"database/sql"

	"github.com/Lambels/relationer/internal"
	"github.com/Lambels/relationer/internal/service"
)

// PostgreSqlStoreService
type PostgreSqlStoreService struct {
	cache service.Cache
	db    *sql.DB
}

// AddPerson
func (s *PostgreSqlStoreService) AddPerson(ctx context.Context, person *internal.Person) error {

}

// AddFriendship
func (s *PostgreSqlStoreService) AddFriendship(ctx context.Context, friendship internal.Friendship) error {

}

// RemovePerson
func (s *PostgreSqlStoreService) RemovePerson(ctx context.Context, id int64) error {

}

// GetPerson
func (s *PostgreSqlStoreService) GetPerson(ctx context.Context, id int64) (internal.Person, error) {

}
