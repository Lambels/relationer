package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Lambels/relationer/internal"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
)

type DB struct {
	db *sql.DB

	DSN string

	Now func() time.Time
}

type Tx struct {
	*sql.Tx
	now time.Time
}

func NewDB(dsn string) *DB {
	db := &DB{
		DSN: dsn,
		Now: time.Now,
	}

	return db
}

func (db *DB) DB() *sql.DB {
	return db.db
}

func (db *DB) Open() (err error) {
	if db.DSN == "" {
		return errors.New("dsn required")
	}

	if db.db, err = sql.Open("postgres", db.DSN); err != nil {
		return err
	}

	// check db with ping.
	if err := db.db.Ping(); err != nil {
		return err
	}

	return nil
}

func (db *DB) Close() error {
	return db.db.Close()
}

func (db *DB) BeginTX(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &Tx{
		Tx:  tx,
		now: db.Now().UTC().Truncate(time.Second),
	}, nil
}

func parsePostgreErr(in error) error {
	var err *pq.Error
	if !errors.As(in, &err) {
		return internal.WrapErrorNil(in, internal.EINTERNAL, "couldnt parse error")
	}

	switch err.Code {
	case pgerrcode.CaseNotFound:
		return internal.WrapError(err, internal.ENOTFOUND, err.Error())
	case pgerrcode.UniqueViolation:
		return internal.WrapError(err, internal.ECONFLICT, err.Error())
	default: // TODO: more error code cheking
		return internal.WrapError(err, internal.EINTERNAL, err.Error())
	}
}
