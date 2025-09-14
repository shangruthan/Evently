-- First, drop the old table and its dependencies if they exist to start clean
DROP TABLE IF EXISTS bookings CASCADE;

-- Recreate the bookings table with a quantity column and a unique constraint
CREATE TABLE bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    quantity INT NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, event_id)
);

CREATE INDEX ON bookings (user_id);
CREATE INDEX ON bookings (event_id);