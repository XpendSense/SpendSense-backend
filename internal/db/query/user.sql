-- name: GetUserByID :one
SELECT id, email, hashed_password, first_name, last_name, is_active, is_superuser, is_verified, created_at
FROM users
WHERE id = $1
LIMIT 1;

-- name: GetUserByEmail :one
SELECT id, email, hashed_password, first_name, last_name, is_active, is_superuser, is_verified, created_at
FROM users
WHERE email = $1
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (email, hashed_password, first_name, last_name)
VALUES ($1, $2, $3, $4)
RETURNING id, email, hashed_password, first_name, last_name, is_active, is_superuser, is_verified, created_at;

-- name: UpdateUser :one
UPDATE users
SET first_name = $2, last_name = $3
WHERE id = $1
RETURNING id, email, hashed_password, first_name, last_name, is_active, is_superuser, is_verified, created_at;

-- name: UpdateUserPassword :exec
UPDATE users
SET hashed_password = $2
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: GetOAuthAccount :one
SELECT id, user_id, oauth_name, account_id, account_email
FROM oauth_account
WHERE oauth_name = $1 AND account_id = $2
LIMIT 1;

-- name: CreateOAuthAccount :one
INSERT INTO oauth_account (user_id, oauth_name, account_id, account_email)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, oauth_name, account_id, account_email;
