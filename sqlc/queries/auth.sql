-- name: CreateUser :exec
INSERT INTO users (
    id,
    email,
    first_name,
    last_name,
    password_hash,
    middle_name,
    avatar_url
  )
VALUES (
    $1,
    $2,
    $3,
    $4,
    sqlc.narg('password_hash'),
    sqlc.narg('middle_name'),
    sqlc.narg('avatar_url')
  );

-- name: UpdateUser :one
UPDATE users
SET first_name = coalesce(sqlc.narg('first_name'), first_name),
  last_name = coalesce(sqlc.narg('last_name'), last_name),
  middle_name = coalesce(sqlc.narg('middle_name'), middle_name),
  avatar_url = coalesce(sqlc.narg('avatar_url'), avatar_url)
WHERE (
    (
      sqlc.narg('id') IS NOT NULL
      AND id = sqlc.narg('id')
    )
    OR (
      sqlc.narg('id') IS NULL
      AND email = sqlc.narg('email')
    )
  )
  AND deleted_at IS NULL
RETURNING id,
  email,
  email_verified,
  mfa_enabled,
  first_name,
  middle_name,
  last_name,
  avatar_url,
  created_at,
  updated_at;

-- name: MarkUserEmailVerified :exec
UPDATE users
SET email_verified = TRUE
WHERE (
    (
      sqlc.narg('id') IS NOT NULL
      AND id = sqlc.narg('id')
    )
    OR (
      sqlc.narg('id') IS NULL
      AND email = sqlc.narg('email')
    )
  )
  AND deleted_at IS NULL;

-- name: UpdateMFA :exec
UPDATE users
SET mfa_enabled = sqlc.arg('mfa_enabled'),
  backup_codes = coalesce(sqlc.narg('backup_codes'), '[]'::jsonb)
WHERE (
    (
      sqlc.narg('id') IS NOT NULL
      AND id = sqlc.narg('id')
    )
    OR (
      sqlc.narg('id') IS NULL
      AND email = sqlc.narg('email')
    )
  )
  AND deleted_at IS NULL;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = sqlc.arg('password_hash')
WHERE (
    (
      sqlc.narg('id') IS NOT NULL
      AND id = sqlc.narg('id')
    )
    OR (
      sqlc.narg('id') IS NULL
      AND email = sqlc.narg('email')
    )
  )
  AND deleted_at IS NULL;

-- name: GetLoginInfoForUser :one
SELECT id,
  email,
  email_verified,
  mfa_enabled,
  password_hash,
  first_name,
  middle_name,
  last_name,
  avatar_url,
  created_at,
  updated_at
FROM users
WHERE (
    (
      sqlc.narg('id') IS NOT NULL
      AND id = sqlc.narg('id')
    )
    OR (
      sqlc.narg('id') IS NULL
      AND email = sqlc.narg('email')
    )
  )
  AND deleted_at IS NULL;

-- name: GetInfoForSessionRefresh :one
SELECT u.first_name AS user_fname,
  u.middle_name AS user_mname,
  u.last_name AS user_lname,
  u.email AS user_email,
  u.email_verified AS user_email_verified,
  u.avatar_url AS user_avatar_url,
  u.mfa_enabled AS user_mfa_enabled,
  o.name AS org_name,
  o.slug AS org_slug,
  o.avatar_url AS org_avatar_url,
  uo.is_org_admin AS user_is_org_admin,
  r.role_name AS user_org_role,
  r.permissions AS user_role_permissions
FROM users u
  INNER JOIN user_orgs uo ON u.id = uo.user_id
  INNER JOIN orgs o ON o.id = uo.org_id
  INNER JOIN roles r ON r.id = uo.role_id
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
  AND uo.is_banned != TRUE
  AND u.deleted_at IS NULL
  AND o.deleted_at IS NULL;

-- name: GetUser :one
SELECT id,
  email,
  email_verified,
  mfa_enabled,
  first_name,
  middle_name,
  last_name,
  avatar_url,
  created_at,
  updated_at
FROM users
WHERE (
    (
      sqlc.narg('id') IS NOT NULL
      AND id = sqlc.narg('id')
    )
    OR (
      sqlc.narg('id') IS NULL
      AND email = sqlc.narg('email')
    )
  )
  AND deleted_at IS NULL;

-- name: SoftDeleteUser :exec
UPDATE users
SET deleted_at = NOW()
WHERE (
    (
      sqlc.narg('id') IS NOT NULL
      AND id = sqlc.narg('id')
    )
    OR (
      sqlc.narg('id') IS NULL
      AND email = sqlc.narg('email')
    )
  )
  AND deleted_at IS NULL;

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
    sqlc.narg('description'),
    sqlc.narg('avatar_url'),
    sqlc.narg('settings')
  );

-- name: UpdateOrg :exec
UPDATE orgs
SET name = coalesce(sqlc.narg('name'), name),
  description = coalesce(sqlc.narg('description'), description),
  avatar_url = coalesce(sqlc.narg('avatar_url'), avatar_url),
  settings = coalesce(sqlc.narg('settings'), settings)
WHERE (
    (
      sqlc.narg('id') IS NOT NULL
      AND id = sqlc.arg('id')
    )
    OR (
      sqlc.narg('id') IS NULL
      AND slug = sqlc.narg('slug')
    )
  )
  AND deleted_at IS NULL;

-- name: GetOrg :one
SELECT id,
  slug,
  name,
  description,
  avatar_url,
  settings,
  created_at,
  updated_at
FROM orgs
WHERE (
    (
      sqlc.narg('id') IS NOT NULL
      AND id = sqlc.narg('id')
    )
    OR (
      sqlc.narg('id') IS NULL
      AND slug = sqlc.narg('slug')
    )
  )
  AND deleted_at IS NULL;

-- name: SoftDeleteOrg :exec
UPDATE orgs
SET deleted_at = NOW()
WHERE (
    (
      sqlc.narg('id') IS NOT NULL
      AND id = sqlc.narg('id')
    )
    OR (
      sqlc.narg('id') IS NULL
      AND slug = sqlc.narg('slug')
    )
  )
  AND deleted_at IS NULL;

-- name: GetOrgByDomain :one
SELECT o.id,
  o.slug,
  o.name,
  o.description,
  o.avatar_url,
  o.settings,
  o.created_at,
  o.updated_at
FROM orgs o
  INNER JOIN org_domains od ON o.id = od.org_id
WHERE od.domain = sqlc.arg('domain')
  AND od.verified = coalesce(sqlc.narg('verified'), FALSE)
  AND od.auto_join = coalesce(sqlc.narg('auto_join'), FALSE)
  AND o.deleted_at IS NULL;

-- name: AddDomainToOrg :exec
INSERT INTO org_domains (org_id, domain)
VALUES (sqlc.arg('org_id'), sqlc.arg('domain')) ON CONFLICT (org_id, domain) DO NOTHING;

-- RemoveDomainFromOrg :exec
DELETE FROM org_domains
WHERE org_id = sqlc.arg('org_id')
  AND domain = sqlc.arg('domain');

-- name: CreateRole :exec
INSERT INTO roles (id, org_id, role_name, permissions)
VALUES ($1, $2, $3, $4);

-- name: GetRolesForOrg :many
SELECT *
FROM roles
WHERE org_id = sqlc.arg('org_id');

-- name: DeleteRole :exec
DELETE FROM roles
WHERE id = sqlc.arg('id')
  AND org_id = sqlc.arg('org_id');

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
  AND o.deleted_at IS NULL
  AND u.deleted_at IS NULL
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
      sqlc.narg('id') IS NOT NULL
      AND id = sqlc.narg('id')
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
      sqlc.narg('id') IS NOT NULL
      AND id = sqlc.narg('id')
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
      sqlc.narg('id') IS NOT NULL
      AND id = sqlc.narg('id')
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
      sqlc.narg('id') IS NOT NULL
      AND id = sqlc.narg('id')
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
    invite_as_admin
  )
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    sqlc.narg('invite_as_admin')
  );

-- name: GetInvitation :one
SELECT *
FROM invitations
WHERE (
    (
      sqlc.narg('id') IS NOT NULL
      AND id = sqlc.narg('id')
    )
    OR (
      sqlc.narg('id') IS NULL
      AND token_hash = sqlc.narg('token_hash')
    )
  )
  AND expires_at > NOW()
  AND invite_status = ANY(sqlc.arg('statuses'));

-- name: RevokeInvitation :exec
DELETE FROM invitations
WHERE (
    (
      sqlc.narg('id') IS NOT NULL
      AND id = sqlc.narg('id')
    )
    OR (
      sqlc.narg('id') IS NULL
      AND email = sqlc.narg('email')
    )
  );

-- name: CreateVerificationToken :exec
INSERT INTO verification_tokens (id, user_id, token_type, token_hash, expires_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetVerificationToken :one
SELECT *
FROM verification_tokens
WHERE token_hash = sqlc.arg('token_hash')
  AND token_type = sqlc.arg('token_type')
  AND expires_at > NOW();