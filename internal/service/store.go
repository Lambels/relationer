package service

import (
	"context"

	"github.com/Lambels/relationer/internal"
)

type PostgreStore interface {
	AddPerson(context.Context, *internal.Person) error

	AddFriendship(context.Context, internal.Friendship) error

	RemovePerson(context.Context, int64) error

	GetPerson(context.Context, int64) (internal.Person, error)
}

type GraphStore interface {
	PostgreStore

	GetDepth(context.Context, int64, int64) (int, error)
}
