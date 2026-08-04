CREATE TABLE users (
    id UUID PRIMARY KEY,
    employee_number VARCHAR(50) NOT NULL,
    name VARCHAR(150) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash TEXT NOT NULL,
    phone VARCHAR(30),
    position VARCHAR(100),
    role VARCHAR(10) NOT NULL,
    account_status VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT users_employee_number_not_empty CHECK (length(trim(employee_number)) > 0),
    CONSTRAINT users_name_not_empty CHECK (length(trim(name)) > 0),
    CONSTRAINT users_email_not_empty CHECK (length(trim(email)) > 0),
    CONSTRAINT users_password_hash_not_empty CHECK (length(trim(password_hash)) > 0),
    CONSTRAINT users_role_allowed CHECK (role IN ('ADMIN', 'USER')),
    CONSTRAINT users_account_status_allowed CHECK (account_status IN ('ACTIVE', 'INACTIVE', 'SUSPENDED')),
    CONSTRAINT users_employee_number_unique UNIQUE (employee_number)
);

CREATE UNIQUE INDEX users_email_lower_unique ON users (lower(email));
