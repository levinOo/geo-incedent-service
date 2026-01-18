-- +goose Up
CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    area GEOGRAPHY(POLYGON, 4326) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_incidents_area ON incidents USING GIST (area);
CREATE INDEX IF NOT EXISTS idx_incidents_is_active ON incidents (is_active);

CREATE TABLE IF NOT EXISTS location_checks (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    user_location GEOGRAPHY(POINT, 4326) NOT NULL,
    is_danger BOOLEAN NOT NULL,
    incident_id UUID REFERENCES incidents(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_location_checks_created_at ON location_checks (created_at);
CREATE INDEX IF NOT EXISTS idx_location_checks_incident_id ON location_checks (incident_id);
-- Optional: Index on user_id if we query by user history
CREATE INDEX IF NOT EXISTS idx_location_checks_user_id ON location_checks (user_id);

-- +goose Down
DROP TABLE IF EXISTS location_checks;
DROP TABLE IF EXISTS incidents;
