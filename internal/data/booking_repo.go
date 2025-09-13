package data

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BookingRepository struct {
	DB *pgxpool.Pool
}

type WaitlistUser struct {
	UserID string
	Email  string
}

func (r *BookingRepository) GetEventForUpdate(ctx context.Context, eventID string) (*EventForUpdate, error) {
	var e EventForUpdate
	query := `SELECT id, capacity, booked_tickets, version FROM events WHERE id = $1`
	err := r.DB.QueryRow(ctx, query, eventID).Scan(&e.ID, &e.Capacity, &e.BookedTickets, &e.Version)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *BookingRepository) CreateBookingInTx(ctx context.Context, tx pgx.Tx, eventID, userID string, quantity int) error {
	query := `INSERT INTO bookings (user_id, event_id) VALUES ($1, $2)`
	for i := 0; i < quantity; i++ {
		_, err := tx.Exec(ctx, query, userID, eventID)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return ErrDuplicate
			}
			return err
		}
	}
	return nil
}

func (r *BookingRepository) UpdateEventAndCreateBooking(ctx context.Context, event *EventForUpdate, userID string, quantity int) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	updateQuery := `
		UPDATE events
		SET booked_tickets = booked_tickets + $3, version = version + 1
		WHERE id = $1 AND version = $2
	`
	tag, err := tx.Exec(ctx, updateQuery, event.ID, event.Version, quantity)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrConflict
	}

	if err := r.CreateBookingInTx(ctx, tx, event.ID, userID, quantity); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *BookingRepository) AddToWaitlist(ctx context.Context, eventID, userID string) error {
	query := `INSERT INTO waitlist_entries (event_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.DB.Exec(ctx, query, eventID, userID)
	return err
}

func (r *BookingRepository) GetBookingForCancellation(ctx context.Context, tx pgx.Tx, bookingID, userID string) (string, error) {
	var eventID string
	query := `DELETE FROM bookings WHERE id = $1 AND user_id = $2 RETURNING event_id`
	err := tx.QueryRow(ctx, query, bookingID, userID).Scan(&eventID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	return eventID, nil
}

func (r *BookingRepository) GetAndRemoveNextFromWaitlist(ctx context.Context, tx pgx.Tx, eventID string) (*WaitlistUser, error) {
	var nextUser WaitlistUser
	query := `
        WITH next_in_line AS (
            SELECT id FROM waitlist_entries
            WHERE event_id = $1
            ORDER BY created_at ASC
            LIMIT 1
            FOR UPDATE SKIP LOCKED
        )
        DELETE FROM waitlist_entries
        WHERE id = (SELECT id FROM next_in_line)
        RETURNING user_id
    `
	var userID string
	err := tx.QueryRow(ctx, query, eventID).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	err = tx.QueryRow(ctx, "SELECT email FROM users WHERE id = $1", userID).Scan(&nextUser.Email)
	if err != nil {
		return nil, err
	}
	nextUser.UserID = userID
	return &nextUser, nil
}

func (r *BookingRepository) DecrementEventTickets(ctx context.Context, tx pgx.Tx, eventID string) error {
	_, err := tx.Exec(ctx, "UPDATE events SET booked_tickets = booked_tickets - 1, version = version + 1 WHERE id = $1", eventID)
	return err
}
