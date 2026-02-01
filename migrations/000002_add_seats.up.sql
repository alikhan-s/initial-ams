-- Updating table flights
ALTER TABLE flights
    ADD COLUMN IF NOT EXISTS total_seats INT NOT NULL DEFAULT 150;

ALTER TABLE flights
DROP CONSTRAINT IF EXISTS check_seats_positive;

ALTER TABLE flights
    ADD CONSTRAINT check_seats_positive CHECK (total_seats > 0);
