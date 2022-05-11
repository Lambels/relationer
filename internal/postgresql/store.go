package postgresql

import (
	"context"

	"github.com/Lambels/relationer/internal"
	"github.com/Lambels/relationer/internal/service"
)

// PostgreSqlStoreService
type PostgreSqlStoreService struct {
	cache service.Cache
	db    *DB
}

func NewPostgresqlStore(db *DB, cache service.Cache) *PostgreSqlStoreService {
	return &PostgreSqlStoreService{
		cache: cache,
		db:    db,
	}
}

// AddPerson
func (s *PostgreSqlStoreService) AddPerson(ctx context.Context, person *internal.Person) error {
	tx, err := s.db.BeginTX(ctx, nil)
	if err != nil {
		return internal.WrapError(err, internal.EINTERNAL, "db.BeginTX")
	}
	defer tx.Rollback()

	if err := addPerson(ctx, tx, person); err != nil {
		return parsePostgreErr(err)
	}

	return internal.WrapErrorNil(tx.Commit(), internal.EINTERNAL, "tx.Commit")
}

// AddFriendship
func (s *PostgreSqlStoreService) AddFriendship(ctx context.Context, friendship internal.Friendship) error {
	tx, err := s.db.BeginTX(ctx, nil)
	if err != nil {
		return internal.WrapError(err, internal.EINTERNAL, "db.BeginTX")
	}
	defer tx.Rollback()

	if err := addFriendship(ctx, tx, friendship); err != nil {
		return parsePostgreErr(err)
	}

	return internal.WrapErrorNil(tx.Commit(), internal.EINTERNAL, "tx.Commit")
}

// RemovePerson
func (s *PostgreSqlStoreService) RemovePerson(ctx context.Context, id int64) error {
	tx, err := s.db.BeginTX(ctx, nil)
	if err != nil {
		return internal.WrapError(err, internal.EINTERNAL, "db.BeginTX")
	}
	defer tx.Rollback()

	if err := removePerson(ctx, tx, id); err != nil {
		return parsePostgreErr(err)
	}

	return internal.WrapErrorNil(tx.Commit(), internal.EINTERNAL, "tx.Commit")
}

func addPerson(ctx context.Context, tx *Tx, person *internal.Person) error {

}

func addFriendship(ctx context.Context, tx *Tx, friendship internal.Friendship) error {

}

func removePerson(ctx context.Context, tx *Tx, id int64) error {

}
