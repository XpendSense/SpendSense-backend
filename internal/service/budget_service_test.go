package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mauro-afa91/spendsense/internal/apperr"
	db "github.com/mauro-afa91/spendsense/internal/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptrI32(i int32) *int32 { return &i }

// ── Tests: Create (auto-owner) ────────────────────────────────────────────────

func TestCreate_AutoAddsOwnerAsBudgetPerson(t *testing.T) {
	userID := uuid.New()
	firstName, lastName := "Alice", "Smith"
	addPersonCalled := false

	budgetRepo := &mockBudgetRepo{
		addPerson: func(_ context.Context, arg db.AddBudgetPersonParams) (db.BudgetToUserMapping, error) {
			addPersonCalled = true
			assert.Equal(t, userID, *arg.UserID)
			assert.Equal(t, "Alice Smith", *arg.UserName)
			return db.BudgetToUserMapping{}, nil
		},
	}
	userRepo := &mockUserRepo{
		getByID: func(_ context.Context, id uuid.UUID) (db.User, error) {
			return db.User{ID: id, Email: "alice@example.com", FirstName: &firstName, LastName: &lastName}, nil
		},
	}

	svc := NewBudgetService(budgetRepo, userRepo)
	budget, err := svc.Create(context.Background(), userID, "My Budget")

	require.NoError(t, err)
	assert.Equal(t, "My Budget", budget.Name)
	assert.True(t, addPersonCalled, "owner should be auto-added as BudgetPerson")
}

func TestCreate_UserLookupFails_BudgetStillReturned(t *testing.T) {
	userID := uuid.New()

	svc := NewBudgetService(&mockBudgetRepo{}, &mockUserRepo{
		getByID: func(_ context.Context, _ uuid.UUID) (db.User, error) {
			return db.User{}, apperr.NotFound("user", userID.String())
		},
	})
	budget, err := svc.Create(context.Background(), userID, "Budget")

	require.NoError(t, err, "budget creation must succeed even when user lookup fails")
	assert.Equal(t, "Budget", budget.Name)
}

func TestCreate_DuplicateName_ReturnsError(t *testing.T) {
	userID := uuid.New()
	budgetRepo := &mockBudgetRepo{
		existsByNameAndUser: func(_ context.Context, _ string, _ uuid.UUID) (bool, error) {
			return true, nil
		},
	}
	svc := NewBudgetService(budgetRepo, &mockUserRepo{})
	_, err := svc.Create(context.Background(), userID, "Duplicate")
	require.Error(t, err)
}

// ── Tests: AddIncome (person attribution) ────────────────────────────────────

func TestAddIncome_WithBudgetPersonID(t *testing.T) {
	userID := uuid.New()
	budgetID := uuid.New()
	personID := int32(42)

	budgetRepo := &mockBudgetRepo{
		getByID: func(_ context.Context, id uuid.UUID) (db.Budget, error) {
			return db.Budget{ID: id, UserID: userID}, nil
		},
		addIncome: func(_ context.Context, arg db.AddIncomeEntryParams) (db.IncomeToBudgetMapping, error) {
			require.NotNil(t, arg.BudgetPersonID)
			assert.Equal(t, personID, *arg.BudgetPersonID)
			return db.IncomeToBudgetMapping{
				BudgetID:       arg.BudgetID,
				Name:           arg.Name,
				Amount:         arg.Amount,
				BudgetPersonID: arg.BudgetPersonID,
			}, nil
		},
	}

	svc := NewBudgetService(budgetRepo, &mockUserRepo{})
	entries, err := svc.AddIncome(context.Background(), budgetID, userID, []IncomeInput{{
		Name:           "Salary",
		Amount:         pgtype.Numeric{},
		BudgetPersonID: &personID,
	}})

	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, personID, *entries[0].BudgetPersonID)
}

func TestAddIncome_Unattributed(t *testing.T) {
	userID := uuid.New()
	budgetID := uuid.New()

	budgetRepo := &mockBudgetRepo{
		getByID: func(_ context.Context, id uuid.UUID) (db.Budget, error) {
			return db.Budget{ID: id, UserID: userID}, nil
		},
	}

	svc := NewBudgetService(budgetRepo, &mockUserRepo{})
	entries, err := svc.AddIncome(context.Background(), budgetID, userID, []IncomeInput{{
		Name:   "Other",
		Amount: pgtype.Numeric{},
	}})

	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Nil(t, entries[0].BudgetPersonID)
}

// ── Tests: UpdateIncome (person attribution) ──────────────────────────────────

func TestUpdateIncome_WithBudgetPersonID(t *testing.T) {
	userID := uuid.New()
	budgetID := uuid.New()
	personID := int32(7)

	budgetRepo := &mockBudgetRepo{
		getByID: func(_ context.Context, id uuid.UUID) (db.Budget, error) {
			return db.Budget{ID: id, UserID: userID}, nil
		},
		updateIncome: func(_ context.Context, arg db.UpdateIncomeEntryParams) (db.IncomeToBudgetMapping, error) {
			require.NotNil(t, arg.BudgetPersonID)
			assert.Equal(t, personID, *arg.BudgetPersonID)
			return db.IncomeToBudgetMapping{BudgetPersonID: arg.BudgetPersonID}, nil
		},
	}

	svc := NewBudgetService(budgetRepo, &mockUserRepo{})
	entry, err := svc.UpdateIncome(context.Background(), 1, budgetID, userID, "Salary", pgtype.Numeric{}, false, &personID)

	require.NoError(t, err)
	assert.Equal(t, personID, *entry.BudgetPersonID)
}

func TestUpdateIncome_ClearAttribution(t *testing.T) {
	userID := uuid.New()
	budgetID := uuid.New()

	budgetRepo := &mockBudgetRepo{
		getByID: func(_ context.Context, id uuid.UUID) (db.Budget, error) {
			return db.Budget{ID: id, UserID: userID}, nil
		},
		updateIncome: func(_ context.Context, arg db.UpdateIncomeEntryParams) (db.IncomeToBudgetMapping, error) {
			assert.Nil(t, arg.BudgetPersonID)
			return db.IncomeToBudgetMapping{}, nil
		},
	}

	svc := NewBudgetService(budgetRepo, &mockUserRepo{})
	_, err := svc.UpdateIncome(context.Background(), 1, budgetID, userID, "Salary", pgtype.Numeric{}, false, nil)
	require.NoError(t, err)
}

// ── Tests: RemovePerson ───────────────────────────────────────────────────────

func TestRemovePerson_Success(t *testing.T) {
	userID := uuid.New()
	budgetID := uuid.New()
	ownerUID := uuid.New()
	personID := int32(5)
	replacementID := int32(6)
	replacementPMID := uuid.New()
	removeCalled := false

	svc := NewBudgetService(&mockBudgetRepo{
		getByID: func(_ context.Context, id uuid.UUID) (db.Budget, error) {
			return db.Budget{ID: id, UserID: userID}, nil
		},
		getPerson: func(_ context.Context, pid int32, _ uuid.UUID) (db.BudgetToUserMapping, error) {
			switch pid {
			case personID:
				return db.BudgetToUserMapping{ID: pid, UserID: &ownerUID, IsActive: true}, nil
			case replacementID:
				return db.BudgetToUserMapping{ID: pid, IsActive: true}, nil
			}
			return db.BudgetToUserMapping{}, apperr.NotFound("person", "")
		},
		softRemovePersonAndReassign: func(_ context.Context, arg db.SoftRemovePersonAndReassignParams) error {
			removeCalled = true
			assert.Equal(t, personID, arg.PersonID)
			assert.Equal(t, replacementPMID, arg.ReplacementPmID)
			return nil
		},
	}, &mockUserRepo{})

	err := svc.RemovePerson(context.Background(), budgetID, personID, replacementID, replacementPMID, userID)
	require.NoError(t, err)
	assert.True(t, removeCalled)
}

func TestRemovePerson_OwnerBlocked(t *testing.T) {
	userID := uuid.New()
	budgetID := uuid.New()
	personID := int32(1)

	svc := NewBudgetService(&mockBudgetRepo{
		getByID: func(_ context.Context, id uuid.UUID) (db.Budget, error) {
			return db.Budget{ID: id, UserID: userID}, nil
		},
		getPerson: func(_ context.Context, _ int32, _ uuid.UUID) (db.BudgetToUserMapping, error) {
			return db.BudgetToUserMapping{ID: personID, UserID: &userID, IsActive: true}, nil
		},
	}, &mockUserRepo{})

	err := svc.RemovePerson(context.Background(), budgetID, personID, 2, uuid.New(), userID)
	require.Error(t, err)
	var valErr *apperr.ValidationError
	require.ErrorAs(t, err, &valErr)
}

func TestRemovePerson_PersonNotFound(t *testing.T) {
	userID := uuid.New()
	budgetID := uuid.New()

	svc := NewBudgetService(&mockBudgetRepo{
		getByID: func(_ context.Context, id uuid.UUID) (db.Budget, error) {
			return db.Budget{ID: id, UserID: userID}, nil
		},
		getPerson: func(_ context.Context, _ int32, _ uuid.UUID) (db.BudgetToUserMapping, error) {
			return db.BudgetToUserMapping{}, apperr.NotFound("person", "99")
		},
	}, &mockUserRepo{})

	err := svc.RemovePerson(context.Background(), budgetID, 99, 2, uuid.New(), userID)
	require.Error(t, err)
	var nfErr *apperr.NotFoundError
	require.ErrorAs(t, err, &nfErr)
}

func TestRemovePerson_ReplacementNotFound(t *testing.T) {
	userID := uuid.New()
	budgetID := uuid.New()
	otherUID := uuid.New()
	personID := int32(5)
	replacementID := int32(99)

	svc := NewBudgetService(&mockBudgetRepo{
		getByID: func(_ context.Context, id uuid.UUID) (db.Budget, error) {
			return db.Budget{ID: id, UserID: userID}, nil
		},
		getPerson: func(_ context.Context, pid int32, _ uuid.UUID) (db.BudgetToUserMapping, error) {
			if pid == personID {
				return db.BudgetToUserMapping{ID: pid, UserID: &otherUID, IsActive: true}, nil
			}
			return db.BudgetToUserMapping{}, apperr.NotFound("person", "99")
		},
	}, &mockUserRepo{})

	err := svc.RemovePerson(context.Background(), budgetID, personID, replacementID, uuid.New(), userID)
	require.Error(t, err)
	var nfErr *apperr.NotFoundError
	require.ErrorAs(t, err, &nfErr)
}
