package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mauro-afa91/spendsense/internal/apperr"
	"github.com/mauro-afa91/spendsense/internal/repository"
	db "github.com/mauro-afa91/spendsense/internal/sqlc"
)

type ExpenseAllocationService struct {
	allocations repository.ExpenseAllocationRepository
	profiles    repository.BudgetProfileRepository
}

func NewExpenseAllocationService(
	allocations repository.ExpenseAllocationRepository,
	profiles repository.BudgetProfileRepository,
) *ExpenseAllocationService {
	return &ExpenseAllocationService{allocations: allocations, profiles: profiles}
}

func (s *ExpenseAllocationService) assertOwner(ctx context.Context, profileID, userID uuid.UUID) error {
	profile, err := s.profiles.GetByID(ctx, profileID)
	if err != nil {
		return err
	}
	if profile.UserID != userID {
		return apperr.Forbidden("access denied")
	}
	return nil
}

func (s *ExpenseAllocationService) List(ctx context.Context, profileID, userID uuid.UUID) ([]db.ExpenseAllocation, error) {
	if err := s.assertOwner(ctx, profileID, userID); err != nil {
		return nil, err
	}
	return s.allocations.List(ctx, profileID)
}

func (s *ExpenseAllocationService) Upsert(ctx context.Context, profileID, userID uuid.UUID, categoryID int32, personID *int32, amount pgtype.Numeric) (db.ExpenseAllocation, error) {
	if err := s.assertOwner(ctx, profileID, userID); err != nil {
		return db.ExpenseAllocation{}, err
	}
	return s.allocations.Upsert(ctx, db.UpsertExpenseAllocationParams{
		BudgetProfileID: profileID,
		CategoryID:      categoryID,
		BudgetPersonID:  personID,
		PlannedAmount:   amount,
	})
}

func (s *ExpenseAllocationService) Delete(ctx context.Context, id int32, profileID, userID uuid.UUID) error {
	if err := s.assertOwner(ctx, profileID, userID); err != nil {
		return err
	}
	return s.allocations.Delete(ctx, db.DeleteExpenseAllocationParams{
		ID:              id,
		BudgetProfileID: profileID,
	})
}
