DROP INDEX IF EXISTS attendance_records_check_out_location_id_idx;
DROP INDEX IF EXISTS attendance_records_check_in_location_id_idx;
DROP INDEX IF EXISTS employee_location_assignments_location_id_idx;
DROP INDEX IF EXISTS employee_location_assignments_user_effective_idx;
DROP INDEX IF EXISTS office_locations_active_idx;

ALTER TABLE attendance_records
    DROP COLUMN IF EXISTS check_out_distance_meters,
    DROP COLUMN IF EXISTS check_out_accuracy_meters,
    DROP COLUMN IF EXISTS check_out_longitude,
    DROP COLUMN IF EXISTS check_out_latitude,
    DROP COLUMN IF EXISTS check_out_location_id,
    DROP COLUMN IF EXISTS check_in_distance_meters,
    DROP COLUMN IF EXISTS check_in_accuracy_meters,
    DROP COLUMN IF EXISTS check_in_longitude,
    DROP COLUMN IF EXISTS check_in_latitude,
    DROP COLUMN IF EXISTS check_in_location_id;

DROP TABLE IF EXISTS employee_location_assignments;
DROP TABLE IF EXISTS office_locations;
