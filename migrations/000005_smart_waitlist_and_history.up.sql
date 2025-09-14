-- Re-add the quantity column to bookings (it was removed in the refactor)
-- This is necessary for the new order-based model
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS quantity INT NOT NULL CHECK (quantity > 0);

-- We remove the unique constraint again. This is the key to allowing multiple distinct
-- bookings (history) from the same user for the same event.
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS bookings_user_id_event_id_key;

-- Add a quantity column to the waitlist to track how many tickets a user wants
ALTER TABLE waitlist_entries ADD COLUMN IF NOT EXISTS quantity INT NOT NULL CHECK (quantity > 0);