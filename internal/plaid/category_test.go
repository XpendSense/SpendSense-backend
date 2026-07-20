package plaid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolvePlaidCategory_IncomePrimaryMapsToIncome(t *testing.T) {
	assert.Equal(t, "Income", ResolvePlaidCategory("INCOME", "INCOME_WAGES"))
	assert.Equal(t, "Income", ResolvePlaidCategory("INCOME", "INCOME_DIVIDENDS"))
	assert.Equal(t, "Income", ResolvePlaidCategory("INCOME", "INCOME_OTHER_INCOME"))
}

func TestResolvePlaidCategory_TransferInIsIntentionallyUnmapped(t *testing.T) {
	// TRANSFER_IN covers ordinary transfers between the user's own accounts,
	// not just income — mapping it to Income would wrongly exclude real
	// inbound transfers from spend totals, so it stays uncategorized.
	assert.Equal(t, "", ResolvePlaidCategory("TRANSFER_IN", "TRANSFER_IN_ACCOUNT_TRANSFER"))
}

func TestResolvePlaidCategory_DetailedOverridesPrimary(t *testing.T) {
	assert.Equal(t, "Groceries", ResolvePlaidCategory("FOOD_AND_DRINK", "FOOD_AND_DRINK_GROCERIES"))
}

func TestResolvePlaidCategory_UnknownReturnsEmpty(t *testing.T) {
	assert.Equal(t, "", ResolvePlaidCategory("SOMETHING_NEW", "SOMETHING_NEW_SUBTYPE"))
}
