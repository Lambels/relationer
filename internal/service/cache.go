package service

import (
	"context"
	"time"
)

type Cache interface {
	Set(context.Context, string, interface{}, time.Duration) error
	Delete(context.Context, string) error
}
