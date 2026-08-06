CREATE EXTENSION IF NOT EXISTS btree_gist;

ALTER TABLE employee_schedule_assignments
    ADD CONSTRAINT employee_schedule_assignments_user_period_no_overlap
    EXCLUDE USING gist (
        user_id WITH =,
        daterange(effective_from, COALESCE(effective_to, 'infinity'::date), '[]') WITH &&
    );
