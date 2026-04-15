// Package decimal provides a type alias and helper functions for arbitrary-precision
// decimal arithmetic using shopspring/decimal. This should be used for all financial
// calculations in the prediction market to avoid floating-point precision errors.
package decimal

import "github.com/shopspring/decimal"

// Decimal is an alias for shopspring/decimal.Decimal to provide consistent
// import path across the codebase.
type Decimal = decimal.Decimal

var (
	// Zero is 0 represented as a Decimal.
	Zero = decimal.Zero
	// One is 1 represented as a Decimal.
	One = decimal.NewFromInt(1)
	// Two is 2 represented as a Decimal.
	Two = decimal.NewFromInt(2)
)

// NewFromInt creates a new Decimal from an int64.
func NewFromInt(i int64) Decimal {
	return decimal.NewFromInt(i)
}

// NewFromFloat creates a new Decimal from a float64.
func NewFromFloat(f float64) Decimal {
	return decimal.NewFromFloat(f)
}

// FromString creates a new Decimal from a string representation.
func FromString(s string) (Decimal, error) {
	return decimal.NewFromString(s)
}
