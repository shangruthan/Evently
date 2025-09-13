package handler

import (
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

	err := h.bookingService.CreateBooking(r.Context(), eventID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEventSoldOut):
			RespondWithError(w, http.StatusConflict, "sold_out", err.Error())
		case errors.Is(err, service.ErrAlreadyBooked):
			RespondWithError(w, http.StatusConflict, "already_booked", err.Error())
		case errors.Is(err, service.ErrBookingConflict):
			RespondWithError(w, http.StatusConflict, "booking_conflict", err.Error())
		default:
			RespondWithError(w, http.StatusInternalServerError, "server_error", "Could not create booking")
		}
		return
	}

	RespondWithJSON(w, http.StatusCreated, map[string]string{"status": "booking created"})
}
