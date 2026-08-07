CREATE TABLE face_verification_grants (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    purpose VARCHAR(20) NOT NULL,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT face_verification_grants_purpose_allowed CHECK (purpose IN ('CHECK_IN', 'CHECK_OUT')),
    CONSTRAINT face_verification_grants_token_hash_unique UNIQUE (token_hash),
    CONSTRAINT face_verification_grants_token_hash_not_empty CHECK (length(trim(token_hash)) > 0),
    CONSTRAINT face_verification_grants_expires_after_created CHECK (expires_at > created_at)
);

CREATE INDEX face_verification_grants_user_purpose_idx
    ON face_verification_grants (user_id, purpose, created_at DESC);

CREATE INDEX face_verification_grants_active_idx
    ON face_verification_grants (token_hash, user_id, purpose)
    WHERE consumed_at IS NULL;
