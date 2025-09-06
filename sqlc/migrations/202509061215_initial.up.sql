-- Nexeres - Schema
-- This file contains the SQL schema for Nexeres.
-- Works with PostgreSQL 17+.
-- Create the "orgs" table to store organization details.
CREATE TABLE IF NOT EXISTS orgs (
  id UUID PRIMARY KEY NOT NULL,
  -- URL-safe slug for the org, used in URLs and as a unique identifier.
  -- It must be unique across all orgs and can only contain lowercase letters, numbers, and hyphens.
  slug VARCHAR(128) UNIQUE NOT NULL CHECK (slug ~ '^[a-z0-9-]+$'),
  name VARCHAR(512) NOT NULL,
  description TEXT,
  avatar_url TEXT,
  domain_verification_secret VARCHAR(128) NOT NULL DEFAULT encode(gen_random_bytes(32), 'hex'),
  settings JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ DEFAULT NULL
);

-- The domain is used to identify the org in Nexeres.
-- It is used for auto-joining users to the org based on their email address.
-- The domain must be a valid email domain.
CREATE TABLE IF NOT EXISTS org_domains (
  -- The domain name, used to identify the org in Nexeres.
  domain VARCHAR(512) PRIMARY KEY,
  org_id UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
  verified BOOLEAN NOT NULL DEFAULT FALSE,
  verified_at TIMESTAMPTZ,
  auto_join BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Enable pgcrypto if not already enabled.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Insert the default org, which is used when multitenancy is disabled.
INSERT INTO orgs (
    id,
    slug,
    name,
    description,
    domain_verification_secret
  )
SELECT '00000000-0000-7000-0000-000000000001',
  'default',
  'Default Org',
  'Default organization for all users in Nexeres Single Tenant Mode.',
  -- random 32-byte secret
  encode(gen_random_bytes(32), 'hex')
WHERE NOT EXISTS (
    SELECT 1
    FROM orgs
    WHERE id = '00000000-0000-7000-0000-000000000001'
  );

-- Users are global entities, they can belong to multiple orgs.
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY NOT NULL,
  email VARCHAR(512) NOT NULL UNIQUE,
  -- Set to true if the user has verified their email address, or if the user is connected via OAuth or SSO.
  email_verified BOOLEAN NOT NULL DEFAULT FALSE,
  -- NULLABLE only if the user is connected via OAuth or SSO.
  password_hash VARCHAR(512),
  -- Whether the user has enabled multi-factor authentication (MFA) for their account.
  mfa_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  -- MFA backup codes for the user, used for multi-factor authentication.
  -- Each code is a unique, randomly generated string that can be used to authenticate the user.
  -- The codes are encrypted in the database.
  -- The user can generate new codes at any time, which will invalidate the old codes.
  -- Only NON-NULL if the user has enabled multi-factor authentication (MFA).
  backup_codes JSONB NOT NULL DEFAULT '[]'::jsonb,
  first_name VARCHAR(512),
  middle_name VARCHAR(512),
  last_name VARCHAR(512),
  avatar_url TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ DEFAULT NULL
);

-- Roles table to define custom roles and their permissions.
CREATE TABLE IF NOT EXISTS roles (
  id UUID PRIMARY KEY NOT NULL,
  org_id UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
  role_name TEXT NOT NULL,
  permissions JSONB NOT NULL DEFAULT '[]'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (org_id, role_name)
);

-- Junction table to associate users with orgs.
CREATE TABLE IF NOT EXISTS user_orgs (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  org_id UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
  -- 'admin' means the user is an admin of the org, can manage org settings and users.
  -- if this is false, the user is a regular member of the org.
  is_org_admin BOOLEAN NOT NULL DEFAULT FALSE,
  -- The role_id references the roles table, which defines custom roles and their permissions.
  -- This allows for more granular control over user permissions in the org.
  role_id UUID REFERENCES roles(id) ON DELETE RESTRICT,
  -- The date and time when the user joined the org.
  joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  -- Updated at column to track the last time the user_orgs record was updated.
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  -- 'banned' means the user is banned from the org, no sessions can be created for the user in the org.
  is_banned BOOLEAN NOT NULL DEFAULT FALSE,
  PRIMARY KEY (user_id, org_id)
);

-- MFA Types, Enum.
CREATE TYPE mfa_type AS ENUM ('email', 'totp');

-- MFA Factors table, per-user.
CREATE TABLE IF NOT EXISTS mfa_factors (
  id UUID PRIMARY KEY NOT NULL,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  -- The type of the MFA factor.
  factor_type mfa_type NOT NULL,
  -- The user-friendly name of the MFA factor, used to identify the factor in the UI.
  name VARCHAR(255) NOT NULL,
  -- The secret, as follows for every type of factor:
  -- TOTP: Base32 secret key
  -- SMS: Phone number (+1234567890) (TODO)
  -- Email: Email address
  -- WebAuthn: Public key data (TODO)
  secret VARCHAR(512) NOT NULL,
  -- Whether the method is verified or not.
  verified BOOLEAN NOT NULL DEFAULT FALSE,
  last_used_at TIMESTAMPTZ DEFAULT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, factor_type, name)
);

-- Sessions table, scoped to orgs and users.
CREATE TABLE IF NOT EXISTS sessions (
  id UUID PRIMARY KEY NOT NULL,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  org_id UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
  -- The session token hash, used to authenticate the user in Nexeres.
  session_token_hash VARCHAR(128) UNIQUE NOT NULL,
  -- The refresh token hash corresponding to this session.
  refresh_token_hash VARCHAR(128) UNIQUE NOT NULL,
  mfa_verified BOOLEAN NOT NULL DEFAULT FALSE,
  ip_address INET NOT NULL,
  user_agent TEXT NOT NULL,
  -- The date and time when mfa was last verified for this session.
  mfa_verified_at TIMESTAMPTZ DEFAULT NULL,
  -- The date and time when the session, and the refresh token will expire.
  expires_at TIMESTAMPTZ NOT NULL,
  -- The date and time when the session was created.
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- A type for the verification tokens.
CREATE TYPE verification_token_type AS ENUM ('email_verification', 'password_reset');

-- Verification tokens table, scoped to users.
CREATE TABLE IF NOT EXISTS verification_tokens (
  id UUID PRIMARY KEY NOT NULL,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  -- The type of the verification token.
  token_type verification_token_type NOT NULL,
  -- The token hash, used to verify the token.
  -- 32 bytes -> sha256 hashed -> hex encoded = 64 characters => we use 128 for safety and flexibility.
  token_hash VARCHAR(128) UNIQUE NOT NULL,
  -- The date and time when the token will expire.
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Invitation Status type, Enum.
CREATE TYPE invitation_status AS ENUM ('pending', 'accepted', 'declined', 'expired');

-- Invitations table, scoped to orgs.
CREATE TABLE IF NOT EXISTS invitations (
  id UUID PRIMARY KEY NOT NULL,
  org_id UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
  -- The email address of the user being invited to the org.
  email VARCHAR(512) NOT NULL,
  -- Whether the invited user should be made an admin of the org upon accepting the invitation.
  invite_as_admin BOOLEAN NOT NULL DEFAULT FALSE,
  -- The user who created the invitation, NULL if the invitation was created by the system.
  invited_by UUID REFERENCES users(id) ON DELETE
  SET NULL,
    -- The token hash used to verify the invitation.
    token_hash VARCHAR(128) NOT NULL UNIQUE,
    -- The date and time when the invitation will expire.
    expires_at TIMESTAMPTZ NOT NULL,
    -- The date and time when the invitation was accepted, NULL if the invitation has not been accepted yet.
    accepted_at TIMESTAMPTZ DEFAULT NULL,
    -- The date and time when the invitation was created.
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- The status of the invitation, one of 'pending', 'accepted', 'declined', 'expired'.
    invite_status invitation_status NOT NULL DEFAULT 'pending'::invitation_status,
    UNIQUE (org_id, email)
);

-- Audit Log Action type, Enum.
CREATE TYPE audit_log_action AS ENUM (
  'create',
  'update',
  'delete',
  'read',
  'login',
  'logout',
  'password_change',
  'invite_accepted',
  'invite_declined',
  'org_settings_updated'
);

-- Audit Log Entity, Enum.
CREATE TYPE audit_log_entity AS ENUM (
  'user',
  'org',
  'org_domain',
  'role',
  'org_membership',
  'session',
  'invitation',
  'mfa_factor'
);

-- Audit Log Actor Type, Enum.
CREATE TYPE log_action_actor_type AS ENUM ('user', 'sysadmin', 'system');

-- Audit Log Table, scoped to orgs and users.
CREATE TABLE IF NOT EXISTS audit_logs (
  id UUID PRIMARY KEY NOT NULL,
  org_id UUID REFERENCES orgs(id) ON DELETE CASCADE,
  user_id UUID REFERENCES users(id) ON DELETE
  SET NULL,
    actor_type log_action_actor_type NOT NULL,
    log_action audit_log_action NOT NULL,
    log_entity audit_log_entity NOT NULL,
    entity_id UUID,
    -- A JSONB object containing additional details about the action.
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- A Function to update updated_at columns
CREATE OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $$ BEGIN NEW.updated_at = NOW();

RETURN NEW;

END;

$$ LANGUAGE plpgsql;

-- Triggers to update updated_at columns on update
-- Trigger for orgs
CREATE TRIGGER trg_orgs_updated_at BEFORE
UPDATE ON orgs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger for org_domains
CREATE TRIGGER trg_org_domains_updated_at BEFORE
UPDATE ON org_domains FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger for users
CREATE TRIGGER trg_users_updated_at BEFORE
UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger for roles
CREATE TRIGGER trg_roles_updated_at BEFORE
UPDATE ON roles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger for user_orgs
CREATE TRIGGER trg_user_orgs_updated_at BEFORE
UPDATE ON user_orgs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger for mfa_factors
CREATE TRIGGER trg_mfa_factors_updated_at BEFORE
UPDATE ON mfa_factors FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Indexes to improve performance
-- 1. org_domains: frequent lookup by org_id for joining users
CREATE INDEX idx_org_domains_org_id ON org_domains(org_id);

-- 2. roles: lookup roles by org_id (common in permission checks)
CREATE INDEX idx_roles_org_id ON roles(org_id);

-- 3. user_orgs: fast access by org_id to list users in an org
CREATE INDEX idx_user_orgs_org_id ON user_orgs(org_id);

-- 4. user_orgs: fast lookup of users by role
CREATE INDEX idx_user_orgs_role_id ON user_orgs(role_id);

-- 5. mfa_factors: lookup factors by user_id
CREATE INDEX idx_mfa_factors_user_id ON mfa_factors(user_id);

-- 6. sessions: lookup sessions by user_id and org_id
CREATE INDEX idx_sessions_user_org ON sessions(user_id, org_id);

-- 7. sessions: expire session queries by expires_at
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- 8. verification_tokens: lookup by user_id
CREATE INDEX idx_verification_tokens_user_id ON verification_tokens(user_id);

-- 9. invitations: lookup invitations by org_id
CREATE INDEX idx_invitations_org_id ON invitations(org_id);

-- 10. invitations: frequently queried by email within org
CREATE INDEX idx_invitations_org_email ON invitations(org_id, email);

-- 11. audit_logs: filter by org_id for org-specific logs
CREATE INDEX idx_audit_logs_org_id ON audit_logs(org_id);

-- 12. audit_logs: filter by user_id
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);

-- 13. audit_logs: filter by created_at (recent logs)
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);

-- 14. users: filter or search by email
CREATE INDEX idx_users_email ON users(email);

-- 15. users: filter by deleted_at (active vs soft-deleted)
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- 16. orgs: filter by deleted_at (active orgs)
CREATE INDEX idx_orgs_deleted_at ON orgs(deleted_at);

-- 17. orgs: filter/search by created_at (recent orgs)
CREATE INDEX idx_orgs_created_at ON orgs(created_at);