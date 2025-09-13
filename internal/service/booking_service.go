package service

import (
	"context"
	"errors"
	"evently/internal/data"
	"log/slog"
)

var (
	ErrEventSoldOut    = errors.New("event is sold out")
	ErrAlreadyBooked   = errors.New("user has already booked this event")
	ErrBookingConflict = errors.New("booking conflict, please try again")
	MaxRetries         = 3
)

type BookingRepo interface {
	GetEventForUpdate(ctx context.Context, eventID string) (*data.EventForUpdate, error)
	UpdateEventAndCreateBooking(ctx context.Context, event *data.EventForUpdate, userID string) error
}

type BookingService struct {
	repo BookingRepo
	log  *slog.Logger
}

func NewBookingService(repo BookingRepo, log *slog.Logger) *BookingService {
	return &BookingService{repo: repo, log: log}
}

func (s *BookingService) CreateBooking(ctx context.Context, eventID, userID string) error {
	for i := 0; i < MaxRetries; i++ {
		event, err := s.repo.GetEventForUpdate(ctx, eventID)
		if err != nil {
			return err
		}

		if event.BookedTickets >= event.Capacity {
			return ErrEventSoldOut
		}

		err = s.repo.UpdateEventAndCreateBooking(ctx, event, userID)
		if err == nil {
			s.log.Info("booking successful", "user_id", userID, "event_id", eventID)
			return nil
		}

		if errors.Is(err, data.ErrDuplicate) {
			return ErrAlreadyBooked
		}

		if errors.Is(err, data.ErrConflict) {
			s.log.Warn("booking conflict detected, retrying...", "attempt", i+1, "event_id", eventID)
			continue
		}

		return err
	}

	return ErrBookingConflict
}
