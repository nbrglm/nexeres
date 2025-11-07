-- Nexeres - Down Migration
-- Rollback schema for Nexeres.
-- 1. Drop Indexes
DROP INDEX IF EXISTS idx_org_domains_org_id;

DROP INDEX IF EXISTS idx_roles_org_id;

DROP INDEX IF EXISTS idx_user_orgs_org_id;

DROP INDEX IF EXISTS idx_user_orgs_role_id;

DROP INDEX IF EXISTS idx_mfa_factors_user_id;

DROP INDEX IF EXISTS idx_sessions_user_org;

DROP INDEX IF EXISTS idx_sessions_expires_at;

DROP INDEX IF EXISTS idx_verification_tokens_user_id;

DROP INDEX IF EXISTS idx_invitations_org_id;

DROP INDEX IF EXISTS idx_invitations_org_email;

DROP INDEX IF EXISTS idx_audit_logs_org_id;

DROP INDEX IF EXISTS idx_audit_logs_user_id;

DROP INDEX IF EXISTS idx_audit_logs_created_at;

DROP INDEX IF EXISTS idx_users_email;

DROP INDEX IF EXISTS idx_orgs_created_at;

-- 2. Drop Triggers
DROP TRIGGER IF EXISTS trg_orgs_updated_at ON orgs;

DROP TRIGGER IF EXISTS trg_org_domains_updated_at ON org_domains;

DROP TRIGGER IF EXISTS trg_users_updated_at ON users;

DROP TRIGGER IF EXISTS trg_roles_updated_at ON roles;

DROP TRIGGER IF EXISTS trg_user_orgs_updated_at ON user_orgs;

DROP TRIGGER IF EXISTS trg_mfa_factors_updated_at ON mfa_factors;

-- 3. Drop Function
DROP FUNCTION IF EXISTS update_updated_at_column;

-- 4. Drop Tables (in reverse dependency order)
DROP TABLE IF EXISTS audit_logs;

DROP TABLE IF EXISTS invitations;

DROP TABLE IF EXISTS verification_tokens;

DROP TABLE IF EXISTS sessions;

DROP TABLE IF EXISTS mfa_factors;

DROP TABLE IF EXISTS user_orgs;

DROP TABLE IF EXISTS users;

DROP TABLE IF EXISTS org_domains;

DROP TABLE IF EXISTS roles;

DROP TABLE IF EXISTS orgs;

-- 5. Drop Types (reverse order of creation)
DROP TYPE IF EXISTS log_action_actor_type;

DROP TYPE IF EXISTS audit_log_entity;

DROP TYPE IF EXISTS audit_log_action;

DROP TYPE IF EXISTS invitation_status;

DROP TYPE IF EXISTS verification_token_type;

DROP TYPE IF EXISTS mfa_type;

-- 6. Drop Extension
DROP EXTENSION IF EXISTS pgcrypto;