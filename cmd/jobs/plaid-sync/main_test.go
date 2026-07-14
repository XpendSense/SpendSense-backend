package main

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	plaidclient "github.com/mauro-afa91/spendsense/internal/plaid"
	sqlcdb "github.com/mauro-afa91/spendsense/internal/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func numeric(t *testing.T, s string) pgtype.Numeric {
	t.Helper()
	var n pgtype.Numeric
	require.NoError(t, n.Scan(s))
	return n
}

func fixedExpense(t *testing.T, id uuid.UUID, name, amount string) sqlcdb.FixedExpense {
	return sqlcdb.FixedExpense{
		ID:            id,
		Name:          name,
		PlannedAmount: numeric(t, amount),
	}
}

func TestScoreBestMatch_AliasHitPicksAmountCompatibleCandidate(t *testing.T) {
	// Two fixed expenses share the same registered alias ("Patreon"), just
	// like two separate pledges from the same merchant. The $5.33 charge
	// must resolve to the $5.33 template, not the (already-consumed) $15 one.
	fe15ID := uuid.New()
	fe5ID := uuid.New()
	fe15 := fixedExpense(t, fe15ID, "Creator Support A", "15.00")
	fe5 := fixedExpense(t, fe5ID, "Creator Support B", "5.33")
	aliases := map[uuid.UUID][]string{
		fe15ID: {"Patreon"},
		fe5ID:  {"Patreon"},
	}

	tx := plaidclient.Transaction{Name: "Patreon", Amount: 5.33}
	score, bestFE, aliasHit, amountOK := scoreBestMatch(tx, nil, nil, []sqlcdb.FixedExpense{fe15, fe5}, aliases)

	require.NotNil(t, bestFE)
	assert.Equal(t, fe5ID, bestFE.ID, "should resolve to the amount-matching template, not just the first alias hit")
	assert.True(t, aliasHit)
	assert.True(t, amountOK)
	assert.Equal(t, 60.0, score, "name(alias)=20 + amount=40, no payment method/category given")
}

func TestScoreBestMatch_AliasHitWithoutAmountMatchDoesNotReachAutoConfirmThreshold(t *testing.T) {
	// Only one Patreon-aliased template exists (planned $15), and a second,
	// unrelated-amount charge with the same name comes in. The alias still
	// identifies the template, but amountOK must be false so the caller
	// doesn't treat this as a confident match against nothing.
	feID := uuid.New()
	fe := fixedExpense(t, feID, "Creator Support", "15.00")
	aliases := map[uuid.UUID][]string{feID: {"Patreon"}}

	tx := plaidclient.Transaction{Name: "Patreon", Amount: 5.33}
	score, bestFE, aliasHit, amountOK := scoreBestMatch(tx, nil, nil, []sqlcdb.FixedExpense{fe}, aliases)

	require.NotNil(t, bestFE)
	assert.True(t, aliasHit, "alias should still identify the template")
	assert.False(t, amountOK)
	assert.Less(t, score, 80.0, "name+pm+category alone (max 60) must never reach the review-queue threshold without amount")
}

func TestScoreBestMatch_FallsBackToWordOverlapWithoutAlias(t *testing.T) {
	feID := uuid.New()
	fe := fixedExpense(t, feID, "Amex Renewal Membership", "150.00")
	tx := plaidclient.Transaction{Name: "RENEWAL MEMBERSHIP FEE", Amount: 150.00}

	score, bestFE, aliasHit, amountOK := scoreBestMatch(tx, nil, nil, []sqlcdb.FixedExpense{fe}, nil)

	require.NotNil(t, bestFE)
	assert.False(t, aliasHit, "no alias registered — this should be a plain word-overlap match")
	assert.True(t, amountOK)
	assert.Equal(t, 60.0, score)
}

func TestScoreBestMatch_NoCandidatesReturnsNil(t *testing.T) {
	tx := plaidclient.Transaction{Name: "Patreon", Amount: 5.33}
	score, bestFE, aliasHit, amountOK := scoreBestMatch(tx, nil, nil, nil, nil)

	assert.Nil(t, bestFE)
	assert.Equal(t, 0.0, score)
	assert.False(t, aliasHit)
	assert.False(t, amountOK)
}

func TestAmountWithinTolerance(t *testing.T) {
	feID := uuid.New()
	fe := fixedExpense(t, feID, "Rent", "1000.00")

	assert.True(t, amountWithinTolerance(1000.00, &fe))
	assert.True(t, amountWithinTolerance(997.50, &fe), "within $3 tolerance")
	assert.False(t, amountWithinTolerance(990.00, &fe), "outside $3 tolerance")
}

func TestNameWordsOverlap(t *testing.T) {
	assert.True(t, nameWordsOverlap("renewal membership fee", "amex renewal"))
	assert.False(t, nameWordsOverlap("patreon", "creator support subscription"), "short/unrelated words should not match")
}
