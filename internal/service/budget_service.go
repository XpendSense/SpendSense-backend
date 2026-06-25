package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mauro-afa91/spendsense/internal/apperr"
	"github.com/mauro-afa91/spendsense/internal/repository"
	db "github.com/mauro-afa91/spendsense/internal/sqlc"
)

type BudgetService struct {
	budgets repository.BudgetRepository
	users   repository.UserRepository
}

func NewBudgetService(budgets repository.BudgetRepository, users repository.UserRepository) *BudgetService {
	return &BudgetService{budgets: budgets, users: users}
}

func (s *BudgetService) List(ctx context.Context, userID uuid.UUID) ([]db.Budget, error) {
	return s.budgets.ListByUserID(ctx, userID)
}

func (s *BudgetService) Get(ctx context.Context, id, userID uuid.UUID) (db.Budget, error) {
	budget, err := s.budgets.GetByID(ctx, id)
	if err != nil {
		return db.Budget{}, err
	}
	if budget.UserID != userID {
		return db.Budget{}, apperr.Forbidden("access denied")
	}
	return budget, nil
}

func (s *BudgetService) Create(ctx context.Context, userID uuid.UUID, name string) (db.Budget, error) {
	exists, err := s.budgets.ExistsByNameAndUser(ctx, name, userID)
	if err != nil {
		return db.Budget{}, fmt.Errorf("budget: check exists: %w", err)
	}
	if exists {
		return db.Budget{}, apperr.Duplicate("budget", "name", name)
	}
	budget, err := s.budgets.Create(ctx, db.CreateBudgetParams{UserID: userID, Name: name})
	if err != nil {
		return db.Budget{}, err
	}
	// Auto-add budget owner as the first BudgetPerson.
	owner, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return budget, nil // non-fatal: budget was created
	}
	parts := []string{}
	if owner.FirstName != nil {
		parts = append(parts, *owner.FirstName)
	}
	if owner.LastName != nil {
		parts = append(parts, *owner.LastName)
	}
	displayName := strings.Join(parts, " ")
	if displayName == "" {
		displayName = owner.Email
	}
	_, _ = s.budgets.AddPerson(ctx, db.AddBudgetPersonParams{
		BudgetID: budget.ID,
		UserName: &displayName,
		UserID:   &userID,
	})
	return budget, nil
}

func (s *BudgetService) Update(ctx context.Context, id, userID uuid.UUID, name string, active bool) (db.Budget, error) {
	if _, err := s.Get(ctx, id, userID); err != nil {
		return db.Budget{}, err
	}
	return s.budgets.Update(ctx, db.UpdateBudgetParams{ID: id, Name: name, Active: active})
}

func (s *BudgetService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if _, err := s.Get(ctx, id, userID); err != nil {
		return err
	}
	return s.budgets.Delete(ctx, id)
}

// ── People ────────────────────────────────────────────────────────────────────

type PersonInput struct {
	UserName string
	UserID   *uuid.UUID
}

func (s *BudgetService) AddPeople(ctx context.Context, budgetID, userID uuid.UUID, people []PersonInput) ([]db.BudgetToUserMapping, error) {
	if _, err := s.Get(ctx, budgetID, userID); err != nil {
		return nil, err
	}
	var results []db.BudgetToUserMapping
	for _, p := range people {
		exists, err := s.budgets.ExistsPerson(ctx, budgetID, p.UserName)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, apperr.Duplicate("person", "name", p.UserName)
		}
		m, err := s.budgets.AddPerson(ctx, db.AddBudgetPersonParams{
			BudgetID: budgetID,
			UserName: &p.UserName,
			UserID:   p.UserID,
		})
		if err != nil {
			return nil, err
		}
		results = append(results, m)
	}
	return results, nil
}

func (s *BudgetService) ListPeople(ctx context.Context, budgetID, userID uuid.UUID) ([]db.BudgetToUserMapping, error) {
	if _, err := s.Get(ctx, budgetID, userID); err != nil {
		return nil, err
	}
	return s.budgets.ListPeople(ctx, budgetID)
}

func (s *BudgetService) RemovePerson(ctx context.Context, budgetID uuid.UUID, personID int32, replacementPersonID int32, replacementPMID uuid.UUID, userID uuid.UUID) error {
	budget, err := s.Get(ctx, budgetID, userID)
	if err != nil {
		return err
	}
	person, err := s.budgets.GetPerson(ctx, personID, budgetID)
	if err != nil {
		return err
	}
	// Protect the budget owner from removal.
	if person.UserID != nil && *person.UserID == budget.UserID {
		return apperr.Invalid("budget owner cannot be removed")
	}
	// No replacement provided — simple soft-delete (caller guarantees no data needs reassigning).
	if replacementPersonID == 0 {
		return s.budgets.SoftRemovePerson(ctx, db.SoftRemovePersonParams{
			PersonID: personID,
			BudgetID: budgetID,
		})
	}
	// Replacement provided — validate it exists in the same budget then reassign atomically.
	if _, err := s.budgets.GetPerson(ctx, replacementPersonID, budgetID); err != nil {
		return apperr.NotFound("replacement_person", fmt.Sprintf("%d", replacementPersonID))
	}
	repID := replacementPersonID
	return s.budgets.SoftRemovePersonAndReassign(ctx, db.SoftRemovePersonAndReassignParams{
		PersonID:            personID,
		BudgetID:            budgetID,
		ReplacementPmID:     replacementPMID,
		ReplacementPersonID: &repID,
	})
}

// ── Income ────────────────────────────────────────────────────────────────────

type IncomeInput struct {
	Name           string
	Amount         pgtype.Numeric
	Recurring      bool
	BudgetPersonID *int32
}

func (s *BudgetService) AddIncome(ctx context.Context, budgetID, ownerID uuid.UUID, entries []IncomeInput) ([]db.IncomeToBudgetMapping, error) {
	if _, err := s.Get(ctx, budgetID, ownerID); err != nil {
		return nil, err
	}
	var results []db.IncomeToBudgetMapping
	for _, e := range entries {
		m, err := s.budgets.AddIncome(ctx, db.AddIncomeEntryParams{
			BudgetID:       budgetID,
			Name:           &e.Name,
			Amount:         e.Amount,
			Recurring:      e.Recurring,
			BudgetPersonID: e.BudgetPersonID,
		})
		if err != nil {
			return nil, err
		}
		results = append(results, m)
	}
	return results, nil
}

func (s *BudgetService) ListIncome(ctx context.Context, budgetID, userID uuid.UUID) ([]db.IncomeToBudgetMapping, error) {
	if _, err := s.Get(ctx, budgetID, userID); err != nil {
		return nil, err
	}
	return s.budgets.ListIncome(ctx, budgetID)
}

func (s *BudgetService) UpdateIncome(ctx context.Context, incomeID int32, budgetID, userID uuid.UUID, name string, amount pgtype.Numeric, recurring bool, budgetPersonID *int32) (db.IncomeToBudgetMapping, error) {
	if _, err := s.Get(ctx, budgetID, userID); err != nil {
		return db.IncomeToBudgetMapping{}, err
	}
	return s.budgets.UpdateIncome(ctx, db.UpdateIncomeEntryParams{
		ID:             incomeID,
		BudgetID:       budgetID,
		Name:           &name,
		Amount:         amount,
		Recurring:      recurring,
		BudgetPersonID: budgetPersonID,
	})
}

func (s *BudgetService) DeleteIncome(ctx context.Context, incomeID int32, budgetID, userID uuid.UUID) error {
	if _, err := s.Get(ctx, budgetID, userID); err != nil {
		return err
	}
	return s.budgets.DeleteIncome(ctx, db.DeleteIncomeEntryParams{
		ID:       incomeID,
		BudgetID: budgetID,
	})
}
