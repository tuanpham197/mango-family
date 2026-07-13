package common

// Error codes — contracts/category-api.md (001) & transaction-api.md (002)
const (
	ErrCodeInternal            = "INTERNAL"
	ErrCodeUnauthenticated     = "UNAUTHENTICATED"
	ErrCodeInvalidCredentials  = "INVALID_CREDENTIALS"
	ErrCodeNoHousehold         = "NO_HOUSEHOLD"
	ErrCodeRecordGone          = "RECORD_GONE"
	ErrCodeConcurrencyConflict = "CONCURRENCY_CONFLICT"
	ErrCodeInvalidRequest      = "INVALID_REQUEST"

	// Category (001)
	ErrCodeTypeImmutable        = "TYPE_IMMUTABLE"
	ErrCodeNestingTooDeep       = "NESTING_TOO_DEEP"
	ErrCodeParentTypeMismatch   = "PARENT_TYPE_MISMATCH"
	ErrCodeNameDuplicateWarning = "NAME_DUPLICATE_WARNING"
	ErrCodeCategoryHasTxns      = "CATEGORY_HAS_TRANSACTIONS"
	ErrCodeReassignTypeMismatch = "REASSIGN_TYPE_MISMATCH"

	// Transaction (001 minimal; 002 mở rộng)
	ErrCodeCategoryRequired     = "CATEGORY_REQUIRED"
	ErrCodeCategoryTypeMismatch = "CATEGORY_TYPE_MISMATCH"
	ErrCodeAmountInvalid        = "AMOUNT_INVALID"
	ErrCodeDescriptionTooLong   = "DESCRIPTION_TOO_LONG"

	// Transaction lifecycle (002)
	ErrCodeAccountRequired          = "ACCOUNT_REQUIRED"
	ErrCodeAccountHouseholdMismatch = "ACCOUNT_HOUSEHOLD_MISMATCH"
	ErrCodeFutureDateNotAllowed     = "FUTURE_DATE_NOT_ALLOWED"
)

// Account types (002 — BR-ACC-001)
const (
	AccountTypeCash    = "CASH"
	AccountTypeBank    = "BANK"
	AccountTypeEWallet = "EWALLET"
	AccountTypeCredit  = "CREDIT"
)

// WebSocket/pubsub event topics
const (
	TopicCategoriesChanged   = "categories_changed"
	TopicTransactionsChanged = "transactions_changed"
	TopicAccountsChanged     = "accounts_changed"
)

// Gin context keys
const (
	CtxKeyUser        = "ctx_current_user"
	CtxKeyHouseholdID = "ctx_household_id"
)

// Transaction/category types
const (
	TypeIncome  = "INCOME"
	TypeExpense = "EXPENSE"
)

const MaxDescriptionLen = 255
