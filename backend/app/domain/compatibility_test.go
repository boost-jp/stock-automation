package domain

import (
	"testing"
	"time"

	"github.com/boost-jp/stock-automation/app/domain/models"
	"github.com/boost-jp/stock-automation/app/infrastructure/client"
	"github.com/google/go-cmp/cmp"
)

// Domain logic unit tests - no database access, pure business logic testing

func TestCalculatePortfolioSummary(t *testing.T) {
	tests := []struct {
		name                string
		portfolio           []*models.Portfolio
		currentPrices       map[string]float64
		expectedValue       float64
		expectedCost        float64
		expectedGain        float64
		expectedGainPercent float64
	}{
		{
			name: "Single holding with profit",
			portfolio: []*models.Portfolio{
				{
					Code:          "1234",
					Name:          "Test Stock",
					Shares:        100,
					PurchasePrice: client.FloatToDecimal(1000.0),
					PurchaseDate:  time.Now(),
				},
			},
			currentPrices: map[string]float64{
				"1234": 1100.0,
			},
			expectedValue:       110000.0,
			expectedCost:        100000.0,
			expectedGain:        10000.0,
			expectedGainPercent: 10.0,
		},
		{
			name: "Single holding with loss",
			portfolio: []*models.Portfolio{
				{
					Code:          "5678",
					Name:          "Test Stock 2",
					Shares:        50,
					PurchasePrice: client.FloatToDecimal(2000.0),
					PurchaseDate:  time.Now(),
				},
			},
			currentPrices: map[string]float64{
				"5678": 1800.0,
			},
			expectedValue:       90000.0,
			expectedCost:        100000.0,
			expectedGain:        -10000.0,
			expectedGainPercent: -10.0,
		},
		{
			name: "Multiple holdings mixed performance",
			portfolio: []*models.Portfolio{
				{
					Code:          "1234",
					Name:          "Test Stock 1",
					Shares:        100,
					PurchasePrice: client.FloatToDecimal(1000.0),
					PurchaseDate:  time.Now(),
				},
				{
					Code:          "5678",
					Name:          "Test Stock 2",
					Shares:        50,
					PurchasePrice: client.FloatToDecimal(2000.0),
					PurchaseDate:  time.Now(),
				},
			},
			currentPrices: map[string]float64{
				"1234": 1200.0, // +20%
				"5678": 1800.0, // -10%
			},
			expectedValue:       210000.0, // (100 * 1200) + (50 * 1800)
			expectedCost:        200000.0, // (100 * 1000) + (50 * 2000)
			expectedGain:        10000.0,  // 210000 - 200000
			expectedGainPercent: 5.0,      // (10000 / 200000) * 100
		},
		{
			name:                "Empty portfolio",
			portfolio:           []*models.Portfolio{},
			currentPrices:       map[string]float64{},
			expectedValue:       0.0,
			expectedCost:        0.0,
			expectedGain:        0.0,
			expectedGainPercent: 0.0,
		},
		{
			name: "Portfolio with missing price data",
			portfolio: []*models.Portfolio{
				{
					Code:          "1234",
					Name:          "Test Stock",
					Shares:        100,
					PurchasePrice: client.FloatToDecimal(1000.0),
					PurchaseDate:  time.Now(),
				},
			},
			currentPrices: map[string]float64{
				"5678": 1100.0, // Different stock code
			},
			expectedValue:       0.0,
			expectedCost:        0.0,
			expectedGain:        0.0,
			expectedGainPercent: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := CalculatePortfolioSummary(tt.portfolio, tt.currentPrices)

			if diff := cmp.Diff(tt.expectedValue, summary.TotalValue); diff != "" {
				t.Errorf("TotalValue mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expectedCost, summary.TotalCost); diff != "" {
				t.Errorf("TotalCost mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expectedGain, summary.TotalGain); diff != "" {
				t.Errorf("TotalGain mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expectedGainPercent, summary.TotalGainPercent); diff != "" {
				t.Errorf("TotalGainPercent mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGeneratePortfolioReport(t *testing.T) {
	tests := []struct {
		name             string
		summary          *PortfolioSummary
		expectedContains []string
	}{
		{
			name: "Report with profit",
			summary: &PortfolioSummary{
				TotalValue:       110000.0,
				TotalCost:        100000.0,
				TotalGain:        10000.0,
				TotalGainPercent: 10.0,
				Holdings: []HoldingSummary{
					{
						Code:          "1234",
						Name:          "Test Stock",
						Shares:        100,
						CurrentPrice:  1100.0,
						PurchasePrice: 1000.0,
						CurrentValue:  110000.0,
						PurchaseCost:  100000.0,
						Gain:          10000.0,
						GainPercent:   10.0,
						LastUpdated:   time.Now(),
					},
				},
				UpdatedAt: time.Now(),
			},
			expectedContains: []string{
				"📊 ポートフォリオレポート",
				"💰 総資産状況",
				"現在価値:",
				"投資元本:",
				"損益:",
				"📋 個別銘柄",
				"Test Stock",
				"1234",
				"保有数:",
				"現在価格:",
				"📈", // Profit icon
			},
		},
		{
			name: "Report with loss",
			summary: &PortfolioSummary{
				TotalValue:       90000.0,
				TotalCost:        100000.0,
				TotalGain:        -10000.0,
				TotalGainPercent: -10.0,
				Holdings: []HoldingSummary{
					{
						Code:          "5678",
						Name:          "Test Stock 2",
						Shares:        50,
						CurrentPrice:  1800.0,
						PurchasePrice: 2000.0,
						CurrentValue:  90000.0,
						PurchaseCost:  100000.0,
						Gain:          -10000.0,
						GainPercent:   -10.0,
						LastUpdated:   time.Now(),
					},
				},
				UpdatedAt: time.Now(),
			},
			expectedContains: []string{
				"📊 ポートフォリオレポート",
				"💰 総資産状況",
				"📋 個別銘柄",
				"Test Stock 2",
				"5678",
				"📉", // Loss icon
			},
		},
		{
			name: "Empty portfolio report",
			summary: &PortfolioSummary{
				TotalValue:       0.0,
				TotalCost:        0.0,
				TotalGain:        0.0,
				TotalGainPercent: 0.0,
				Holdings:         []HoldingSummary{},
				UpdatedAt:        time.Now(),
			},
			expectedContains: []string{
				"ポートフォリオにデータがありません",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := GeneratePortfolioReport(tt.summary)

			for _, expectedString := range tt.expectedContains {
				if !contains(report, expectedString) {
					t.Errorf("Report should contain expected string: %s\nActual report:\n%s", expectedString, report)
				}
			}
		})
	}
}

func TestPortfolioCalculations(t *testing.T) {
	t.Run("Gain calculation", func(t *testing.T) {
		// Test gain calculation for individual holding
		holding := HoldingSummary{
			Shares:        100,
			CurrentPrice:  1100.0,
			PurchasePrice: 1000.0,
		}

		expectedCurrentValue := float64(holding.Shares) * holding.CurrentPrice
		expectedPurchaseCost := float64(holding.Shares) * holding.PurchasePrice
		expectedGain := expectedCurrentValue - expectedPurchaseCost
		expectedGainPercent := (expectedGain / expectedPurchaseCost) * 100

		if expectedCurrentValue != 110000.0 {
			t.Errorf("Expected current value 110000, got %f", expectedCurrentValue)
		}
		if expectedPurchaseCost != 100000.0 {
			t.Errorf("Expected purchase cost 100000, got %f", expectedPurchaseCost)
		}
		if expectedGain != 10000.0 {
			t.Errorf("Expected gain 10000, got %f", expectedGain)
		}
		if expectedGainPercent != 10.0 {
			t.Errorf("Expected gain percent 10.0, got %f", expectedGainPercent)
		}
	})

	t.Run("Portfolio aggregation", func(t *testing.T) {
		holdings := []HoldingSummary{
			{
				CurrentValue: 110000.0,
				PurchaseCost: 100000.0,
				Gain:         10000.0,
			},
			{
				CurrentValue: 90000.0,
				PurchaseCost: 100000.0,
				Gain:         -10000.0,
			},
		}

		var totalValue, totalCost, totalGain float64
		for _, h := range holdings {
			totalValue += h.CurrentValue
			totalCost += h.PurchaseCost
			totalGain += h.Gain
		}

		if totalValue != 200000.0 {
			t.Errorf("Expected total value 200000, got %f", totalValue)
		}
		if totalCost != 200000.0 {
			t.Errorf("Expected total cost 200000, got %f", totalCost)
		}
		if totalGain != 0.0 {
			t.Errorf("Expected total gain 0, got %f", totalGain)
		}
	})
}
