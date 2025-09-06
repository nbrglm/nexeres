-- Nexeres - Schema
-- This file contains the SQL schema for Nexeres.
-- Works with PostgreSQL 17+.
CREATE TABLE IF NOT EXISTS orgs (
  id UUID PRIMARY KEY NOT NULL,
  -- URL-safe slug for the org, used in URLs and as a unique identifier.
  -- It must be unique across all orgs and can only contain lowercase letters, numbers, and hyphens.
  slug VARCHAR(64) UNIQUE NOT NULL,
  name VARCHAR(512) NOT NULL,
  description TEXT,
  avatar_url TEXT,
  settings JSONB NOT NULL DEFAULT '{}'::jsonb,
  domain_verification_secret VARCHAR(256) NOT NULL,
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
  verification_token_hash VARCHAR(512) NOT NULL,
  verified_at TIMESTAMPTZ,
  auto_join BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert the default org, which is used when no org is specified.
INSERT INTO orgs (id, slug, name)
VALUES (
    '019735ab-f216-717f-9d12-e3915453c8d0',
    'default',
    'Default Tenant'
  ) ON CONFLICT (slug) DO NOTHING;

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
  last_name VARCHAR(512),
  avatar_url TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ DEFAULT NULL
);

-- Junction table to associate users with orgs.
CREATE TABLE IF NOT EXISTS user_orgs (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  org_id UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
  -- The role of the user in the org, one of 'owner', 'admin', 'member'.
  -- 'owner' has full control over the org, 'admin' can manage users and settings, 'member' has limited access.
  -- The default role is 'member'.
  nexeres_role VARCHAR(16) NOT NULL DEFAULT 'member' CHECK (
    nexeres_role IN ('owner', 'admin', 'member')
  ),
  -- The role_id references the roles table, which defines custom roles and their permissions.
  -- This allows for more granular control over user permissions in the org.
  role_id UUID REFERENCES roles(id) ON DELETE
  SET NULL,
    -- The date and time when the user joined the org.
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- 'banned' means the user is banned from the org, no sessions can be created for the user in the org.
    is_banned BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (user_id, org_id)
);

-- OAuth providers configuration table, per-org.
CREATE TABLE IF NOT EXISTS oauth_providers (
  id UUID PRIMARY KEY NOT NULL,
  org_id UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
  -- One of, 'google', 'github', 'microsoft', 'apple', or the name of a custom OAuth provider.
  -- This field is used to identify the provider in Nexeres.
  -- It must be unique across all providers for the org.
  -- Custom providers can be added by the org, but they must follow the same naming conventions.
  -- Custom providers must be registered with Nexeres before they can be used.
  provider VARCHAR(64) NOT NULL,
  -- The client ID of the OAuth provider, used to identify the application in the OAuth flow.
  client_id VARCHAR(512) NOT NULL,
  -- The client secret of the OAuth provider, used to authenticate the application in the OAuth flow.
  -- Stored as encrypted text in the database.
  client_secret VARCHAR(512) NOT NULL,
  -- The scopes requested by the application during the OAuth flow.
  scopes TEXT [] NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (org_id, provider)
);

-- User and OAuth identity association table. These are global, not per-org, which would have meant duplicate entries for the same user.
CREATE TABLE IF NOT EXISTS user_oauth_identities (
  id UUID PRIMARY KEY NOT NULL,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  provider VARCHAR(64) NOT NULL,
  -- The unique identifier for the user in the OAuth provider, used to link the user to their OAuth identity.
  provider_user_id VARCHAR(512) NOT NULL,
  -- The email address associated with the OAuth identity.
  provider_user_email VARCHAR(512) NOT NULL,
  provider_data JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, provider, provider_user_id)
);

-- MFA Factors table, per-user.
CREATE TABLE IF NOT EXISTS mfa_factors (
  id UUID PRIMARY KEY NOT NULL,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  -- The type of the MFA factor, one of 'totp', 'sms', 'email', 'webauthn'
  TYPE VARCHAR(16) NOT NULL,
  -- The user-friendly name of the MFA factor, used to identify the factor in the UI.
  name VARCHAR(255) NOT NULL,
  -- The secret, as follows for every type of factor:
  -- TOTP: Base32 secret key
  -- SMS: Phone number (+1234567890)
  -- Email: Email address
  -- WebAuthn: Public key data
  secret VARCHAR(512) NOT NULL,
  -- Whether the method is verified or not.
  verified BOOLEAN NOT NULL DEFAULT FALSE,
  last_used_at TIMESTAMPTZ DEFAULT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, TYPE, name)
);

-- Sessions table, scoped to orgs and users.
CREATE TABLE IF NOT EXISTS sessions (
  id UUID PRIMARY KEY NOT NULL,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  org_id UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
  -- The session token hash, used to authenticate the user in Nexeres.
  token_hash VARCHAR(512) UNIQUE NOT NULL,
  -- The refresh token hash corresponding to this session.
  refresh_token_hash VARCHAR(512) UNIQUE NOT NULL,
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

-- Verification tokens table, used for email verification, password reset, etc.
CREATE TABLE IF NOT EXISTS verification_tokens (
  id UUID PRIMARY KEY NOT NULL,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  -- The type of the verification token, one of 'email_verification', 'password_reset', etc.
  TYPE VARCHAR(64) NOT NULL,
  -- The hash of the token, sha256 hashed.
  token_hash BYTEA NOT NULL UNIQUE,
  -- The date and time when the token will expire.
  expires_at TIMESTAMPTZ NOT NULL,
  -- The date and time when the token was created.
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Scope table, for OIDC and API access.
CREATE TABLE IF NOT EXISTS scopes (
  id UUID PRIMARY KEY NOT NULL,
  -- The name of the scope, used to identify the scope in Nexeres.
  -- It must be unique across all scopes.
  -- Eg. 'drive:read', 'profile:read', etc.
  name VARCHAR(255) UNIQUE NOT NULL,
  -- The service this scope belongs to, usually used for NBRGLM's services, but can be used to customize the platform for your own services.
  -- For NBRGLM, this is may be something like 'drive', 'calendar', 'auth', etc.
  -- For custom services, this can be any string that identifies the service.
  service VARCHAR(255) NOT NULL,
  -- A user-friendly description of the scope, used to display the scope in the UI.
  description TEXT,
  -- This scope is enabled by default for all users and orgs and apps can access it without explicit consent.
  -- This is useful for scopes that are required for the basic functionality of Nexeres.
  -- For example, the 'profile:read' scope is enabled by default for all users and orgs.
  -- This includes scopes like 'openid' which are auto-granted to all clients.
  is_default BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO scopes (id, name, service, description, is_default)
VALUES (
    '01973b1d-a8e8-765a-be3e-ac91b12d2c55',
    'openid',
    'auth',
    'OpenID Connect authentication.',
    TRUE
  ),
  (
    '01973b1d-c799-7d94-8779-dc28f1c13e39',
    'profile:read',
    'auth',
    'Read your profile information such as avatar, or display name.',
    false
  ),
  (
    '01973b20-8c7e-72f5-8cf8-14fb93a03141',
    'profile:write',
    'auth',
    'Change your profile information such as avatar, or display name.',
    false
  ),
  (
    '01973b1d-f588-7226-809c-de006b41bbe4',
    'email:read',
    'auth',
    'Read your email address.',
    false
  ),
  (
    '01973b20-efc0-7337-804b-a5319e12a435',
    'email:write',
    'auth',
    'Change your email address.',
    FALSE
  ) ON CONFLICT (name) DO NOTHING;

-- OIDC Clients (For providing OIDC authentication to external applications)
CREATE TABLE IF NOT EXISTS oidc_clients (
  id UUID PRIMARY KEY NOT NULL,
  org_id UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
  -- The client ID of the OIDC client, used to identify the client in the OIDC flow.
  client_id VARCHAR(512) NOT NULL UNIQUE,
  -- The client secret of the OIDC client, used to authenticate the client in the OIDC flow.
  -- Stored after hashing.
  -- This is optional due to PKCE (Proof Key for Code Exchange) support, which allows clients to authenticate without a client secret.
  client_secret VARCHAR(512),
  -- The name of the OIDC client, used to identify the client in the UI.
  name VARCHAR(512) NOT NULL,
  -- The redirect URIs for the OIDC client, used to redirect the user back to the application after authentication.
  redirect_uris TEXT [] NOT NULL,
  -- The scopes requested by the OIDC client during the authentication flow.
  scopes TEXT [] DEFAULT ARRAY ['openid'],
  grant_types TEXT [] DEFAULT ARRAY ['authorization_code'],
  response_types TEXT [] DEFAULT ARRAY ['code'],
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- OIDC Auth Codes (With PKCE support)
CREATE TABLE IF NOT EXISTS oidc_auth_codes (
  id uuid PRIMARY KEY NOT NULL,
  code VARCHAR(512) NOT NULL UNIQUE,
  client_id UUID NOT NULL REFERENCES oidc_clients(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  -- This org id is used to identify the org that the client belongs to.
  -- Given a user is signing up, we need to make sure that once they authenticate, we check
  -- the org they belong to. If the user is not part of the org that has this oidc client, the login/signup will fail.
  -- If the user is part of the org, we will create a session for them in the org.
  org_id UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
  redirect_uri TEXT NOT NULL,
  scopes TEXT [] NOT NULL DEFAULT ARRAY ['openid'],
  nonce VARCHAR(512) NOT NULL,
  -- PKCE (Proof Key for Code Exchange) support.
  -- This is used to prevent authorization code interception attacks.
  -- The code challenge is a hashed version of the code verifier, which is sent in the token request.
  code_challenge VARCHAR(512) NOT NULL,
  -- The code challenge method is either 'S256' (SHA-256) or 'plain'.
  code_challenge_method VARCHAR(32) NOT NULL DEFAULT 'S256',
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- OIDC Access Tokens (always scoped to an org)
CREATE TABLE IF NOT EXISTS oidc_access_tokens (
  id UUID PRIMARY KEY NOT NULL,
  token_hash VARCHAR(512) NOT NULL UNIQUE,
  client_id UUID NOT NULL REFERENCES oidc_clients(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  org_id UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
  scopes TEXT [] NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- OIDC Refresh Tokens
CREATE TABLE IF NOT EXISTS oidc_refresh_tokens (
  id UUID PRIMARY KEY NOT NULL,
  token_hash VARCHAR(512) NOT NULL UNIQUE,
  access_token_id UUID NOT NULL REFERENCES oidc_access_tokens(id) ON DELETE CASCADE,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Invitations (scoped to an org)
CREATE TABLE IF NOT EXISTS invitations (
  id UUID PRIMARY KEY NOT NULL,
  org_id UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
  -- The email address of the user being invited to the org.
  email VARCHAR(512) NOT NULL,
  -- The role of the user in the org, one of 'owner', 'admin', 'member'.
  role VARCHAR(16) NOT NULL DEFAULT 'member',
  -- The user who created the invitation, NULL if the invitation was created by the system.
  invited_by UUID REFERENCES users(id) ON DELETE
  SET NULL,
    -- The token used to identify the invitation, used to accept the invitation.
    token VARCHAR(1024) NOT NULL UNIQUE,
    -- The date and time when the invitation will expire.
    expires_at TIMESTAMPTZ NOT NULL,
    -- The date and time when the invitation was accepted, NULL if the invitation has not been accepted yet.
    accepted_at TIMESTAMPTZ DEFAULT NULL,
    -- The date and time when the invitation was created.
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- The status of the invitation, one of 'pending', 'accepted', 'declined', 'expired'.
    STATUS VARCHAR(16) NOT NULL DEFAULT 'pending',
    UNIQUE (org_id, email)
);

-- Audit Log table for tracking changes in Nexeres.
CREATE TABLE IF NOT EXISTS audit_logs (
  id UUID PRIMARY KEY NOT NULL,
  org_id UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
  user_id UUID REFERENCES users(id) ON DELETE
  SET NULL,
    -- The action performed, such as 'login', 'logout', 'mfa_enabled', etc.
    ACTION VARCHAR(255) NOT NULL,
    -- Resource type, such as user, org, session, etc.
    resource_type VARCHAR(255) NOT NULL,
    resource_id UUID,
    ip_address INET,
    user_agent TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Essential indexes for Nexeres schema
-- orgs
CREATE INDEX idx_orgs_slug ON orgs(slug);

CREATE INDEX idx_orgs_domain ON org_domains(domain)
WHERE domain IS NOT NULL;

CREATE INDEX idx_orgs_deleted_at ON orgs(deleted_at)
WHERE deleted_at IS NULL;

-- Users
CREATE INDEX idx_users_email ON users(email)
WHERE deleted_at IS NULL;

CREATE INDEX idx_users_deleted_at ON users(deleted_at)
WHERE deleted_at IS NULL;

-- User orgs (junction table - heavily queried)
CREATE INDEX idx_user_orgs_user_id ON user_orgs(user_id);

CREATE INDEX idx_user_orgs_org_id ON user_orgs(org_id);

CREATE INDEX idx_user_orgs_status ON user_orgs(STATUS);

CREATE INDEX idx_user_orgs_role ON user_orgs(role);

-- OAuth Providers
CREATE INDEX idx_oauth_providers_org_id ON oauth_providers(org_id);

CREATE INDEX idx_oauth_providers_enabled ON oauth_providers(enabled)
WHERE enabled = TRUE;

-- User OAuth Identities
CREATE INDEX idx_user_oauth_identities_user_id ON user_oauth_identities(user_id);

CREATE INDEX idx_user_oauth_identities_provider ON user_oauth_identities(provider, provider_user_id);

CREATE INDEX idx_user_oauth_identities_email ON user_oauth_identities(provider_user_email);

-- MFA Factors
CREATE INDEX idx_mfa_factors_user_id ON mfa_factors(user_id);

CREATE INDEX idx_mfa_factors_type ON mfa_factors(TYPE);

CREATE INDEX idx_mfa_factors_verified ON mfa_factors(verified)
WHERE verified = TRUE;

-- Sessions (critical for auth performance)
CREATE INDEX idx_sessions_user_id ON sessions(user_id);

CREATE INDEX idx_sessions_org_id ON sessions(org_id);

CREATE INDEX idx_sessions_token_hash ON sessions(token_hash);

CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- Scopes
CREATE INDEX idx_scopes_name ON scopes(name);

CREATE INDEX idx_scopes_service ON scopes(service);

CREATE INDEX idx_scopes_default ON scopes(is_default)
WHERE is_default = TRUE;

-- OIDC Clients
CREATE INDEX idx_oidc_clients_org_id ON oidc_clients(org_id);

CREATE INDEX idx_oidc_clients_client_id ON oidc_clients(client_id);

-- OIDC Auth Codes (time-sensitive)
CREATE INDEX idx_oidc_auth_codes_code ON oidc_auth_codes(code);

CREATE INDEX idx_oidc_auth_codes_expires_at ON oidc_auth_codes(expires_at);

CREATE INDEX idx_oidc_auth_codes_client_id ON oidc_auth_codes(client_id);

CREATE INDEX idx_oidc_auth_codes_user_id ON oidc_auth_codes(user_id);

-- OIDC Access Tokens (frequently queried)
CREATE INDEX idx_oidc_access_tokens_token_hash ON oidc_access_tokens(token_hash);

CREATE INDEX idx_oidc_access_tokens_expires_at ON oidc_access_tokens(expires_at);

CREATE INDEX idx_oidc_access_tokens_user_id ON oidc_access_tokens(user_id);

CREATE INDEX idx_oidc_access_tokens_client_id ON oidc_access_tokens(client_id);

-- OIDC Refresh Tokens
CREATE INDEX idx_oidc_refresh_tokens_token_hash ON oidc_refresh_tokens(token_hash);

CREATE INDEX idx_oidc_refresh_tokens_expires_at ON oidc_refresh_tokens(expires_at);

CREATE INDEX idx_oidc_refresh_tokens_access_token_id ON oidc_refresh_tokens(access_token_id);

-- Invitations
CREATE INDEX idx_invitations_org_id ON invitations(org_id);

CREATE INDEX idx_invitations_email ON invitations(email);

CREATE INDEX idx_invitations_token ON invitations(token);

CREATE INDEX idx_invitations_expires_at ON invitations(expires_at);

CREATE INDEX idx_invitations_status ON invitations(STATUS);

-- Audit Logs (write-heavy, read by org/user)
CREATE INDEX idx_audit_logs_org_id ON audit_logs(org_id);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);

CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);

CREATE INDEX idx_audit_logs_action ON audit_logs(ACTION);

CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);

-- Composite indexes for common query patterns
CREATE INDEX idx_sessions_user_org ON sessions(user_id, org_id);

CREATE INDEX idx_user_orgs_user_status ON user_orgs(user_id, STATUS);

CREATE INDEX idx_invitations_org_email ON invitations(org_id, email);

CREATE INDEX idx_audit_logs_org_created ON audit_logs(org_id, created_at DESC);