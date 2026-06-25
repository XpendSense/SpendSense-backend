-- name: ListBudgetsByUser :many
SELECT id, user_id, name, start_date, end_date, active
FROM budget
WHERE user_id = $1
ORDER BY active DESC, name;

-- name: GetBudgetByID :one
SELECT id, user_id, name, start_date, end_date, active
FROM budget
WHERE id = $1
LIMIT 1;

-- name: ExistsBudgetByNameAndUser :one
SELECT EXISTS (
    SELECT 1 FROM budget WHERE name = $1 AND user_id = $2
) AS exists;

-- name: CreateBudget :one
INSERT INTO budget (user_id, name)
VALUES ($1, $2)
RETURNING id, user_id, name, start_date, end_date, active;

-- name: UpdateBudget :one
UPDATE budget
SET name = $2, active = $3
WHERE id = $1
RETURNING id, user_id, name, start_date, end_date, active;

-- name: DeleteBudget :exec
DELETE FROM budget
WHERE id = $1;

-- name: ListBudgetPeople :many
SELECT id, budget_id, user_name, user_id, is_active
FROM budget_to_user_mapping
WHERE budget_id = $1 AND is_active = TRUE
ORDER BY id;

-- name: GetBudgetPersonByID :one
SELECT id, budget_id, user_name, user_id, is_active
FROM budget_to_user_mapping
WHERE id = $1 AND budget_id = $2
LIMIT 1;

-- name: ExistsBudgetPerson :one
SELECT EXISTS (
    SELECT 1 FROM budget_to_user_mapping WHERE budget_id = $1 AND user_name = $2 AND is_active = TRUE
) AS exists;

-- name: AddBudgetPerson :one
INSERT INTO budget_to_user_mapping (budget_id, user_name, user_id)
VALUES ($1, $2, $3)
RETURNING id, budget_id, user_name, user_id, is_active;

-- name: SoftRemovePersonAndReassign :exec
WITH reassign_transactions AS (
    UPDATE transaction
    SET payment_method_id = sqlc.arg('replacement_pm_id')::uuid
    WHERE payment_method_id IN (
        SELECT id FROM payment_methods WHERE budget_person_id = sqlc.arg('person_id')
    )
      AND budget_id = sqlc.arg('budget_id')::uuid
),
soft_delete_pms AS (
    UPDATE payment_methods
    SET is_active = FALSE
    WHERE budget_person_id = sqlc.arg('person_id')
),
reassign_income AS (
    UPDATE income_to_budget_mapping
    SET budget_person_id = sqlc.arg('replacement_person_id')
    WHERE budget_person_id = sqlc.arg('person_id') AND budget_id = sqlc.arg('budget_id')::uuid
)
UPDATE budget_to_user_mapping
SET is_active = FALSE
WHERE budget_to_user_mapping.id = sqlc.arg('person_id') AND budget_to_user_mapping.budget_id = sqlc.arg('budget_id')::uuid;

-- name: SoftRemovePerson :exec
WITH soft_delete_pms AS (
    UPDATE payment_methods
    SET is_active = FALSE
    WHERE budget_person_id = sqlc.arg('person_id')
)
UPDATE budget_to_user_mapping
SET is_active = FALSE
WHERE budget_to_user_mapping.id = sqlc.arg('person_id') AND budget_to_user_mapping.budget_id = sqlc.arg('budget_id')::uuid;

-- name: ListIncomeEntries :many
SELECT id, budget_id, user_id, name, amount, recurring, budget_person_id
FROM income_to_budget_mapping
WHERE budget_id = $1
ORDER BY id;

-- name: AddIncomeEntry :one
INSERT INTO income_to_budget_mapping (budget_id, name, amount, recurring, budget_person_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, budget_id, user_id, name, amount, recurring, budget_person_id;

-- name: UpdateIncomeEntry :one
UPDATE income_to_budget_mapping
SET name = $3, amount = $4, recurring = $5, budget_person_id = $6
WHERE id = $1 AND budget_id = $2
RETURNING id, budget_id, user_id, name, amount, recurring, budget_person_id;

-- name: DeleteIncomeEntry :exec
DELETE FROM income_to_budget_mapping
WHERE id = $1 AND budget_id = $2;
