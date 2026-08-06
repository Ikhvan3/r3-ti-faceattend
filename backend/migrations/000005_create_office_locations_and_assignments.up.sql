CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE office_locations (
    id UUID PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    address TEXT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    radius_meters INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT office_locations_name_not_empty CHECK (length(trim(name)) > 0),
    CONSTRAINT office_locations_latitude_valid CHECK (latitude >= -90 AND latitude <= 90),
    CONSTRAINT office_locations_longitude_valid CHECK (longitude >= -180 AND longitude <= 180),
    CONSTRAINT office_locations_radius_valid CHECK (radius_meters >= 10 AND radius_meters <= 2000)
);

CREATE TABLE employee_location_assignments (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    office_location_id UUID NOT NULL REFERENCES office_locations(id) ON DELETE RESTRICT,
    effective_from DATE NOT NULL,
    effective_to DATE NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT employee_location_assignments_date_range_valid CHECK (
        effective_to IS NULL OR effective_to >= effective_from
    )
);

ALTER TABLE employee_location_assignments
    ADD CONSTRAINT employee_location_assignments_user_period_no_overlap
    EXCLUDE USING gist (
        user_id WITH =,
        daterange(effective_from, COALESCE(effective_to, 'infinity'::date), '[]') WITH &&
    );

ALTER TABLE attendance_records
    ADD COLUMN check_in_location_id UUID NULL REFERENCES office_locations(id) ON DELETE RESTRICT,
    ADD COLUMN check_in_latitude DOUBLE PRECISION NULL,
    ADD COLUMN check_in_longitude DOUBLE PRECISION NULL,
    ADD COLUMN check_in_accuracy_meters DOUBLE PRECISION NULL,
    ADD COLUMN check_in_distance_meters DOUBLE PRECISION NULL,
    ADD COLUMN check_out_location_id UUID NULL REFERENCES office_locations(id) ON DELETE RESTRICT,
    ADD COLUMN check_out_latitude DOUBLE PRECISION NULL,
    ADD COLUMN check_out_longitude DOUBLE PRECISION NULL,
    ADD COLUMN check_out_accuracy_meters DOUBLE PRECISION NULL,
    ADD COLUMN check_out_distance_meters DOUBLE PRECISION NULL,
    ADD CONSTRAINT attendance_records_check_in_latitude_valid CHECK (
        check_in_latitude IS NULL OR (check_in_latitude >= -90 AND check_in_latitude <= 90)
    ),
    ADD CONSTRAINT attendance_records_check_in_longitude_valid CHECK (
        check_in_longitude IS NULL OR (check_in_longitude >= -180 AND check_in_longitude <= 180)
    ),
    ADD CONSTRAINT attendance_records_check_out_latitude_valid CHECK (
        check_out_latitude IS NULL OR (check_out_latitude >= -90 AND check_out_latitude <= 90)
    ),
    ADD CONSTRAINT attendance_records_check_out_longitude_valid CHECK (
        check_out_longitude IS NULL OR (check_out_longitude >= -180 AND check_out_longitude <= 180)
    ),
    ADD CONSTRAINT attendance_records_check_in_accuracy_valid CHECK (
        check_in_accuracy_meters IS NULL OR check_in_accuracy_meters >= 0
    ),
    ADD CONSTRAINT attendance_records_check_out_accuracy_valid CHECK (
        check_out_accuracy_meters IS NULL OR check_out_accuracy_meters >= 0
    ),
    ADD CONSTRAINT attendance_records_check_in_distance_valid CHECK (
        check_in_distance_meters IS NULL OR check_in_distance_meters >= 0
    ),
    ADD CONSTRAINT attendance_records_check_out_distance_valid CHECK (
        check_out_distance_meters IS NULL OR check_out_distance_meters >= 0
    );

CREATE INDEX office_locations_active_idx ON office_locations (is_active);
CREATE INDEX employee_location_assignments_user_effective_idx
    ON employee_location_assignments (user_id, effective_from, effective_to);
CREATE INDEX employee_location_assignments_location_id_idx
    ON employee_location_assignments (office_location_id);
CREATE INDEX attendance_records_check_in_location_id_idx
    ON attendance_records (check_in_location_id);
CREATE INDEX attendance_records_check_out_location_id_idx
    ON attendance_records (check_out_location_id);
