CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    actor_email VARCHAR(255) NOT NULL,
    actor_role VARCHAR(20) NOT NULL,
    action VARCHAR(80) NOT NULL,
    entity_type VARCHAR(80) NOT NULL,
    entity_id UUID NULL,
    target_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    target_label VARCHAR(255) NULL,
    reason TEXT NOT NULL,
    before_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    after_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT audit_logs_actor_email_not_empty CHECK (length(trim(actor_email)) > 0),
    CONSTRAINT audit_logs_actor_role_not_empty CHECK (length(trim(actor_role)) > 0),
    CONSTRAINT audit_logs_action_not_empty CHECK (length(trim(action)) > 0),
    CONSTRAINT audit_logs_entity_type_not_empty CHECK (length(trim(entity_type)) > 0),
    CONSTRAINT audit_logs_reason_not_empty CHECK (length(trim(reason)) >= 5),
    CONSTRAINT audit_logs_reason_length CHECK (length(reason) <= 1000)
);

CREATE INDEX audit_logs_created_at_desc_idx ON audit_logs (created_at DESC);
CREATE INDEX audit_logs_action_created_at_idx ON audit_logs (action, created_at DESC);
CREATE INDEX audit_logs_entity_created_at_idx ON audit_logs (entity_type, entity_id, created_at DESC);
CREATE INDEX audit_logs_target_user_created_at_idx ON audit_logs (target_user_id, created_at DESC);
CREATE INDEX audit_logs_actor_user_created_at_idx ON audit_logs (actor_user_id, created_at DESC);
