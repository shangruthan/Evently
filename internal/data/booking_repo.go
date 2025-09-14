package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BookingRepository struct {
	DB *pgxpool.Pool
}

type UserBooking struct {
	BookingID   string    `json:"booking_id"`
	EventID     string    `json:"event_id"`
	EventName   string    `json:"event_name"`
	Quantity    int       `json:"quantity"`
	BookingTime time.Time `json:"booking_time"`
}

type WaitlistUser struct {
	UserID   string
	Email    string
	Quantity int
}

func (r *BookingRepository) GetEventForUpdate(ctx context.Context, eventID string) (*EventForUpdate, error) {
	var e EventForUpdate
	query := `SELECT id, capacity, booked_tickets, version FROM events WHERE id = $1`
	err := r.DB.QueryRow(ctx, query, eventID).Scan(&e.ID, &e.Capacity, &e.BookedTickets, &e.Version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &e, nil
}

func (r *BookingRepository) GetByUserID(ctx context.Context, userID string) ([]UserBooking, error) {
	query := `
		SELECT b.id, e.id, e.name, b.quantity, b.created_at
		FROM bookings b JOIN events e ON b.event_id = e.id
		WHERE b.user_id = $1 ORDER BY b.created_at DESC
	`
	rows, err := r.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []UserBooking
	for rows.Next() {
		var booking UserBooking
		if err := rows.Scan(&booking.BookingID, &booking.EventID, &booking.EventName, &booking.Quantity, &booking.BookingTime); err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}
	return bookings, nil
}

func (r *BookingRepository) HasWaitlist(ctx context.Context, eventID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM waitlist_entries WHERE event_id = $1)`
	err := r.DB.QueryRow(ctx, query, eventID).Scan(&exists)
	return exists, err
}

func (r *BookingRepository) CreateBooking(ctx context.Context, event *EventForUpdate, userID string, quantity int) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	updateEventQuery := `
		UPDATE events SET booked_tickets = booked_tickets + $3, version = version + 1
		WHERE id = $1 AND version = $2
	`
	tag, err := tx.Exec(ctx, updateEventQuery, event.ID, event.Version, quantity)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrConflict
	}

	insertBookingQuery := `INSERT INTO bookings (user_id, event_id, quantity) VALUES ($1, $2, $3)`
	_, err = tx.Exec(ctx, insertBookingQuery, userID, event.ID, quantity)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *BookingRepository) AddToWaitlist(ctx context.Context, eventID, userID string, quantity int) error {
	// This is the corrected query
	query := `
        INSERT INTO waitlist_entries (event_id, user_id, quantity) VALUES ($1, $2, $3)
        ON CONFLICT (user_id, event_id) DO NOTHING
    `
	_, err := r.DB.Exec(ctx, query, eventID, userID, quantity)
	return err
}

func (r *BookingRepository) UpdateBookingForCancellation(ctx context.Context, tx pgx.Tx, bookingID, userID string, quantityToCancel int) (string, int, error) {
	var eventID string
	var finalQuantity int
	query := `
        UPDATE bookings SET quantity = quantity - $3
        WHERE id = $1 AND user_id = $2 AND quantity >= $3
        RETURNING event_id, quantity
    `
	err := tx.QueryRow(ctx, query, bookingID, userID, quantityToCancel).Scan(&eventID, &finalQuantity)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", 0, errors.New("booking not found, or you are trying to cancel too many tickets")
		}
		return "", 0, err
	}
	if finalQuantity == 0 {
		if _, err := tx.Exec(ctx, "DELETE FROM bookings WHERE id = $1", bookingID); err != nil {
			return "", 0, err
		}
	}
	return eventID, quantityToCancel, nil
}

func (r *BookingRepository) FindAndRemoveMatchingWaitlistEntry(ctx context.Context, tx pgx.Tx, eventID string, availableTickets int) (*WaitlistUser, error) {
	var user WaitlistUser
	query := `
        WITH next_in_line AS (
            SELECT id, user_id, quantity FROM waitlist_entries
            WHERE event_id = $1 AND quantity <= $2
            ORDER BY created_at ASC LIMIT 1 FOR UPDATE SKIP LOCKED
        )
        DELETE FROM waitlist_entries WHERE id = (SELECT id FROM next_in_line)
        RETURNING user_id, quantity
    `
	err := tx.QueryRow(ctx, query, eventID, availableTickets).Scan(&user.UserID, &user.Quantity)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	err = tx.QueryRow(ctx, "SELECT email FROM users WHERE id = $1", user.UserID).Scan(&user.Email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *BookingRepository) DecrementEventTickets(ctx context.Context, tx pgx.Tx, eventID string, quantity int) error {
	_, err := tx.Exec(ctx, "UPDATE events SET booked_tickets = booked_tickets - $2, version = version + 1 WHERE id = $1", eventID, quantity)
	return err
}
