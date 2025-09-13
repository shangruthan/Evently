package data

import "time"

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Event struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Venue         string    `json:"venue"`
	StartTime     time.Time `json:"start_time"`
	Capacity      int       `json:"capacity"`
	BookedTickets int       `json:"booked_tickets"`
	Version       int       `json:"-"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Booking struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	EventID   string    `json:"event_id"`
	CreatedAt time.Time `json:"created_at"`
}

type EventForUpdate struct {
	ID            string
	Capacity      int
	BookedTickets int
	Version       int
}
