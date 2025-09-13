// package data

// import (
// 	"context"
// 	"errors"

// 	"github.com/jackc/pgx/v5"
// 	"github.com/jackc/pgx/v5/pgxpool"
// )

// type EventRepository struct {
// 	DB *pgxpool.Pool
// }

// func (r *EventRepository) GetAll(ctx context.Context) ([]Event, error) {
// 	query := `SELECT id, name, venue, start_time, capacity, booked_tickets FROM events ORDER BY start_time ASC`

// 	rows, err := r.DB.Query(ctx, query)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var events []Event
// 	for rows.Next() {
// 		var event Event
// 		err := rows.Scan(
// 			&event.ID,
// 			&event.Name,
// 			&event.Venue,
// 			&event.StartTime,
// 			&event.Capacity,
// 			&event.BookedTickets,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}
// 		events = append(events, event)
// 	}

// 	return events, nil
// }

// func (r *EventRepository) GetByID(ctx context.Context, id string) (*Event, error) {
// 	query := `SELECT id, name, venue, start_time, capacity, booked_tickets FROM events WHERE id = $1`
// 	var event Event
// 	err := r.DB.QueryRow(ctx, query, id).Scan(
// 		&event.ID,
// 		&event.Name,
// 		&event.Venue,
// 		&event.StartTime,
// 		&event.Capacity,
// 		&event.BookedTickets,
// 	)
// 	if err != nil {
// 		if errors.Is(err, pgx.ErrNoRows) {
// 			return nil, ErrNotFound
// 		}
// 		return nil, err
// 	}
// 	return &event, nil
// }

package data

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepository struct {
	DB *pgxpool.Pool
}

func (r *EventRepository) GetAll(ctx context.Context) ([]Event, error) {
	query := `SELECT id, name, venue, start_time, capacity, booked_tickets FROM events ORDER BY start_time ASC`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
		err := rows.Scan(
			&event.ID,
			&event.Name,
			&event.Venue,
			&event.StartTime,
			&event.Capacity,
			&event.BookedTickets,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

func (r *EventRepository) GetByID(ctx context.Context, id string) (*Event, error) {
	query := `SELECT id, name, venue, start_time, capacity, booked_tickets FROM events WHERE id = $1`
	var event Event
	err := r.DB.QueryRow(ctx, query, id).Scan(
		&event.ID,
		&event.Name,
		&event.Venue,
		&event.StartTime,
		&event.Capacity,
		&event.BookedTickets,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &event, nil
}

func (r *EventRepository) Create(ctx context.Context, event *Event) error {
	query := `
		INSERT INTO events (name, venue, start_time, capacity)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at, version, booked_tickets
	`
	args := []interface{}{event.Name, event.Venue, event.StartTime, event.Capacity}
	return r.DB.QueryRow(ctx, query, args...).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt, &event.Version, &event.BookedTickets)
}
