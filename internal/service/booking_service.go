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
	ErrBookingConflict = errors.New("booking conflict, please try again")
	ErrJoinWaitlist    = errors.New("tickets are reserved for the waitlist, please join the waitlist")
	ErrAddedToWaitlist = errors.New("not enough tickets available, you have been added to the waitlist")
	MaxRetries         = 3
)

type BookingRepo interface {
	GetEventForUpdate(ctx context.Context, eventID string) (*data.EventForUpdate, error)
	CreateBooking(ctx context.Context, event *data.EventForUpdate, userID string, quantity int) error
	HasWaitlist(ctx context.Context, eventID string) (bool, error)
	AddToWaitlist(ctx context.Context, eventID, userID string, quantity int) error
	CancelBooking(ctx context.Context, bookingID, userID string, quantity int) error
	GetUserBookings(ctx context.Context, userID string) ([]data.UserBooking, error)
}

type BookingService struct {
	repo BookingRepo
	log  *slog.Logger
}

func NewBookingService(repo BookingRepo, log *slog.Logger) *BookingService {
	return &BookingService{repo: repo, log: log}
}

func (s *BookingService) CreateBooking(ctx context.Context, eventID, userID string, quantity int) error {
	hasWaitlist, err := s.repo.HasWaitlist(ctx, eventID)
	if err != nil {
		return err
	}

	for i := 0; i < MaxRetries; i++ {
		event, err := s.repo.GetEventForUpdate(ctx, eventID)
		if err != nil {
			return err
		}

		if (event.BookedTickets + quantity) > event.Capacity {
			if err := s.repo.AddToWaitlist(ctx, eventID, userID, quantity); err != nil {
				return err
			}
			return ErrAddedToWaitlist
		}

		if hasWaitlist {
			if err := s.repo.AddToWaitlist(ctx, eventID, userID, quantity); err != nil {
				return err
			}
			return ErrJoinWaitlist
		}

		err = s.repo.CreateBooking(ctx, event, userID, quantity)
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

func (s *BookingService) CancelBooking(ctx context.Context, bookingID, userID string, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity to cancel must be positive")
	}
	return s.repo.CancelBooking(ctx, bookingID, userID, quantity)
}

func (s *BookingService) GetUserBookings(ctx context.Context, userID string) ([]data.UserBooking, error) {
	return s.repo.GetUserBookings(ctx, userID)
}

// --- Interface Implementation ---
type BookingRepositoryWithTx struct {
	DB *pgxpool.Pool
	*data.BookingRepository
}

func (r *BookingRepositoryWithTx) CancelBooking(ctx context.Context, bookingID, userID string, quantity int) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	eventID, qtyCancelled, err := r.UpdateBookingForCancellation(ctx, tx, bookingID, userID, quantity)
	if err != nil {
		return err
	}

	ticketsAvailable := qtyCancelled
	for ticketsAvailable > 0 {
		waitlister, err := r.FindAndRemoveMatchingWaitlistEntry(ctx, tx, eventID, ticketsAvailable)
		if err != nil {
			return err
		}
		if waitlister == nil {
			break // No more matching entries on the waitlist
		}

		// A user from the waitlist gets their tickets.
		// We don't need to change event.booked_tickets, it's just a transfer.
		insertQuery := `INSERT INTO bookings (user_id, event_id, quantity) VALUES ($1, $2, $3)`
		if _, err := tx.Exec(ctx, insertQuery, waitlister.UserID, eventID, waitlister.Quantity); err != nil {
			return err
		}
		slog.Default().Info("auto-booked tickets for waitlisted user", "user_id", waitlister.UserID, "quantity", waitlister.Quantity)
		ticketsAvailable -= waitlister.Quantity
	}

	if ticketsAvailable > 0 {
		if err := r.DecrementEventTickets(ctx, tx, eventID, ticketsAvailable); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *BookingRepositoryWithTx) GetUserBookings(ctx context.Context, userID string) ([]data.UserBooking, error) {
	return r.BookingRepository.GetByUserID(ctx, userID)
}

func (r *BookingRepositoryWithTx) CreateBooking(ctx context.Context, event *data.EventForUpdate, userID string, quantity int) error {
	return r.BookingRepository.CreateBooking(ctx, event, userID, quantity)
}
