package rest

import (
	"net/http"

	"github.com/Lambels/relationer/internal/service"
	"github.com/go-chi/chi/v5"
)

type HandlerService struct {
	store  service.GraphStore
	broker service.MessageBroker
}

func NewHandlerService(store service.GraphStore, broker service.MessageBroker) *HandlerService {
	return &HandlerService{
		store:  store,
		broker: broker,
	}
}

func (h *HandlerService) RegisterRouter(mux chi.Router) {
	// people
	mux.Get("/people", h.getAll)
	mux.Post("/people", h.addPerson)
	mux.Get("/people/{id}", h.getPerson)
	mux.Delete("/people/{id}", h.removePerson)

	// friendship
	mux.Post("/frienship", h.addFriendship)
	mux.Get("/friendship/{id}", h.getFriendship)
	mux.Get("/friendship/depth/{id1}/{id2}", h.getDepth)
}

func (h *HandlerService) addPerson(w http.ResponseWriter, r *http.Request) {

}

func (h *HandlerService) addFriendship(w http.ResponseWriter, r *http.Request) {

}

func (h *HandlerService) removePerson(w http.ResponseWriter, r *http.Request) {

}

func (h *HandlerService) getDepth(w http.ResponseWriter, r *http.Request) {

}

func (h *HandlerService) getFriendship(w http.ResponseWriter, r *http.Request) {

}

func (h *HandlerService) getPerson(w http.ResponseWriter, r *http.Request) {

}

func (h *HandlerService) getAll(w http.ResponseWriter, r *http.Request) {

}
