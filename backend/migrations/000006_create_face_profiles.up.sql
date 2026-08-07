CREATE TABLE face_profiles (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    embedding DOUBLE PRECISION[] NOT NULL,
    embedding_model VARCHAR(100) NOT NULL,
    embedding_version VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL,
    enrolled_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT face_profiles_user_unique UNIQUE (user_id),
    CONSTRAINT face_profiles_embedding_not_empty CHECK (array_length(embedding, 1) > 0),
    CONSTRAINT face_profiles_embedding_model_not_empty CHECK (length(trim(embedding_model)) > 0),
    CONSTRAINT face_profiles_embedding_version_not_empty CHECK (length(trim(embedding_version)) > 0),
    CONSTRAINT face_profiles_status_allowed CHECK (status IN ('ENROLLED')),
    CONSTRAINT face_profiles_enrolled_at_required CHECK (
        status <> 'ENROLLED' OR enrolled_at IS NOT NULL
    )
);

CREATE INDEX face_profiles_user_id_idx ON face_profiles (user_id);
CREATE INDEX face_profiles_status_idx ON face_profiles (status);
