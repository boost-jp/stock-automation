package domain

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/boost-jp/stock-automation/app/domain/models"
)

// PortfolioService handles portfolio business logic.
type PortfolioService struct{}

// NewPortfolioService creates a new portfolio service.
func NewPortfolioService() *PortfolioService {
	return &PortfolioService{}
}

// PortfolioSummary represents portfolio performance summary.
type PortfolioSummary struct {
	TotalValue       float64
	TotalCost        float64
	TotalGain        float64
	TotalGainPercent float64
	Holdings         []HoldingSummary
	UpdatedAt        time.Time
}

// HoldingSummary represents individual holding performance.
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

// CalculatePortfolioSummary calculates portfolio performance using domain model methods.
func (s *PortfolioService) CalculatePortfolioSummary(
	portfolios []*models.Portfolio,
	currentPrices map[string]float64,
) *PortfolioSummary {
	summary := &PortfolioSummary{
		TotalValue: 0,
		TotalCost:  0,
		Holdings:   make([]HoldingSummary, 0),
		UpdatedAt:  time.Now(),
	}

	for _, holding := range portfolios {
		currentPrice, exists := currentPrices[holding.Code]
		if !exists {
			continue // Skip if no current price available
		}

		// Use domain model methods for calculations
		currentValue := holding.CalculateCurrentValue(currentPrice)
		purchaseCost := holding.CalculatePurchaseCost()
		gain := holding.CalculateGain(currentPrice)
		gainPercent := holding.CalculateGainPercent(currentPrice)

		holdingSummary := HoldingSummary{
			Code:          holding.Code,
			Name:          holding.Name,
			Shares:        holding.Shares,
			CurrentPrice:  currentPrice,
			PurchasePrice: holding.GetPurchasePrice(),
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

// GeneratePortfolioReport generates a formatted report.
func (s *PortfolioService) GeneratePortfolioReport(summary *PortfolioSummary) string {
	if len(summary.Holdings) == 0 {
		return "ポートフォリオにデータがありません"
	}

	report := "📊 ポートフォリオレポート\n\n"

	// 総資産状況
	report += "💰 総資産状況\n"
	report += "━━━━━━━━━━━━━━━━━━━━\n"
	report += fmt.Sprintf("現在価値: ¥%s\n", formatCurrency(summary.TotalValue))
	report += fmt.Sprintf("投資元本: ¥%s\n", formatCurrency(summary.TotalCost))

	gainIcon := "📈"
	if summary.TotalGain < 0 {
		gainIcon = "📉"
	}

	report += fmt.Sprintf("損益: %s ¥%s (%.2f%%)\n\n",
		gainIcon,
		formatCurrency(summary.TotalGain),
		summary.TotalGainPercent)

	// 個別銘柄
	report += "📋 個別銘柄\n"
	report += "━━━━━━━━━━━━━━━━━━━━\n"

	for _, holding := range summary.Holdings {
		icon := "📈"
		if holding.Gain < 0 {
			icon = "📉"
		}

		report += fmt.Sprintf("%s %s (%s)\n", icon, holding.Name, holding.Code)
		report += fmt.Sprintf("  保有数: %d株 @ ¥%s\n", holding.Shares, formatCurrency(holding.PurchasePrice))
		report += fmt.Sprintf("  現在価格: ¥%s\n", formatCurrency(holding.CurrentPrice))
		report += fmt.Sprintf("  損益: ¥%s (%.2f%%)\n\n",
			formatCurrency(holding.Gain),
			holding.GainPercent)
	}

	return report
}

// formatCurrency formats a float64 as Japanese currency with comma separators.
func formatCurrency(value float64) string {
	// Round to 0 decimal places
	rounded := math.Round(value)
	str := fmt.Sprintf("%.0f", rounded)

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

// addCommaToNumber adds comma separators to a number string.
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

// ValidatePortfolio validates a portfolio entry using domain model.
func (s *PortfolioService) ValidatePortfolio(portfolio *models.Portfolio) error {
	return portfolio.Validate()
}

// CalculateHoldingValue calculates the current value of a holding using domain model.
func (s *PortfolioService) CalculateHoldingValue(portfolio *models.Portfolio, currentPrice float64) float64 {
	return portfolio.CalculateCurrentValue(currentPrice)
}

// CalculateHoldingReturn calculates the return rate of a holding using domain model.
func (s *PortfolioService) CalculateHoldingReturn(portfolio *models.Portfolio, currentPrice float64) float64 {
	return portfolio.CalculateGainPercent(currentPrice)
}
