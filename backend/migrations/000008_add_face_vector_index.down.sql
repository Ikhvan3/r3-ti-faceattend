DROP INDEX IF EXISTS face_profiles_embedding_hnsw_idx;
DROP INDEX IF EXISTS face_profiles_model_status_idx;
ALTER TABLE face_profiles DROP COLUMN IF EXISTS embedding_vector;

-- The vector extension is intentionally kept installed because PostgreSQL
-- extensions can be shared by other schemas/features in the same database.
