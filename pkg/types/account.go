// NeuroScript Version: 0.7.4
// File version: 1
// Purpose: Defines the AccountName type for strong typing of account identifiers.
// filename: pkg/types/account.go
// nlines: 11
// risk_rating: LOW

package types

// AccountName is a typed string for account identifiers. It ensures that
// account names are not accidentally confused with other string types in the system.
type AccountName string
