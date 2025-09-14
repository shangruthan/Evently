ALTER TABLE waitlist_entries DROP COLUMN IF EXISTS quantity;
ALTER TABLE bookings ADD CONSTRAINT bookings_user_id_event_id_key UNIQUE (user_id, event_id);
ALTER TABLE bookings DROP COLUMN IF EXISTS quantity;