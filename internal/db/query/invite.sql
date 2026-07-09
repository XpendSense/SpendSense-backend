-- name: CreateInvite :one
INSERT INTO budget_invite (budget_profile_id, email, role, invited_by, budget_person_id, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, budget_profile_id, email, role, token, status, invited_by, budget_person_id, expires_at, created_at;

-- name: GetInviteByToken :one
SELECT
    bi.id, bi.budget_profile_id, bi.email, bi.role, bi.token, bi.status,
    bi.invited_by, bi.budget_person_id, bi.expires_at, bi.created_at,
    bp.name  AS budget_name,
    COALESCE(u.first_name || ' ' || u.last_name, u.email) AS inviter_name
FROM budget_invite bi
JOIN budget_profile bp ON bp.id = bi.budget_profile_id
JOIN users u ON u.id = bi.invited_by
WHERE bi.token = $1
LIMIT 1;

-- name: GetInviteByID :one
SELECT id, budget_profile_id, email, role, token, status, invited_by, budget_person_id, expires_at, created_at
FROM budget_invite
WHERE id = $1
LIMIT 1;

-- name: ListInvitesByProfile :many
SELECT id, budget_profile_id, email, role, token, status, invited_by, budget_person_id, expires_at, created_at
FROM budget_invite
WHERE budget_profile_id = $1
ORDER BY created_at DESC;

-- name: UpdateInviteStatus :one
UPDATE budget_invite
SET status = $2
WHERE id = $1
RETURNING id, budget_profile_id, email, role, token, status, invited_by, budget_person_id, expires_at, created_at;
