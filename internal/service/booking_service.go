package service

import (
	"context"
	"errors"
	"evently/internal/data"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrEventSoldOut    = errors.New("not enough tickets available")
	ErrAlreadyBooked   = errors.New("user has already booked this event")
	ErrBookingConflict = errors.New("booking conflict, please try again")
	ErrAddedToWaitlist = errors.New("event is full, you have been added to the waitlist")
	MaxRetries         = 3
)

type BookingRepo interface {
	GetEventForUpdate(ctx context.Context, eventID string) (*data.EventForUpdate, error)
	UpdateEventAndCreateBooking(ctx context.Context, event *data.EventForUpdate, userID string, quantity int) error
	AddToWaitlist(ctx context.Context, eventID, userID string) error
	CancelBooking(ctx context.Context, bookingID, userID string) (*data.WaitlistUser, error)
}

type BookingService struct {
	repo BookingRepo
	log  *slog.Logger
}

func NewBookingService(repo BookingRepo, log *slog.Logger) *BookingService {
	return &BookingService{repo: repo, log: log}
}

func (s *BookingService) CreateBooking(ctx context.Context, eventID, userID string, quantity int) error {
	for i := 0; i < MaxRetries; i++ {
		event, err := s.repo.GetEventForUpdate(ctx, eventID)
		if err != nil {
			return err
		}

		if (event.BookedTickets + quantity) > event.Capacity {
			// Not enough tickets, so we can only offer to waitlist for a single spot
			if err := s.repo.AddToWaitlist(ctx, eventID, userID); err != nil {
				return err
			}
			return ErrAddedToWaitlist
		}

		err = s.repo.UpdateEventAndCreateBooking(ctx, event, userID, quantity)
		if err == nil {
			s.log.Info("booking successful", "user_id", userID, "event_id", eventID, "quantity", quantity)
			return nil
		}

		if errors.Is(err, data.ErrConflict) {
			s.log.Warn("booking conflict detected, retrying...", "attempt", i+1, "event_id", eventID)
			continue
		}

		return err
	}

	return ErrBookingConflict
}

func (s *BookingService) CancelBooking(ctx context.Context, bookingID, userID string) error {
	nextInLine, err := s.repo.CancelBooking(ctx, bookingID, userID)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return errors.New("booking not found or you do not have permission to cancel it")
		}
		return err
	}

	if nextInLine != nil {
		s.log.Info("auto-booking ticket for user from waitlist", "user_id", nextInLine.UserID, "user_email", nextInLine.Email)
	}

	return nil
}

// Implement the interface for the actual repository
type BookingRepositoryWithTx struct {
	DB *pgxpool.Pool
	*data.BookingRepository
}

func (r *BookingRepositoryWithTx) CancelBooking(ctx context.Context, bookingID, userID string) (*data.WaitlistUser, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	eventID, err := r.GetBookingForCancellation(ctx, tx, bookingID, userID)
	if err != nil {
		return nil, err
	}

	nextInLine, err := r.GetAndRemoveNextFromWaitlist(ctx, tx, eventID)
	if err != nil {
		return nil, err
	}

	if nextInLine != nil {
		if err := r.CreateBookingInTx(ctx, tx, eventID, nextInLine.UserID, 1); err != nil {
			return nil, err
		}
	} else {
		if err := r.DecrementEventTickets(ctx, tx, eventID); err != nil {
			return nil, err
		}
	}

	return nextInLine, tx.Commit(ctx)
}
