package plaid

import "strings"

// PlaidPaymentTypeID maps a Plaid account type/subtype to a SpendSense
// payment_type DB id (matches proto PaymentType enum values).
//
// Seeded payment_type ids: Cash=1, Credit=2, Debit=3, Digital Wallet=4,
// Bank Transfer=5, Crypto=6, Investment=7.
func PlaidPaymentTypeID(accountType, accountSubtype string) int32 {
	switch strings.ToLower(accountType) {
	case "credit":
		return 2 // Credit
	case "investment", "brokerage":
		return 7 // Investment
	case "depository":
		if strings.ToLower(accountSubtype) == "savings" {
			return 5 // Bank Transfer — savings accounts transfer funds
		}
		return 3 // Debit — checking and other depository accounts
	default:
		return 5 // Bank Transfer — safest fallback for loans and unknowns
	}
}

// PlaidAccountName returns a human-readable payment method name for a Plaid
// account, appending the masked digits when available.
func PlaidAccountName(name, mask string) string {
	if mask != "" {
		return name + " ···" + mask
	}
	return name
}
