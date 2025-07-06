package fixture

import (
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/boost-jp/stock-automation/app/infrastructure/client"
)

// Test IDs for consistent test data
const (
	// Portfolio IDs
	PortfolioID1 = "01HG5Y0YP5JQZKMC5E3R4N2K7V"
	PortfolioID2 = "01HG5Y0YP6JQZKMC5E3R4N2K7W"
	PortfolioID3 = "01HG5Y0YP7JQZKMC5E3R4N2K7X"
	PortfolioID4 = "01HG5Y0YP8JQZKMC5E3R4N2K7Y"

	// Watch List IDs
	WatchListID1 = "01HG5Y0YP9JQZKMC5E3R4N2K7Z"
	WatchListID2 = "01HG5Y0YPAJQZKMC5E3R4N2K80"
	WatchListID3 = "01HG5Y0YPBJQZKMC5E3R4N2K81"
	WatchListID4 = "01HG5Y0YPCJQZKMC5E3R4N2K82"

	// Stock Price IDs
	StockPriceID1 = "01HG5Y0YPDJQZKMC5E3R4N2K83"
	StockPriceID2 = "01HG5Y0YPEJQZKMC5E3R4N2K84"
	StockPriceID3 = "01HG5Y0YPFJQZKMC5E3R4N2K85"
	StockPriceID4 = "01HG5Y0YPGJQZKMC5E3R4N2K86"

	// Technical Indicator IDs
	TechnicalIndicatorID1 = "01HG5Y0YPHJQZKMC5E3R4N2K87"
	TechnicalIndicatorID2 = "01HG5Y0YPIJQZKMC5E3R4N2K88"
	TechnicalIndicatorID3 = "01HG5Y0YPJJQZKMC5E3R4N2K89"
	TechnicalIndicatorID4 = "01HG5Y0YPKJQZKMC5E3R4N2K8A"
)

// Common test stock codes and names
const (
	// Stock codes
	ToyotaCode   = "7203"
	SonyCode     = "6758"
	SoftBankCode = "9984"
	NintendoCode = "7974"

	// Stock names
	ToyotaName   = "トヨタ自動車"
	SonyName     = "ソニーグループ"
	SoftBankName = "ソフトバンクグループ"
	NintendoName = "任天堂"
)

// NullTimeFrom creates a null.Time from a time.Time
func NullTimeFrom(t time.Time) null.Time {
	return null.TimeFrom(t)
}

// NullStringFrom creates a null.String from a string
func NullStringFrom(s string) null.String {
	return null.StringFrom(s)
}

// NullBoolFrom creates a null.Bool from a bool
func NullBoolFrom(b bool) null.Bool {
	return null.BoolFrom(b)
}

// NullFloat64From creates a null.Float64 from a float64
func NullFloat64From(f float64) null.Float64 {
	return null.Float64From(f)
}

// NullInt64From creates a null.Int64 from an int64
func NullInt64From(i int64) null.Int64 {
	return null.Int64From(i)
}

// NullDecimalFrom creates a types.NullDecimal from a float64
func NullDecimalFrom(f float64) types.NullDecimal {
	return client.FloatToNullDecimal(f)
}
