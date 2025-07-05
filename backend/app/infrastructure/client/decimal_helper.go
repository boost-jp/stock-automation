package client

import (
	"fmt"

	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/ericlagergren/decimal"
)

// FloatToDecimal converts float64 to types.Decimal.
func FloatToDecimal(value float64) types.Decimal {
	// Create decimal from string representation to maintain precision
	decimalStr := fmt.Sprintf("%.6f", value)
	d := new(decimal.Big)
	d.SetString(decimalStr)
	return types.Decimal{Big: d}
}

// FloatToDecimalPtr converts float64 to *types.Decimal.
func FloatToDecimalPtr(value float64) *types.Decimal {
	decimal := FloatToDecimal(value)
	return &decimal
}

// DecimalToFloat converts types.Decimal to float64.
func DecimalToFloat(d types.Decimal) float64 {
	if d.Big == nil {
		return 0.0
	}
	f, _ := d.Big.Float64()
	return f
}

// FloatToNullDecimal converts float64 to types.NullDecimal.
func FloatToNullDecimal(value float64) types.NullDecimal {
	d := new(decimal.Big)
	d.SetFloat64(value)
	return types.NullDecimal{Big: d}
}

// NullDecimalToFloat converts types.NullDecimal to float64.
func NullDecimalToFloat(nd types.NullDecimal) float64 {
	if nd.Big == nil {
		return 0.0
	}
	f, _ := nd.Big.Float64()
	return f
}

// floatToDecimal converts float64 to types.Decimal.
func floatToDecimal(value float64) types.Decimal {
	return FloatToDecimal(value)
}

// floatToDecimalPtr converts float64 to *types.Decimal.
func floatToDecimalPtr(value float64) *types.Decimal {
	return FloatToDecimalPtr(value)
}