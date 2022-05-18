package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Lambels/relationer/internal"
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

	// friendship
	mux.Post("/friendship", h.addFriendship)
	mux.Get("/friendship/depth/{id1}/{id2}", h.getDepth)

	// id to use parser middleware.
	mux.Route("/", func(r chi.Router) {
		r.Use(idContextValidator)

		r.Get("/people/{id}", h.getPerson)
		r.Delete("/people/{id}", h.removePerson)
		r.Get("/friendship/{id}", h.getFriendship)
	})
}

func (h *HandlerService) addPerson(w http.ResponseWriter, r *http.Request) {
	var person *internal.Person
	if err := json.NewDecoder(r.Body).Decode(&person); err != nil {
		sendErrorResponse(w, internal.WrapError(err, internal.ECONFLICT, "invalid json body"))
		return
	}

	if err := h.store.AddPerson(r.Context(), person); err != nil {
		sendErrorResponse(w, err)
		return
	}

	if err := h.broker.CreatedPerson(r.Context(), person); err != nil {
		sendErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *HandlerService) addFriendship(w http.ResponseWriter, r *http.Request) {
	var friendship internal.Friendship
	if err := json.NewDecoder(r.Body).Decode(&friendship); err != nil {
		sendErrorResponse(w, internal.WrapError(err, internal.ECONFLICT, "invalid json body"))
		return
	}

	if err := h.store.AddFriendship(r.Context(), friendship); err != nil {
		sendErrorResponse(w, err)
		return
	}

	if err := h.broker.CreatedFriendship(r.Context(), friendship); err != nil {
		sendErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *HandlerService) removePerson(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(idKey{}).(int64)

	if err := h.store.RemovePerson(r.Context(), id); err != nil {
		sendErrorResponse(w, err)
		return
	}

	if err := h.broker.DeletedPerson(r.Context(), id); err != nil {
		sendErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *HandlerService) getFriendship(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(idKey{}).(int64)

	friendship, err := h.store.GetFriendship(r.Context(), id)
	if err != nil {
		sendErrorResponse(w, err)
		return
	}

	sendResponse(w, friendship, http.StatusOK)
}

func (h *HandlerService) getPerson(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(idKey{}).(int64)

	person, err := h.store.GetPerson(r.Context(), id)
	if err != nil {
		sendErrorResponse(w, err)
		return
	}

	sendResponse(w, person, http.StatusOK)
}

func (h *HandlerService) getAll(w http.ResponseWriter, r *http.Request) {
	data, err := h.store.GetAll(r.Context())
	if err != nil {
		sendErrorResponse(w, err)
		return
	}

	sendResponse(w, data, http.StatusOK)
}

func (h *HandlerService) getDepth(w http.ResponseWriter, r *http.Request) {
	id1, err := strconv.Atoi(chi.URLParam(r, "id1"))
	if err != nil {
		sendErrorResponse(w, internal.Errorf(internal.ECONFLICT, "invalid id1"))
		return
	}
	id2, err := strconv.Atoi(chi.URLParam(r, "id2"))
	if err != nil {
		sendErrorResponse(w, internal.Errorf(internal.ECONFLICT, "invalid id2"))
		return
	}

	depth, err := h.store.GetDepth(r.Context(), int64(id1), int64(id2))
	if err != nil {
		sendErrorResponse(w, err)
		return
	}

	sendResponse(w, GetDepthResponse{Depth: depth}, http.StatusOK)
}

func idContextValidator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			sendErrorResponse(w, internal.Errorf(internal.ECONFLICT, "invalid id"))
			return
		}

		ctx := context.WithValue(r.Context(), idKey{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
