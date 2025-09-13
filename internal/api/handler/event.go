// package handler

// import (
// 	"evently/internal/data"
// 	"net/http"

// 	"github.com/go-chi/chi/v5"
// )

// type EventHandler struct {
// 	eventRepo *data.EventRepository
// }

// func NewEventHandler(eventRepo *data.EventRepository) *EventHandler {
// 	return &EventHandler{eventRepo: eventRepo}
// }

// func (h *EventHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
// 	events, err := h.eventRepo.GetAll(r.Context())
// 	if err != nil {
// 		RespondWithError(w, http.StatusInternalServerError, "server_error", "Could not fetch events")
// 		return
// 	}
// 	RespondWithJSON(w, http.StatusOK, events)
// }

// func (h *EventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
// 	id := chi.URLParam(r, "id")
// 	event, err := h.eventRepo.GetByID(r.Context(), id)
// 	if err != nil {
// 		RespondWithError(w, http.StatusNotFound, "not_found", "Event not found")
// 		return
// 	}
// 	RespondWithJSON(w, http.StatusOK, event)
// }

package handler

import (
	"encoding/json"
	"evently/internal/data"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type EventHandler struct {
	eventRepo *data.EventRepository
}

func NewEventHandler(eventRepo *data.EventRepository) *EventHandler {
	return &EventHandler{eventRepo: eventRepo}
}

func (h *EventHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.eventRepo.GetAll(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "server_error", "Could not fetch events")
		return
	}
	RespondWithJSON(w, http.StatusOK, events)
}

func (h *EventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	event, err := h.eventRepo.GetByID(r.Context(), id)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "not_found", "Event not found")
		return
	}
	RespondWithJSON(w, http.StatusOK, event)
}

func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name      string    `json:"name"`
		Venue     string    `json:"venue"`
		StartTime time.Time `json:"start_time"`
		Capacity  int       `json:"capacity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload")
		return
	}

	event := &data.Event{
		Name:      input.Name,
		Venue:     input.Venue,
		StartTime: input.StartTime,
		Capacity:  input.Capacity,
	}

	err := h.eventRepo.Create(r.Context(), event)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "server_error", "Could not create event")
		return
	}

	RespondWithJSON(w, http.StatusCreated, event)
}
