package noop

import (
	"context"

	"github.com/Lambels/relationer/internal"
)

type NoopStore struct{}

func NewNoopStore() NoopStore {
	return NoopStore{}
}

func (s NoopStore) AddPerson(context.Context, *internal.Person) error {
	return nil
}

func (s NoopStore) AddFriendship(context.Context, internal.Friendship) error {
	return nil
}

func (s NoopStore) RemovePerson(context.Context, int64) error {
	return nil
}
