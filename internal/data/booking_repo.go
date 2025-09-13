package data

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BookingRepository struct {
	DB *pgxpool.Pool
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

func (r *BookingRepository) UpdateEventAndCreateBooking(ctx context.Context, event *EventForUpdate, userID string) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	updateQuery := `
		UPDATE events
		SET booked_tickets = booked_tickets + 1, version = version + 1
		WHERE id = $1 AND version = $2
	`
	tag, err := tx.Exec(ctx, updateQuery, event.ID, event.Version)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrConflict
	}

	insertQuery := `INSERT INTO bookings (user_id, event_id) VALUES ($1, $2)`
	_, err = tx.Exec(ctx, insertQuery, userID, event.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrDuplicate
		}
		return err
	}

	return tx.Commit(ctx)
}
