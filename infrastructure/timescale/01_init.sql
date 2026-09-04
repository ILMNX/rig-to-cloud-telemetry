-- Drilling telemetry hypertable for local TimescaleDB development.

CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE TABLE IF NOT EXISTS drilling_telemetry (
    time        TIMESTAMPTZ       NOT NULL,
    well_id     TEXT              NOT NULL,
    bit_depth   DOUBLE PRECISION,
    rop         DOUBLE PRECISION,
    wob         DOUBLE PRECISION,
    gamma_ray   DOUBLE PRECISION
);

SELECT create_hypertable('drilling_telemetry', 'time', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_drilling_telemetry_well_id_time
    ON drilling_telemetry (well_id, time DESC);
