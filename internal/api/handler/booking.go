package handler

import (
	"encoding/json"
	"errors"
	"evently/internal/api/middleware"
	"evently/internal/service"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type BookingHandler struct {
	bookingService *service.BookingService
}

func NewBookingHandler(bookingService *service.BookingService) *BookingHandler {
	return &BookingHandler{bookingService: bookingService}
}

func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "id")
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "unauthorized", "Invalid user context")
		return
	}

	var input struct {
		Quantity int `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid_request", "Could not decode request body")
		return
	}

	if input.Quantity <= 0 {
		RespondWithError(w, http.StatusBadRequest, "invalid_quantity", "Quantity must be greater than zero")
		return
	}

	err := h.bookingService.CreateBooking(r.Context(), eventID, userID, input.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAddedToWaitlist):
			RespondWithJSON(w, http.StatusAccepted, map[string]string{"status": err.Error()})
		case errors.Is(err, service.ErrEventSoldOut):
			RespondWithError(w, http.StatusConflict, "sold_out", err.Error())
		case errors.Is(err, service.ErrBookingConflict):
			RespondWithError(w, http.StatusConflict, "booking_conflict", err.Error())
		default:
			RespondWithError(w, http.StatusInternalServerError, "server_error", "Could not create booking")
		}
		return
	}

	RespondWithJSON(w, http.StatusCreated, map[string]interface{}{"status": "booking created", "quantity": input.Quantity})
}

func (h *BookingHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	bookingID := chi.URLParam(r, "id")
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "unauthorized", "Invalid user context")
		return
	}

	err := h.bookingService.CancelBooking(r.Context(), bookingID, userID)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
