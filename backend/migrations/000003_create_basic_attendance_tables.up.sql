CREATE TABLE work_schedules (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    grace_minutes INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT work_schedules_name_not_empty CHECK (length(trim(name)) > 0),
    CONSTRAINT work_schedules_grace_minutes_non_negative CHECK (grace_minutes >= 0),
    CONSTRAINT work_schedules_name_unique UNIQUE (name)
);

CREATE TABLE employee_schedule_assignments (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    schedule_id UUID NOT NULL REFERENCES work_schedules(id) ON DELETE RESTRICT,
    effective_from DATE NOT NULL,
    effective_to DATE NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT employee_schedule_assignments_date_range_valid CHECK (
        effective_to IS NULL OR effective_to >= effective_from
    ),
    CONSTRAINT employee_schedule_assignments_user_schedule_from_unique UNIQUE (
        user_id, schedule_id, effective_from
    )
);

CREATE TABLE attendance_records (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    schedule_id UUID NOT NULL REFERENCES work_schedules(id) ON DELETE RESTRICT,
    attendance_date DATE NOT NULL,
    check_in_at TIMESTAMPTZ NOT NULL,
    check_out_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT attendance_records_check_out_after_check_in CHECK (
        check_out_at IS NULL OR check_out_at >= check_in_at
    ),
    CONSTRAINT attendance_records_user_date_unique UNIQUE (user_id, attendance_date)
);

CREATE INDEX work_schedules_active_idx ON work_schedules (is_active);
CREATE INDEX employee_schedule_assignments_user_effective_idx
    ON employee_schedule_assignments (user_id, effective_from, effective_to);
CREATE INDEX employee_schedule_assignments_schedule_id_idx
    ON employee_schedule_assignments (schedule_id);
CREATE INDEX attendance_records_user_date_desc_idx
    ON attendance_records (user_id, attendance_date DESC);
CREATE INDEX attendance_records_schedule_id_idx
    ON attendance_records (schedule_id);
