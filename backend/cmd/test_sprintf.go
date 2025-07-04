package main

import (
	"fmt"
	"stock-automation/internal/analysis"
	"stock-automation/internal/models"
	"time"
)

func main() {
	// Test data
	portfolio := []models.Portfolio{
		{
			ID:            1,
			Code:          "7203",
			Name:          "トヨタ自動車",
			Shares:        100,
			PurchasePrice: 2000.0,
			PurchaseDate:  time.Now().AddDate(0, -1, 0),
		},
		{
			ID:            2,
			Code:          "6758",
			Name:          "ソニーグループ",
			Shares:        50,
			PurchasePrice: 12000.0,
			PurchaseDate:  time.Now().AddDate(0, -2, 0),
		},
	}
	
	// Current prices
	currentPrices := map[string]float64{
		"7203": 2500.0, // +25% gain
		"6758": 11000.0, // -8.33% loss
	}
	
	// Calculate portfolio summary
	summary := analysis.CalculatePortfolioSummary(portfolio, currentPrices)
	
	// Generate report
	report := analysis.GeneratePortfolioReport(summary)
	
	fmt.Println("=== Portfolio Report Test ===")
	fmt.Println(report)
	
	fmt.Println("=== Summary Values ===")
	fmt.Printf("Total Value: ¥%.0f\n", summary.TotalValue)
	fmt.Printf("Total Cost: ¥%.0f\n", summary.TotalCost)
	fmt.Printf("Total Gain: ¥%.0f (%.2f%%)\n", summary.TotalGain, summary.TotalGainPercent)
}