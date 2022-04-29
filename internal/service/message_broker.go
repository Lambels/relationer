package service

import (
	"context"

	"github.com/Lambels/relationer/internal"
)

type MessageBroker interface {
	CreatedPerson(context.Context, *internal.Person) error
	CreatedFriendship(context.Context, internal.Friendship) error
	DeletedPerson(context.Context, int64) error
}
