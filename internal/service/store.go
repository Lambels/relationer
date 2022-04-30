package service

import (
	"context"

	"github.com/Lambels/relationer/internal"
)

type Store interface {
	AddPerson(context.Context, *internal.Person) error

	AddFriendship(context.Context, internal.Friendship) error

	RemovePerson(context.Context, int64) error

	GetPerson(context.Context, int64) (internal.Person, error)
}

type GraphStore interface {
	Store

	GetDepth(context.Context, int64, int64) (int, error)
}
