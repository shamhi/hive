CREATE TABLE IF NOT EXISTS assignments (
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL,
    drone_id TEXT NOT NULL,
    status TEXT NOT NULL,
    target_lat DOUBLE PRECISION NULL,
    target_lon DOUBLE PRECISION NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_assignments_drone_id ON assignments(drone_id);
CREATE INDEX IF NOT EXISTS idx_assignments_status ON assignments(status);
