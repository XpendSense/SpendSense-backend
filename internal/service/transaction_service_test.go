package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mauro-afa91/spendsense/internal/apperr"
	db "github.com/mauro-afa91/spendsense/internal/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Mock transaction repo ─────────────────────────────────────────────────────

type mockTransactionRepo struct {
	list                    func(context.Context, db.ListTransactionsParams) ([]db.Transaction, error)
	getByID                 func(context.Context, uuid.UUID) (db.Transaction, error)
	create                  func(context.Context, db.CreateTransactionParams) (db.Transaction, error)
	update                  func(context.Context, db.UpdateTransactionParams) (db.Transaction, error)
	delete                  func(context.Context, db.DeleteTransactionParams) error
	getCategory             func(context.Context, int32) (db.GetCategoryRow, error)
	listCategories          func(context.Context, uuid.UUID) ([]db.ListCategoriesRow, error)
	createCategory          func(context.Context, db.CreateCategoryParams) (db.CreateCategoryRow, error)
	updateCategory          func(context.Context, db.UpdateCategoryParams) (db.UpdateCategoryRow, error)
	deleteCategoryAndReassign func(context.Context, db.DeleteCategoryAndReassignParams) error
	listPaymentMethods               func(context.Context, uuid.UUID) ([]db.ListPaymentMethodsRow, error)
	createPaymentMethod              func(context.Context, db.CreatePaymentMethodParams) (db.PaymentMethod, error)
	updatePaymentMethod              func(context.Context, db.UpdatePaymentMethodParams) (db.PaymentMethod, error)
	getPaymentMethod                 func(context.Context, uuid.UUID) (db.PaymentMethod, error)
	deletePaymentMethodAndReassign   func(context.Context, db.DeletePaymentMethodAndReassignParams) error
}

func (m *mockTransactionRepo) List(ctx context.Context, arg db.ListTransactionsParams) ([]db.Transaction, error) {
	if m.list != nil {
		return m.list(ctx, arg)
	}
	return nil, nil
}

func (m *mockTransactionRepo) GetByID(ctx context.Context, id uuid.UUID) (db.Transaction, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id)
	}
	return db.Transaction{}, apperr.NotFound("transaction", id.String())
}

func (m *mockTransactionRepo) Create(ctx context.Context, arg db.CreateTransactionParams) (db.Transaction, error) {
	if m.create != nil {
		return m.create(ctx, arg)
	}
	return db.Transaction{}, nil
}

func (m *mockTransactionRepo) Update(ctx context.Context, arg db.UpdateTransactionParams) (db.Transaction, error) {
	if m.update != nil {
		return m.update(ctx, arg)
	}
	return db.Transaction{}, nil
}

func (m *mockTransactionRepo) Delete(ctx context.Context, arg db.DeleteTransactionParams) error {
	if m.delete != nil {
		return m.delete(ctx, arg)
	}
	return nil
}

func (m *mockTransactionRepo) GetCategory(ctx context.Context, id int32) (db.GetCategoryRow, error) {
	if m.getCategory != nil {
		return m.getCategory(ctx, id)
	}
	return db.GetCategoryRow{}, apperr.NotFound("category", "0")
}

func (m *mockTransactionRepo) ListCategories(ctx context.Context, userID uuid.UUID) ([]db.ListCategoriesRow, error) {
	if m.listCategories != nil {
		return m.listCategories(ctx, userID)
	}
	return nil, nil
}

func (m *mockTransactionRepo) CreateCategory(ctx context.Context, arg db.CreateCategoryParams) (db.CreateCategoryRow, error) {
	if m.createCategory != nil {
		return m.createCategory(ctx, arg)
	}
	return db.CreateCategoryRow{}, nil
}

func (m *mockTransactionRepo) UpdateCategory(ctx context.Context, arg db.UpdateCategoryParams) (db.UpdateCategoryRow, error) {
	if m.updateCategory != nil {
		return m.updateCategory(ctx, arg)
	}
	return db.UpdateCategoryRow{}, nil
}

func (m *mockTransactionRepo) DeleteCategoryAndReassign(ctx context.Context, arg db.DeleteCategoryAndReassignParams) error {
	if m.deleteCategoryAndReassign != nil {
		return m.deleteCategoryAndReassign(ctx, arg)
	}
	return nil
}

func (m *mockTransactionRepo) ListPaymentMethods(ctx context.Context, userID uuid.UUID) ([]db.ListPaymentMethodsRow, error) {
	if m.listPaymentMethods != nil {
		return m.listPaymentMethods(ctx, userID)
	}
	return nil, nil
}

func (m *mockTransactionRepo) CreatePaymentMethod(ctx context.Context, arg db.CreatePaymentMethodParams) (db.PaymentMethod, error) {
	if m.createPaymentMethod != nil {
		return m.createPaymentMethod(ctx, arg)
	}
	return db.PaymentMethod{}, nil
}

func (m *mockTransactionRepo) UpdatePaymentMethod(ctx context.Context, arg db.UpdatePaymentMethodParams) (db.PaymentMethod, error) {
	if m.updatePaymentMethod != nil {
		return m.updatePaymentMethod(ctx, arg)
	}
	return db.PaymentMethod{}, nil
}

func (m *mockTransactionRepo) GetPaymentMethod(ctx context.Context, id uuid.UUID) (db.PaymentMethod, error) {
	if m.getPaymentMethod != nil {
		return m.getPaymentMethod(ctx, id)
	}
	return db.PaymentMethod{}, nil
}

func (m *mockTransactionRepo) DeletePaymentMethodAndReassign(ctx context.Context, arg db.DeletePaymentMethodAndReassignParams) error {
	if m.deletePaymentMethodAndReassign != nil {
		return m.deletePaymentMethodAndReassign(ctx, arg)
	}
	return nil
}

// ── Mock budget repo ──────────────────────────────────────────────────────────

type mockBudgetRepo struct {
	getByID             func(context.Context, uuid.UUID) (db.Budget, error)
	existsByNameAndUser func(context.Context, string, uuid.UUID) (bool, error)
	create              func(context.Context, db.CreateBudgetParams) (db.Budget, error)
	addPerson           func(context.Context, db.AddBudgetPersonParams) (db.BudgetToUserMapping, error)
	addIncome           func(context.Context, db.AddIncomeEntryParams) (db.IncomeToBudgetMapping, error)
	updateIncome        func(context.Context, db.UpdateIncomeEntryParams) (db.IncomeToBudgetMapping, error)
}

func (m *mockBudgetRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]db.Budget, error) {
	return nil, nil
}

func (m *mockBudgetRepo) GetByID(ctx context.Context, id uuid.UUID) (db.Budget, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id)
	}
	return db.Budget{}, apperr.NotFound("budget", id.String())
}

func (m *mockBudgetRepo) ExistsByNameAndUser(ctx context.Context, name string, userID uuid.UUID) (bool, error) {
	if m.existsByNameAndUser != nil {
		return m.existsByNameAndUser(ctx, name, userID)
	}
	return false, nil
}

func (m *mockBudgetRepo) Create(ctx context.Context, arg db.CreateBudgetParams) (db.Budget, error) {
	if m.create != nil {
		return m.create(ctx, arg)
	}
	return db.Budget{ID: uuid.New(), UserID: arg.UserID, Name: arg.Name}, nil
}

func (m *mockBudgetRepo) Update(ctx context.Context, arg db.UpdateBudgetParams) (db.Budget, error) {
	return db.Budget{}, nil
}

func (m *mockBudgetRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockBudgetRepo) ListPeople(ctx context.Context, budgetID uuid.UUID) ([]db.BudgetToUserMapping, error) {
	return nil, nil
}

func (m *mockBudgetRepo) ExistsPerson(ctx context.Context, budgetID uuid.UUID, userName string) (bool, error) {
	return false, nil
}

func (m *mockBudgetRepo) AddPerson(ctx context.Context, arg db.AddBudgetPersonParams) (db.BudgetToUserMapping, error) {
	if m.addPerson != nil {
		return m.addPerson(ctx, arg)
	}
	return db.BudgetToUserMapping{BudgetID: arg.BudgetID, UserName: arg.UserName, UserID: arg.UserID}, nil
}

func (m *mockBudgetRepo) RemovePerson(ctx context.Context, arg db.RemoveBudgetPersonParams) error {
	return nil
}

func (m *mockBudgetRepo) ListIncome(ctx context.Context, budgetID uuid.UUID) ([]db.IncomeToBudgetMapping, error) {
	return nil, nil
}

func (m *mockBudgetRepo) AddIncome(ctx context.Context, arg db.AddIncomeEntryParams) (db.IncomeToBudgetMapping, error) {
	if m.addIncome != nil {
		return m.addIncome(ctx, arg)
	}
	return db.IncomeToBudgetMapping{
		BudgetID:       arg.BudgetID,
		Name:           arg.Name,
		Amount:         arg.Amount,
		Recurring:      arg.Recurring,
		BudgetPersonID: arg.BudgetPersonID,
	}, nil
}

func (m *mockBudgetRepo) UpdateIncome(ctx context.Context, arg db.UpdateIncomeEntryParams) (db.IncomeToBudgetMapping, error) {
	if m.updateIncome != nil {
		return m.updateIncome(ctx, arg)
	}
	return db.IncomeToBudgetMapping{
		ID:             arg.ID,
		BudgetID:       arg.BudgetID,
		Name:           arg.Name,
		Amount:         arg.Amount,
		Recurring:      arg.Recurring,
		BudgetPersonID: arg.BudgetPersonID,
	}, nil
}

func (m *mockBudgetRepo) DeleteIncome(ctx context.Context, arg db.DeleteIncomeEntryParams) error {
	return nil
}

// ── UpdatePaymentMethod tests ─────────────────────────────────────────────────

func TestUpdatePaymentMethod_Success(t *testing.T) {
	typeID := int32(2) // CREDIT
	methodID := uuid.New()
	userID := uuid.New()
	expected := db.PaymentMethod{
		ID:            methodID,
		Name:          "Chase Visa",
		PaymentTypeID: &typeID,
		UserID:        &userID,
	}

	svc := NewTransactionService(
		&mockTransactionRepo{
			updatePaymentMethod: func(_ context.Context, arg db.UpdatePaymentMethodParams) (db.PaymentMethod, error) {
				assert.Equal(t, methodID, arg.ID)
				assert.Equal(t, userID, arg.UserID)
				assert.Equal(t, "Chase Visa", arg.Name)
				return expected, nil
			},
		},
		&mockBudgetRepo{},
	)

	result, err := svc.UpdatePaymentMethod(context.Background(), db.UpdatePaymentMethodParams{
		ID:     methodID,
		Name:   "Chase Visa",
		UserID: userID,
	})

	require.NoError(t, err)
	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, expected.Name, result.Name)
	assert.Equal(t, expected.PaymentTypeID, result.PaymentTypeID)
}

func TestUpdatePaymentMethod_NotFound_WhenUserDoesNotOwnMethod(t *testing.T) {
	svc := NewTransactionService(
		&mockTransactionRepo{
			updatePaymentMethod: func(_ context.Context, arg db.UpdatePaymentMethodParams) (db.PaymentMethod, error) {
				return db.PaymentMethod{}, apperr.NotFound("payment_method", arg.ID.String())
			},
		},
		&mockBudgetRepo{},
	)

	_, err := svc.UpdatePaymentMethod(context.Background(), db.UpdatePaymentMethodParams{
		ID:     uuid.New(),
		Name:   "Renamed",
		UserID: uuid.New(),
	})

	require.Error(t, err)
	var notFound *apperr.NotFoundError
	require.ErrorAs(t, err, &notFound)
	assert.Equal(t, "payment_method", notFound.Resource)
}

// ── CreateCategory tests ──────────────────────────────────────────────────────

func TestCreateCategory_Success(t *testing.T) {
	userID := uuid.New()

	svc := NewTransactionService(
		&mockTransactionRepo{
			createCategory: func(_ context.Context, arg db.CreateCategoryParams) (db.CreateCategoryRow, error) {
				assert.Equal(t, "Hobbies", arg.Name)
				assert.Equal(t, userID, arg.UserID)
				return db.CreateCategoryRow{ID: 42, Name: "Hobbies", IsSystem: false}, nil
			},
		},
		&mockBudgetRepo{},
	)

	result, err := svc.CreateCategory(context.Background(), db.CreateCategoryParams{
		Name:   "Hobbies",
		UserID: userID,
	})

	require.NoError(t, err)
	assert.Equal(t, int32(42), result.ID)
	assert.Equal(t, "Hobbies", result.Name)
	assert.False(t, result.IsSystem)
}

// ── UpdateCategory tests ──────────────────────────────────────────────────────

func TestUpdateCategory_Success(t *testing.T) {
	userID := uuid.New()

	svc := NewTransactionService(
		&mockTransactionRepo{
			updateCategory: func(_ context.Context, arg db.UpdateCategoryParams) (db.UpdateCategoryRow, error) {
				assert.Equal(t, int32(10), arg.ID)
				assert.Equal(t, "Fun Money", arg.Name)
				assert.Equal(t, userID, arg.UserID)
				return db.UpdateCategoryRow{ID: 10, Name: "Fun Money", IsSystem: false}, nil
			},
		},
		&mockBudgetRepo{},
	)

	result, err := svc.UpdateCategory(context.Background(), db.UpdateCategoryParams{
		ID:     10,
		Name:   "Fun Money",
		UserID: userID,
	})

	require.NoError(t, err)
	assert.Equal(t, "Fun Money", result.Name)
}

func TestUpdateCategory_NotFound(t *testing.T) {
	svc := NewTransactionService(
		&mockTransactionRepo{
			updateCategory: func(_ context.Context, arg db.UpdateCategoryParams) (db.UpdateCategoryRow, error) {
				return db.UpdateCategoryRow{}, apperr.NotFound("category", "99")
			},
		},
		&mockBudgetRepo{},
	)

	_, err := svc.UpdateCategory(context.Background(), db.UpdateCategoryParams{
		ID:     99,
		Name:   "Ghost",
		UserID: uuid.New(),
	})

	require.Error(t, err)
	var notFound *apperr.NotFoundError
	require.ErrorAs(t, err, &notFound)
	assert.Equal(t, "category", notFound.Resource)
}

// ── DeleteCategory tests ──────────────────────────────────────────────────────

func TestDeleteCategory_Success(t *testing.T) {
	userID := uuid.New()
	budgetID := uuid.New()
	catID := int32(5)
	replacementID := int32(1)

	svc := NewTransactionService(
		&mockTransactionRepo{
			getCategory: func(_ context.Context, id int32) (db.GetCategoryRow, error) {
				if id == catID {
					return db.GetCategoryRow{ID: catID, Name: "Old", IsSystem: false, UserID: &userID}, nil
				}
				return db.GetCategoryRow{ID: replacementID, Name: "Entertainment", IsSystem: true}, nil
			},
			deleteCategoryAndReassign: func(_ context.Context, arg db.DeleteCategoryAndReassignParams) error {
				assert.Equal(t, catID, arg.ID)
				assert.Equal(t, userID, arg.UserID)
				assert.Equal(t, &replacementID, arg.ReplacementID)
				assert.Equal(t, budgetID, arg.BudgetID)
				return nil
			},
		},
		&mockBudgetRepo{},
	)

	err := svc.DeleteCategory(context.Background(), catID, replacementID, budgetID, userID)
	require.NoError(t, err)
}

func TestDeleteCategory_Forbidden_WhenSystemCategory(t *testing.T) {
	userID := uuid.New()

	svc := NewTransactionService(
		&mockTransactionRepo{
			getCategory: func(_ context.Context, id int32) (db.GetCategoryRow, error) {
				return db.GetCategoryRow{ID: id, Name: "Entertainment", IsSystem: true}, nil
			},
		},
		&mockBudgetRepo{},
	)

	err := svc.DeleteCategory(context.Background(), 1, 2, uuid.New(), userID)
	require.Error(t, err)
	var forbidden *apperr.ForbiddenError
	require.ErrorAs(t, err, &forbidden)
}

func TestDeleteCategory_Forbidden_WhenNotOwner(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	catID := int32(5)

	svc := NewTransactionService(
		&mockTransactionRepo{
			getCategory: func(_ context.Context, id int32) (db.GetCategoryRow, error) {
				return db.GetCategoryRow{ID: catID, Name: "Old", IsSystem: false, UserID: &otherUserID}, nil
			},
		},
		&mockBudgetRepo{},
	)

	err := svc.DeleteCategory(context.Background(), catID, 1, uuid.New(), userID)
	require.Error(t, err)
	var forbidden *apperr.ForbiddenError
	require.ErrorAs(t, err, &forbidden)
}

func TestDeleteCategory_NotFound_WhenCategoryMissing(t *testing.T) {
	svc := NewTransactionService(
		&mockTransactionRepo{
			getCategory: func(_ context.Context, id int32) (db.GetCategoryRow, error) {
				return db.GetCategoryRow{}, apperr.NotFound("category", "99")
			},
		},
		&mockBudgetRepo{},
	)

	err := svc.DeleteCategory(context.Background(), 99, 1, uuid.New(), uuid.New())
	require.Error(t, err)
	var notFound *apperr.NotFoundError
	require.ErrorAs(t, err, &notFound)
}

// ── DeletePaymentMethod tests ─────────────────────────────────────────────────

func TestDeletePaymentMethod_Success(t *testing.T) {
	userID := uuid.New()
	methodID := uuid.New()
	replacementID := uuid.New()
	budgetID := uuid.New()

	svc := NewTransactionService(
		&mockTransactionRepo{
			getPaymentMethod: func(_ context.Context, id uuid.UUID) (db.PaymentMethod, error) {
				return db.PaymentMethod{ID: id, Name: "method", UserID: &userID}, nil
			},
			deletePaymentMethodAndReassign: func(_ context.Context, arg db.DeletePaymentMethodAndReassignParams) error {
				assert.Equal(t, methodID, arg.ID)
				assert.Equal(t, userID, arg.UserID)
				assert.Equal(t, replacementID, arg.ReplacementID)
				assert.Equal(t, budgetID, arg.BudgetID)
				return nil
			},
		},
		&mockBudgetRepo{},
	)

	err := svc.DeletePaymentMethod(context.Background(), methodID, replacementID, budgetID, userID)
	require.NoError(t, err)
}

func TestDeletePaymentMethod_Forbidden_WhenNotOwner(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	methodID := uuid.New()

	svc := NewTransactionService(
		&mockTransactionRepo{
			getPaymentMethod: func(_ context.Context, id uuid.UUID) (db.PaymentMethod, error) {
				return db.PaymentMethod{ID: id, Name: "method", UserID: &otherUserID}, nil
			},
		},
		&mockBudgetRepo{},
	)

	err := svc.DeletePaymentMethod(context.Background(), methodID, uuid.New(), uuid.New(), userID)
	require.Error(t, err)
	var forbidden *apperr.ForbiddenError
	require.ErrorAs(t, err, &forbidden)
}

func TestDeletePaymentMethod_Forbidden_WhenReplacementNotOwned(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	methodID := uuid.New()
	replacementID := uuid.New()

	svc := NewTransactionService(
		&mockTransactionRepo{
			getPaymentMethod: func(_ context.Context, id uuid.UUID) (db.PaymentMethod, error) {
				if id == methodID {
					return db.PaymentMethod{ID: id, Name: "mine", UserID: &userID}, nil
				}
				return db.PaymentMethod{ID: id, Name: "other", UserID: &otherUserID}, nil
			},
		},
		&mockBudgetRepo{},
	)

	err := svc.DeletePaymentMethod(context.Background(), methodID, replacementID, uuid.New(), userID)
	require.Error(t, err)
	var forbidden *apperr.ForbiddenError
	require.ErrorAs(t, err, &forbidden)
}

func TestDeletePaymentMethod_NotFound_WhenMethodMissing(t *testing.T) {
	svc := NewTransactionService(
		&mockTransactionRepo{
			getPaymentMethod: func(_ context.Context, id uuid.UUID) (db.PaymentMethod, error) {
				return db.PaymentMethod{}, apperr.NotFound("payment_method", id.String())
			},
		},
		&mockBudgetRepo{},
	)

	err := svc.DeletePaymentMethod(context.Background(), uuid.New(), uuid.New(), uuid.New(), uuid.New())
	require.Error(t, err)
	var notFound *apperr.NotFoundError
	require.ErrorAs(t, err, &notFound)
}
