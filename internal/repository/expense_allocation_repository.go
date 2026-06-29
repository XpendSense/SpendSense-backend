package repository

import (
	"context"

	"github.com/google/uuid"
	db "github.com/mauro-afa91/spendsense/internal/sqlc"
)

type ExpenseAllocationRepository interface {
	List(ctx context.Context, profileID uuid.UUID) ([]db.ExpenseAllocation, error)
	Upsert(ctx context.Context, arg db.UpsertExpenseAllocationParams) (db.ExpenseAllocation, error)
	Delete(ctx context.Context, arg db.DeleteExpenseAllocationParams) error
}

type expenseAllocationRepository struct {
	q *db.Queries
}

func NewExpenseAllocationRepository(q *db.Queries) ExpenseAllocationRepository {
	return &expenseAllocationRepository{q: q}
}

func (r *expenseAllocationRepository) List(ctx context.Context, profileID uuid.UUID) ([]db.ExpenseAllocation, error) {
	return r.q.ListExpenseAllocations(ctx, profileID)
}

func (r *expenseAllocationRepository) Upsert(ctx context.Context, arg db.UpsertExpenseAllocationParams) (db.ExpenseAllocation, error) {
	return r.q.UpsertExpenseAllocation(ctx, arg)
}

func (r *expenseAllocationRepository) Delete(ctx context.Context, arg db.DeleteExpenseAllocationParams) error {
	return r.q.DeleteExpenseAllocation(ctx, arg)
}
