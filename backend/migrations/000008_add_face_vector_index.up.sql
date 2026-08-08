CREATE EXTENSION IF NOT EXISTS vector;

ALTER TABLE face_profiles
    ADD COLUMN embedding_vector vector(128);

UPDATE face_profiles
SET embedding_vector = ('[' || array_to_string(embedding, ',') || ']')::vector(128)
WHERE embedding_vector IS NULL;

ALTER TABLE face_profiles
    ALTER COLUMN embedding_vector SET NOT NULL;

CREATE INDEX face_profiles_model_status_idx
    ON face_profiles (embedding_model, embedding_version, status);

CREATE INDEX face_profiles_embedding_hnsw_idx
    ON face_profiles
    USING hnsw (embedding_vector vector_cosine_ops)
    WHERE status = 'ENROLLED';
