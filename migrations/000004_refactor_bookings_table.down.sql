-- This down migration is destructive, as we can't easily recover the old data structure.
-- For development, dropping the table is acceptable.
DROP TABLE IF EXISTS bookings;

-- You would then need to re-run the original migration to go back.