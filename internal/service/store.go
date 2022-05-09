package service

import (
	"context"

	"github.com/Lambels/relationer/internal"
)

// Store stores people and their many-to-many relations
// represented by friendships.
type Store interface {
	AddPerson(context.Context, *internal.Person) error

	AddFriendship(context.Context, internal.Friendship) error

	RemovePerson(context.Context, int64) error
}

// GraphStore is a bi-directional graph ds representing
// relation-ships between people.
//
// Should be used to interface with persistent store and provide extra functionality over typicall
// store.
type GraphStore interface {
	Store

	GetDepth(context.Context, int64, int64) (int, error)

	GetFriendship(context.Context, int64) (internal.Friendship, error)

	GetPerson(context.Context, int64) (*internal.Person, error)

	GetAll(context.Context) []internal.Friendship
}
