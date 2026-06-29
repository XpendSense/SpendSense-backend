-- name: ListExpenseAllocations :many
SELECT id, budget_profile_id, category_id, budget_person_id, planned_amount
FROM expense_allocation
WHERE budget_profile_id = $1
ORDER BY category_id, id;

-- name: UpsertExpenseAllocation :one
INSERT INTO expense_allocation (budget_profile_id, category_id, budget_person_id, planned_amount)
VALUES (sqlc.arg('budget_profile_id')::uuid, sqlc.arg('category_id'), sqlc.arg('budget_person_id'), sqlc.arg('planned_amount'))
ON CONFLICT (budget_profile_id, category_id, COALESCE(budget_person_id, -1))
DO UPDATE SET planned_amount = EXCLUDED.planned_amount
RETURNING id, budget_profile_id, category_id, budget_person_id, planned_amount;

-- name: DeleteExpenseAllocation :exec
DELETE FROM expense_allocation WHERE id = $1 AND budget_profile_id = $2;
