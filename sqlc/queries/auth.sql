-- name: CreateUser :exec
INSERT INTO users (
    id,
    email,
    name,
    password_hash,
    avatar_url,
    ext_id
  )
VALUES (
    $1,
    $2,
    $3,
    sqlc.narg('password_hash'),
    sqlc.narg('avatar_url'),
    sqlc.narg('ext_id')
  );

-- name: UpdateUser :one
UPDATE users
SET name = coalesce(sqlc.narg('name'), name),
  avatar_url = coalesce(sqlc.narg('avatar_url'), avatar_url),
  ext_id = coalesce(sqlc.narg('ext_id'), ext_id)
WHERE (
    sqlc.narg('id')::uuid IS NOT NULL
    AND id = sqlc.narg('id')::uuid
  )
  OR (
    sqlc.narg('id') IS NULL
    AND email = sqlc.narg('email')
  )
RETURNING id,
  email,
  email_verified,
  mfa_enabled,
  name,
  avatar_url,
  created_at,
  updated_at;

-- name: MarkUserEmailVerified :exec
UPDATE users
SET email_verified = TRUE
WHERE (
    (
      sqlc.narg('id')::uuid IS NOT NULL
      AND id = sqlc.narg('id')::uuid
    )
    OR (
      sqlc.narg('id') IS NULL
      AND email = sqlc.narg('email')
    )
  );

-- name: UpdateMFA :exec
UPDATE users
SET mfa_enabled = sqlc.arg('mfa_enabled'),
  backup_codes = coalesce(sqlc.narg('backup_codes'), '[]'::jsonb)
WHERE (
    (
      sqlc.narg('id')::uuid IS NOT NULL
      AND id = sqlc.narg('id')::uuid
    )
    OR (
      sqlc.narg('id') IS NULL
      AND email = sqlc.narg('email')
    )
  );

-- name: AddMFAFactor :exec
INSERT INTO mfa_factors (id, user_id, factor_type, name, secret, verified)
VALUES ($1, $2, $3, $4, $5, sqlc.narg('verified'));

-- name: GetMFAFactors :many
SELECT *
FROM mfa_factors
WHERE user_id = sqlc.arg('user_id')
  AND (
    sqlc.narg('verified')::BOOLEAN IS NULL
    OR verified = sqlc.narg('verified')::BOOLEAN
  )
  AND (
    sqlc.narg('factor_type')::mfa_type IS NULL
    OR factor_type = sqlc.narg('factor_type')::mfa_type
  )
  AND (
    sqlc.narg('secret')::VARCHAR IS NULL
    OR secret = sqlc.narg('secret')::VARCHAR
  );

-- name: DeleteMFAFactor :exec
DELETE FROM mfa_factors
WHERE id = sqlc.arg('id')
  AND user_id = sqlc.arg('user_id');

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = sqlc.arg('password_hash')
WHERE (
    (
      sqlc.narg('id')::uuid IS NOT NULL
      AND id = sqlc.narg('id')::uuid
    )
    OR (
      sqlc.narg('id') IS NULL
      AND email = sqlc.narg('email')
    )
  );

-- name: GetLoginInfoForUser :one
SELECT id,
  email,
  email_verified,
  mfa_enabled,
  password_hash,
  name,
  avatar_url,
  ext_id,
  created_at,
  updated_at
FROM users
WHERE (
    (
      sqlc.narg('id')::uuid IS NOT NULL
      AND id = sqlc.narg('id')::uuid
    )
    OR (
      sqlc.narg('id') IS NULL
      AND email = sqlc.narg('email')
    )
  );

-- name: GetInfoForSessionRefresh :one
SELECT o.slug AS org_slug,
  o.name AS org_name,
  o.avatar_url AS org_avatar_url,
  u.email AS user_email,
  u.email_verified AS user_email_verified,
  u.mfa_enabled AS user_mfa_enabled,
  uo.is_org_admin AS user_is_org_admin,
  u.name AS user_name,
  u.avatar_url AS user_avatar_url,
  r.role_name AS user_org_role,
  r.id AS user_org_role_id,
  -- Select role permissions, if role is NULL (no custom role assigned), return empty array
  -- only for caching permissions in redis, not to include in the JWT
  COALESCE(r.permissions, '[]'::jsonb) AS user_org_role_permissions
FROM users u
  INNER JOIN user_orgs uo ON u.id = uo.user_id
  INNER JOIN orgs o ON o.id = uo.org_id
  LEFT JOIN roles r ON r.id = uo.role_id
WHERE (
    (
      sqlc.narg('user_id') IS NOT NULL
      AND u.id = sqlc.narg('user_id')
    )
    OR (
      sqlc.narg('user_id') IS NULL
      AND u.email = sqlc.narg('user_email')
    )
  )
  AND (
    (
      sqlc.narg('org_id') IS NOT NULL
      AND o.id = sqlc.narg('org_id')
    )
    OR (
      sqlc.narg('org_id') IS NULL
      AND o.slug = sqlc.narg('org_slug')
    )
  )
  AND uo.is_banned IS NOT TRUE;

-- name: GetUser :one
SELECT id,
  email,
  email_verified,
  mfa_enabled,
  name,
  avatar_url,
  ext_id,
  created_at,
  updated_at
FROM users
WHERE (
    sqlc.narg('id')::uuid IS NOT NULL
    AND id = sqlc.narg('id')::uuid
  )
  OR (
    sqlc.narg('id') IS NULL
    AND email = sqlc.narg('email')
  );

-- name: DeleteUserById :exec
DELETE FROM users
WHERE id = $1;

-- name: DeleteUserByEmail :exec
DELETE FROM users
WHERE email = $1;

-- name: CreateOrg :exec
INSERT INTO orgs (
    id,
    slug,
    name,
    description,
    avatar_url,
    settings
  )
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
  );

-- name: UpdateOrg :one
UPDATE orgs
SET name = coalesce(sqlc.narg('name'), name),
  description = coalesce(sqlc.narg('description'), description),
  avatar_url = coalesce(sqlc.narg('avatar_url'), avatar_url),
  settings = coalesce(sqlc.narg('settings'), settings)
WHERE (
    sqlc.narg('id')::uuid IS NOT NULL
    AND id = sqlc.narg('id')::uuid
  )
  OR (
    sqlc.narg('id') IS NULL
    AND slug = sqlc.narg('slug')
  )
RETURNING orgs.name,
  orgs.description,
  orgs.avatar_url,
  orgs.settings;

-- name: GetOrg :one
SELECT *
FROM orgs
WHERE (
    sqlc.narg('id')::uuid IS NOT NULL
    AND id = sqlc.narg('id')::uuid
  )
  OR (
    sqlc.narg('id') IS NULL
    AND slug = sqlc.narg('slug')
  );

-- name: DeleteOrgById :one
DELETE FROM orgs
WHERE id = $1
RETURNING *;

-- name: DeleteOrgBySlug :one
DELETE FROM orgs
WHERE slug = $1
RETURNING *;

-- name: GetOrgByDomain :one
SELECT sqlc.embed(o),
  od.auto_join_role_id AS auto_join_role_id
FROM orgs o
  INNER JOIN org_domains od ON o.id = od.org_id
WHERE od.domain = sqlc.arg('domain')
  AND (
    sqlc.narg('verified')::BOOLEAN IS NULL
    OR od.verified = sqlc.narg('verified')::BOOLEAN
  )
  AND (
    sqlc.narg('auto_join')::BOOLEAN IS NULL
    OR od.auto_join = sqlc.narg('auto_join')::BOOLEAN
  );

-- name: CreateOrgDomain :one
INSERT INTO org_domains (
    domain,
    org_id,
    auto_join,
    auto_join_role_id,
    auto_join_role_name
  )
VALUES (
    sqlc.arg('domain'),
    sqlc.arg('org_id'),
    sqlc.narg('auto_join'),
    sqlc.narg('auto_join_role_id'),
    sqlc.narg('auto_join_role_name')
  )
RETURNING *;

-- name: UpdateOrgDomain :one
UPDATE org_domains
SET auto_join = coalesce(sqlc.narg('auto_join'), auto_join),
  auto_join_role_id = coalesce(
    sqlc.narg('auto_join_role_id'),
    auto_join_role_id
  ),
  auto_join_role_name = coalesce(
    sqlc.narg('auto_join_role_name'),
    auto_join_role_name
  )
WHERE domain = sqlc.arg('domain')
  AND org_id = sqlc.arg('org_id')
RETURNING *;

-- name: UpdateOrgDomainVerification :exec
UPDATE org_domains
SET verified = sqlc.arg('verified')
WHERE domain = sqlc.arg('domain')
  AND org_id = sqlc.arg('org_id');

-- name: RemoveDomainFromOrg :one
DELETE FROM org_domains
WHERE org_id = sqlc.arg('org_id')
  AND domain = sqlc.arg('domain')
RETURNING *;

-- name: GetOrgById :one
SELECT *
FROM orgs
WHERE id = sqlc.arg('org_id');

-- name: GetOrgDomains :many
SELECT *
FROM org_domains
WHERE org_id = sqlc.arg('org_id');

-- name: GetOrgDomain :one
SELECT *
FROM org_domains
WHERE domain = $1;

-- name: CreateRole :one
INSERT INTO roles (id, org_id, role_name, permissions, role_desc)
VALUES ($1, $2, $3, $4, sqlc.narg('role_desc'))
RETURNING *;

-- name: UpdateRole :one
UPDATE roles
SET role_name = sqlc.arg('role_name'),
  permissions = sqlc.arg('permissions'),
  role_desc = sqlc.arg('role_desc')
WHERE id = sqlc.arg('role_id')
  AND org_id = sqlc.arg('org_id')
RETURNING *;

-- name: GetRolesForOrg :many
SELECT *
FROM roles
WHERE org_id = sqlc.arg('org_id');

-- name: GetMinimalRolesForOrg :many
SELECT id,
  role_name,
  role_desc,
  created_at,
  updated_at
FROM roles
WHERE org_id = sqlc.arg('org_id');

-- name: DeleteRole :one
DELETE FROM roles
WHERE id = sqlc.arg('role_id')
  AND org_id = sqlc.arg('org_id')
RETURNING *;

-- name: AddUserToOrg :exec
INSERT INTO user_orgs (user_id, org_id, role_id, is_org_admin)
VALUES (
    sqlc.arg('user_id'),
    sqlc.arg('org_id'),
    sqlc.narg('role_id'),
    sqlc.narg('is_org_admin')
  ) ON CONFLICT (user_id, org_id) DO NOTHING;

-- name: UpdateUserRoleInOrg :exec
UPDATE user_orgs
SET role_id = coalesce(sqlc.narg('role_id'), role_id),
  is_org_admin = coalesce(sqlc.narg('is_org_admin'), is_org_admin)
WHERE user_id = sqlc.arg('user_id')
  AND org_id = sqlc.arg('org_id');

-- name: RemoveCustomUserRoleInOrg :exec
UPDATE user_orgs
SET role_id = NULL
WHERE user_id = sqlc.arg('user_id')
  AND org_id = sqlc.arg('org_id');

-- name: UpdateUserBanInOrg :exec
UPDATE user_orgs
SET is_banned = sqlc.arg('is_banned')
WHERE user_id = sqlc.arg('user_id')
  AND org_id = sqlc.arg('org_id');

-- name: RemoveUserFromOrg :exec
DELETE FROM user_orgs
WHERE user_id = sqlc.arg('user_id')
  AND org_id = sqlc.arg('org_id');

-- name: GetUserOrgs :many
SELECT sqlc.embed(o),
  sqlc.embed(uo),
  sqlc.embed(r)
FROM orgs o
  INNER JOIN user_orgs uo ON o.id = uo.org_id
  INNER JOIN roles r ON r.id = uo.role_id
  INNER JOIN users u ON u.id = uo.user_id
WHERE (
    (
      sqlc.narg('user_id') IS NOT NULL
      AND uo.user_id = sqlc.arg('user_id')
    )
    OR (
      sqlc.narg('user_id') IS NULL
      AND u.email = sqlc.narg('user_email')
    )
  )
  AND uo.is_banned != TRUE;

-- name: CreateSession :exec
INSERT INTO sessions (
    id,
    user_id,
    org_id,
    session_token_hash,
    refresh_token_hash,
    ip_address,
    user_agent,
    expires_at,
    mfa_verified,
    mfa_verified_at
  )
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    sqlc.narg('mfa_verified'),
    sqlc.narg('mfa_verified_at')
  );

-- name: UpdateSessionMFA :exec
UPDATE sessions
SET mfa_verified = sqlc.arg('mfa_verified'),
  mfa_verified_at = CASE
    WHEN sqlc.arg('mfa_verified') = TRUE THEN NOW()
    ELSE NULL
  END
WHERE (
    (
      sqlc.narg('id')::uuid IS NOT NULL
      AND id = sqlc.narg('id')::uuid
    )
    OR (
      sqlc.narg('id') IS NULL
      AND refresh_token_hash = sqlc.narg('refresh_token_hash')
    )
  )
  AND expires_at > NOW();

-- name: GetSession :one
SELECT *
FROM sessions
WHERE (
    (
      sqlc.narg('id')::uuid IS NOT NULL
      AND id = sqlc.narg('id')::uuid
    )
    OR (
      sqlc.narg('id') IS NULL
      AND refresh_token_hash = sqlc.narg('refresh_token_hash')
    )
  )
  AND expires_at > NOW();

-- name: RefreshSession :exec
UPDATE sessions
SET session_token_hash = sqlc.arg('session_token_hash'),
  refresh_token_hash = sqlc.arg('refresh_token_hash'),
  expires_at = sqlc.arg('expires_at'),
  user_agent = coalesce(sqlc.narg('user_agent'), user_agent),
  ip_address = coalesce(sqlc.narg('ip_address'), ip_address)
WHERE (
    (
      sqlc.narg('id')::uuid IS NOT NULL
      AND id = sqlc.narg('id')::uuid
    )
    OR (
      sqlc.narg('id') IS NULL
      AND refresh_token_hash = sqlc.narg('old_refresh_token_hash')
    )
  )
  AND expires_at > NOW();

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE (
    (
      sqlc.narg('id')::uuid IS NOT NULL
      AND id = sqlc.narg('id')::uuid
    )
    OR (
      sqlc.narg('id') IS NULL
      AND refresh_token_hash = sqlc.narg('refresh_token_hash')
    )
  );

-- name: GetSessionForUser :many
SELECT *
FROM sessions
WHERE user_id = sqlc.arg('user_id')
  AND expires_at > NOW();

-- name: GetSessionsForOrg :many
SELECT *
FROM sessions
WHERE org_id = sqlc.arg('org_id')
  AND expires_at > NOW();

-- name: GetSessionsForUserInOrg :many
SELECT *
FROM sessions
WHERE user_id = sqlc.arg('user_id')
  AND org_id = sqlc.arg('org_id')
  AND expires_at > NOW();

-- name: CreateInvitation :exec
INSERT INTO invitations (
    id,
    org_id,
    email,
    invited_by,
    token_hash,
    expires_at,
    role_id,
    invite_as_admin
  )
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    sqlc.narg('role_id'),
    sqlc.narg('invite_as_admin')
  );

-- name: GetInvitation :one
SELECT *
FROM invitations
WHERE (
    (
      sqlc.narg('id')::uuid IS NOT NULL
      AND id = sqlc.narg('id')::uuid
    )
    OR (
      sqlc.narg('id') IS NULL
      AND token_hash = sqlc.narg('token_hash')
    )
  )
  AND expires_at > NOW()
  AND invite_status = ANY(sqlc.arg('statuses')::invitation_status []);

-- name: RevokeInvitation :exec
DELETE FROM invitations
WHERE (
    (
      sqlc.narg('id')::uuid IS NOT NULL
      AND id = sqlc.narg('id')::uuid
    )
    OR (
      sqlc.narg('id') IS NULL
      AND email = sqlc.narg('email')
    )
  );

-- name: CreateVerificationToken :exec
INSERT INTO verification_tokens (
    id,
    user_id,
    email,
    token_type,
    token_hash,
    expires_at
  )
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetVerificationToken :one
SELECT *
FROM verification_tokens
WHERE token_hash = sqlc.arg('token_hash')
  AND token_type = sqlc.arg('token_type')
  AND expires_at > NOW();

-- name: CreateAuditLog :exec
INSERT INTO audit_logs (
    id,
    actor_type,
    log_action,
    log_entity,
    details,
    org_id,
    user_id,
    entity_id,
    ip_address,
    user_agent
  )
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);