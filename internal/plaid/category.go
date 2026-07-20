package plaid

// plaidDetailedToCategory maps Plaid personal_finance_category.detailed values
// to SpendSense system category names. Only entries that differ from the
// primary-level default are listed here; everything else falls through to
// plaidPrimaryToCategory.
var plaidDetailedToCategory = map[string]string{
	// FOOD_AND_DRINK — split groceries out of the generic "Eating Out" default
	"FOOD_AND_DRINK_GROCERIES": "Groceries",

	// GENERAL_MERCHANDISE — pet supplies break out of generic Shopping
	"GENERAL_MERCHANDISE_PET_SUPPLIES": "Pet",

	// MEDICAL — vet bills belong with Pet, not Wellness
	"MEDICAL_VETERINARY_SERVICES": "Pet",

	// PERSONAL_CARE — laundry/dry cleaning is a household service, not personal care
	"PERSONAL_CARE_LAUNDRY_AND_DRY_CLEANING": "Services",

	// GENERAL_SERVICES — override the Services default for specific subtypes
	"GENERAL_SERVICES_INSURANCE":  "Insurance",
	"GENERAL_SERVICES_AUTOMOTIVE": "Auto", // oil changes, car washes, repairs, towing
	"GENERAL_SERVICES_CHILDCARE":  "Baby", // babysitters, daycare

	// TRANSPORTATION — only gas and tolls map to Gas; everything else is Misc
	"TRANSPORTATION_GAS":   "Gas",
	"TRANSPORTATION_TOLLS": "Gas",
}

// plaidPrimaryToCategory maps Plaid personal_finance_category.primary values
// to SpendSense system category names. Used as a fallback when the detailed
// key has no specific override.
var plaidPrimaryToCategory = map[string]string{
	"FOOD_AND_DRINK":            "Eating Out",
	"GENERAL_MERCHANDISE":       "Shopping",
	"HOME_IMPROVEMENT":          "House",
	"MEDICAL":                   "Wellness",
	"PERSONAL_CARE":             "Wellness",
	"GENERAL_SERVICES":          "Services",
	"TRANSPORTATION":            "Misc",
	"TRAVEL":                    "Travel",
	"RENT_AND_UTILITIES":        "Rent",
	"ENTERTAINMENT":             "Entertainment",
	"BANK_FEES":                 "Misc",
	"GOVERNMENT_AND_NON_PROFIT": "Misc",
	// INCOME covers Plaid's own wages/dividends/interest/refund/etc.
	// classification — mapping it here means a payroll deposit is recognized
	// as Income from Plaid's PFC data alone, without depending on the
	// transaction name containing the literal word "payroll" (see
	// syncResolveCategory in internal/service/plaid_sync.go for that
	// name-based override, kept as a fallback for accounts where Plaid
	// doesn't return personal_finance_category data at all).
	"INCOME": "Income",
	// TRANSFER_OUT and LOAN_PAYMENTS are intentionally left unmapped — credit
	// card bill payments are filtered separately (see isCreditCardPayment in
	// client.go); other transfers/loan payments fall through to uncategorized
	// rather than being guessed at. TRANSFER_IN is also left unmapped since it
	// covers ordinary transfers between the user's own accounts, not just
	// income, and mislabeling those as Income would incorrectly exclude real
	// inbound transfers from spend totals.
}

// ResolvePlaidCategory returns the SpendSense system category name for a Plaid
// transaction. Detailed is checked first, primary is the fallback. Returns ""
// when no mapping exists (transaction will be imported uncategorized).
func ResolvePlaidCategory(primary, detailed string) string {
	if cat, ok := plaidDetailedToCategory[detailed]; ok {
		return cat
	}
	return plaidPrimaryToCategory[primary]
}
