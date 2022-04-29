package rest

import "github.com/Lambels/relationer/internal/service"

type HandlerService struct {
	gStore service.GraphStore
	dStore service.PostgreStore
	cache  service.Cache
}
