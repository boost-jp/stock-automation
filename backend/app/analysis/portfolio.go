package analysis

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/boost-jp/stock-automation/app/models"
)

// PortfolioSummary represents portfolio performance summary
type PortfolioSummary struct {
	TotalValue       float64
	TotalCost        float64
	TotalGain        float64
	TotalGainPercent float64
	Holdings         []HoldingSummary
	UpdatedAt        time.Time
}

// HoldingSummary represents individual holding performance
type HoldingSummary struct {
	Code          string
	Name          string
	Shares        int
	CurrentPrice  float64
	PurchasePrice float64
	CurrentValue  float64
	PurchaseCost  float64
	Gain          float64
	GainPercent   float64
	LastUpdated   time.Time
}

// CalculatePortfolioSummary calculates portfolio performance
func CalculatePortfolioSummary(portfolio []models.Portfolio, currentPrices map[string]float64) *PortfolioSummary {
	summary := &PortfolioSummary{
		Holdings:  make([]HoldingSummary, 0),
		UpdatedAt: time.Now(),
	}

	for _, holding := range portfolio {
		currentPrice, exists := currentPrices[holding.Code]
		if !exists {
			continue // Skip if no current price available
		}

		currentValue := float64(holding.Shares) * currentPrice
		purchaseCost := float64(holding.Shares) * holding.PurchasePrice
		gain := currentValue - purchaseCost
		gainPercent := (gain / purchaseCost) * 100

		holdingSummary := HoldingSummary{
			Code:          holding.Code,
			Name:          holding.Name,
			Shares:        holding.Shares,
			CurrentPrice:  currentPrice,
			PurchasePrice: holding.PurchasePrice,
			CurrentValue:  currentValue,
			PurchaseCost:  purchaseCost,
			Gain:          gain,
			GainPercent:   gainPercent,
			LastUpdated:   time.Now(),
		}

		summary.Holdings = append(summary.Holdings, holdingSummary)
		summary.TotalValue += currentValue
		summary.TotalCost += purchaseCost
	}

	summary.TotalGain = summary.TotalValue - summary.TotalCost
	if summary.TotalCost > 0 {
		summary.TotalGainPercent = (summary.TotalGain / summary.TotalCost) * 100
	}

	return summary
}

// GeneratePortfolioReport generates a formatted report
func GeneratePortfolioReport(summary *PortfolioSummary) string {
	if len(summary.Holdings) == 0 {
		return "ポートフォリオにデータがありません"
	}

	report := "📊 ポートフォリオレポート\n\n"

	// 総資産状況
	report += "💰 総資産状況\n"
	report += "━━━━━━━━━━━━━━━━━━━━\n"
	report += sprintf("現在価値: ¥%.0f\n", summary.TotalValue)
	report += sprintf("投資元本: ¥%.0f\n", summary.TotalCost)

	gainIcon := "📈"
	if summary.TotalGain < 0 {
		gainIcon = "📉"
	}

	report += sprintf("損益: %s ¥%.0f (%.2f%%)\n\n", gainIcon, summary.TotalGain, summary.TotalGainPercent)

	// 個別銘柄
	report += "📋 個別銘柄\n"
	report += "━━━━━━━━━━━━━━━━━━━━\n"

	for _, holding := range summary.Holdings {
		icon := "📈"
		if holding.Gain < 0 {
			icon = "📉"
		}

		report += sprintf("%s %s (%s)\n", icon, holding.Name, holding.Code)
		report += sprintf("  保有数: %d株 @ ¥%.0f\n", holding.Shares, holding.PurchasePrice)
		report += sprintf("  現在価格: ¥%.0f\n", holding.CurrentPrice)
		report += sprintf("  損益: ¥%.0f (%.2f%%)\n\n", holding.Gain, holding.GainPercent)
	}

	return report
}

// Helper function for string formatting - Japanese report formatting
func sprintf(format string, args ...interface{}) string {
	switch format {
	case "現在価値: ¥%.0f\n":
		return "現在価値: ¥" + formatCurrency(args[0].(float64)) + "\n"
	case "投資元本: ¥%.0f\n":
		return "投資元本: ¥" + formatCurrency(args[0].(float64)) + "\n"
	case "損益: %s ¥%.0f (%.2f%%)\n\n":
		return "損益: " + args[0].(string) + " ¥" + formatCurrency(args[1].(float64)) + " (" + formatPercent(args[2].(float64)) + "%)\n\n"
	case "%s %s (%s)\n":
		return args[0].(string) + " " + args[1].(string) + " (" + args[2].(string) + ")\n"
	case "  保有数: %d株 @ ¥%.0f\n":
		return "  保有数: " + formatInt(args[0].(int)) + "株 @ ¥" + formatCurrency(args[1].(float64)) + "\n"
	case "  現在価格: ¥%.0f\n":
		return "  現在価格: ¥" + formatCurrency(args[0].(float64)) + "\n"
	case "  損益: ¥%.0f (%.2f%%)\n\n":
		return "  損益: ¥" + formatCurrency(args[0].(float64)) + " (" + formatPercent(args[1].(float64)) + "%)\n\n"
	default:
		// Fallback to standard fmt.Sprintf for unknown formats
		return fmt.Sprintf(format, args...)
	}
}

// formatCurrency formats a float64 as Japanese currency with comma separators
func formatCurrency(f float64) string {
	// Convert to string without decimals
	str := fmt.Sprintf("%.0f", f)

	// Handle negative numbers
	isNegative := false
	if strings.HasPrefix(str, "-") {
		isNegative = true
		str = str[1:] // Remove the negative sign
	}

	// Add comma separators
	formatted := addCommaToNumber(str)

	// Add back negative sign if needed
	if isNegative {
		formatted = "-" + formatted
	}

	return formatted
}

// formatPercent formats a float64 as percentage with 2 decimal places
func formatPercent(f float64) string {
	return fmt.Sprintf("%.2f", f)
}

// formatInt formats an integer as string with comma separators
func formatInt(i int) string {
	return addCommaToNumber(strconv.Itoa(i))
}

// addCommaToNumber adds comma separators to a number string
func addCommaToNumber(s string) string {
	n := len(s)
	if n <= 3 {
		return s
	}

	var result strings.Builder
	for i, digit := range s {
		if i > 0 && (n-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(digit)
	}

	return result.String()
}
