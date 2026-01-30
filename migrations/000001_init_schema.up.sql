CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Users & Auth
CREATE TABLE users (
                       id BIGSERIAL PRIMARY KEY,
                       full_name TEXT NOT NULL,
                       email TEXT UNIQUE NOT NULL,
                       password_hash TEXT NOT NULL,
                       role VARCHAR(20) NOT NULL CHECK (role IN ('PASSENGER', 'STAFF', 'ADMIN')),
                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Infrastructure (Airport Ops)
CREATE TABLE terminals (
                           id BIGSERIAL PRIMARY KEY,
                           name TEXT UNIQUE NOT NULL -- e.g. "Terminal 1", "International"
);

CREATE TABLE gates (
                       id BIGSERIAL PRIMARY KEY,
                       terminal_id BIGINT NOT NULL REFERENCES terminals(id) ON DELETE CASCADE,
                       code TEXT NOT NULL, -- e.g. "A1", "B2"
                       status VARCHAR(20) NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'CLOSED', 'MAINTENANCE')),
                       UNIQUE (terminal_id, code) -- Prevent duplicate gate codes in the same terminal
);

-- 3. Flights
CREATE TABLE flights (
                         id BIGSERIAL PRIMARY KEY,
                         flight_no VARCHAR(10) UNIQUE NOT NULL, -- e.g. "KC-881"
                         origin VARCHAR(3) NOT NULL, -- IATA code, e.g. "NQZ"
                         destination VARCHAR(3) NOT NULL, -- IATA code, e.g. "GUW"
                         gate_id BIGINT REFERENCES gates(id) ON DELETE SET NULL, -- Gate can be null initially
                         departure_time TIMESTAMPTZ NOT NULL,
                         arrival_time TIMESTAMPTZ NOT NULL,
                         status VARCHAR(20) NOT NULL DEFAULT 'SCHEDULED'
                             CHECK (status IN ('SCHEDULED', 'DELAYED', 'BOARDING', 'DEPARTED', 'CANCELLED')),
                         version INT NOT NULL DEFAULT 1, -- For Optimistic Locking
                         created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                         updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for searching flights
CREATE INDEX idx_flights_schedule ON flights(origin, destination, departure_time);

-- 4. Passengers (Profile linked to User)
CREATE TABLE passengers (
                            id BIGSERIAL PRIMARY KEY,
                            user_id BIGINT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                            passport_no TEXT UNIQUE NOT NULL, -- Encrypted in app level ideally, but text for now
                            phone TEXT
);

-- 5. Bookings & Tickets
CREATE TABLE tickets (
                         id BIGSERIAL PRIMARY KEY,
                         flight_id BIGINT NOT NULL REFERENCES flights(id),
                         passenger_id BIGINT NOT NULL REFERENCES passengers(id),
                         seat_no VARCHAR(5), -- Nullable until check-in
                         price DECIMAL(10, 2) NOT NULL CHECK (price >= 0),
                         status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'CANCELLED')),
                         created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                         UNIQUE (flight_id, seat_no) -- Prevent double booking the same seat
);

-- 6. Baggage
CREATE TABLE baggage (
                         id BIGSERIAL PRIMARY KEY,
                         ticket_id BIGINT NOT NULL REFERENCES tickets(id),
                         tag_code VARCHAR(50) UNIQUE NOT NULL,
                         status VARCHAR(20) NOT NULL DEFAULT 'CREATED'
                             CHECK (status IN ('CREATED', 'RECEIVED', 'SCREENED', 'LOADED', 'UNLOADED', 'DELIVERED', 'LOST')),
                         updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);