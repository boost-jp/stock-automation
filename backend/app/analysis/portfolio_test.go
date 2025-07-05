package analysis

import (
	"testing"
	"time"

	"github.com/boost-jp/stock-automation/app/domain/models"
	"github.com/boost-jp/stock-automation/app/infrastructure/client"

	"github.com/stretchr/testify/assert"
)

func TestCalculatePortfolioSummary(t *testing.T) {
	tests := []struct {
		name                string
		portfolio           []models.Portfolio
		currentPrices       map[string]float64
		expectedValue       float64
		expectedCost        float64
		expectedGain        float64
		expectedGainPercent float64
	}{
		{
			name: "Single holding with profit",
			portfolio: []models.Portfolio{
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
			portfolio: []models.Portfolio{
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
			portfolio: []models.Portfolio{
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
			portfolio:           []models.Portfolio{},
			currentPrices:       map[string]float64{},
			expectedValue:       0.0,
			expectedCost:        0.0,
			expectedGain:        0.0,
			expectedGainPercent: 0.0,
		},
		{
			name: "Portfolio with missing price data",
			portfolio: []models.Portfolio{
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

			assert.Equal(t, tt.expectedValue, summary.TotalValue, "Total value mismatch")
			assert.Equal(t, tt.expectedCost, summary.TotalCost, "Total cost mismatch")
			assert.Equal(t, tt.expectedGain, summary.TotalGain, "Total gain mismatch")
			assert.Equal(t, tt.expectedGainPercent, summary.TotalGainPercent, "Total gain percent mismatch")
			assert.NotZero(t, summary.UpdatedAt, "UpdatedAt should be set")
		})
	}
}

func TestCalculatePortfolioSummary_HoldingDetails(t *testing.T) {
	portfolio := []models.Portfolio{
		{
			Code:          "1234",
			Name:          "Test Stock",
			Shares:        100,
			PurchasePrice: 1000.0,
			PurchaseDate:  time.Now(),
		},
	}
	currentPrices := map[string]float64{
		"1234": 1100.0,
	}

	summary := CalculatePortfolioSummary(portfolio, currentPrices)

	assert.Len(t, summary.Holdings, 1, "Should have one holding")

	holding := summary.Holdings[0]
	assert.Equal(t, "1234", holding.Code, "Code mismatch")
	assert.Equal(t, "Test Stock", holding.Name, "Name mismatch")
	assert.Equal(t, 100, holding.Shares, "Shares mismatch")
	assert.Equal(t, 1100.0, holding.CurrentPrice, "Current price mismatch")
	assert.Equal(t, 1000.0, holding.PurchasePrice, "Purchase price mismatch")
	assert.Equal(t, 110000.0, holding.CurrentValue, "Current value mismatch")
	assert.Equal(t, 100000.0, holding.PurchaseCost, "Purchase cost mismatch")
	assert.Equal(t, 10000.0, holding.Gain, "Gain mismatch")
	assert.Equal(t, 10.0, holding.GainPercent, "Gain percent mismatch")
	assert.NotZero(t, holding.LastUpdated, "LastUpdated should be set")
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

			assert.NotEmpty(t, report, "Report should not be empty")

			for _, expectedString := range tt.expectedContains {
				assert.Contains(t, report, expectedString, "Report should contain expected string: %s", expectedString)
			}
		})
	}
}

func TestHoldingSummaryCalculations(t *testing.T) {
	tests := []struct {
		name                 string
		shares               int
		purchasePrice        float64
		currentPrice         float64
		expectedCurrentValue float64
		expectedPurchaseCost float64
		expectedGain         float64
		expectedGainPercent  float64
	}{
		{
			name:                 "Basic profit calculation",
			shares:               100,
			purchasePrice:        1000.0,
			currentPrice:         1100.0,
			expectedCurrentValue: 110000.0,
			expectedPurchaseCost: 100000.0,
			expectedGain:         10000.0,
			expectedGainPercent:  10.0,
		},
		{
			name:                 "Basic loss calculation",
			shares:               50,
			purchasePrice:        2000.0,
			currentPrice:         1800.0,
			expectedCurrentValue: 90000.0,
			expectedPurchaseCost: 100000.0,
			expectedGain:         -10000.0,
			expectedGainPercent:  -10.0,
		},
		{
			name:                 "No change calculation",
			shares:               200,
			purchasePrice:        500.0,
			currentPrice:         500.0,
			expectedCurrentValue: 100000.0,
			expectedPurchaseCost: 100000.0,
			expectedGain:         0.0,
			expectedGainPercent:  0.0,
		},
		{
			name:                 "Fractional shares",
			shares:               75,
			purchasePrice:        1333.33,
			currentPrice:         1400.0,
			expectedCurrentValue: 105000.0,
			expectedPurchaseCost: 99999.75,
			expectedGain:         5000.25,
			expectedGainPercent:  5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			portfolio := []models.Portfolio{
				{
					Code:          "TEST",
					Name:          "Test Stock",
					Shares:        tt.shares,
					PurchasePrice: tt.purchasePrice,
					PurchaseDate:  time.Now(),
				},
			}
			currentPrices := map[string]float64{
				"TEST": tt.currentPrice,
			}

			summary := CalculatePortfolioSummary(portfolio, currentPrices)

			assert.Len(t, summary.Holdings, 1, "Should have one holding")

			holding := summary.Holdings[0]
			assert.InDelta(t, tt.expectedCurrentValue, holding.CurrentValue, 0.01, "Current value mismatch")
			assert.InDelta(t, tt.expectedPurchaseCost, holding.PurchaseCost, 0.01, "Purchase cost mismatch")
			assert.InDelta(t, tt.expectedGain, holding.Gain, 0.01, "Gain mismatch")
			assert.InDelta(t, tt.expectedGainPercent, holding.GainPercent, 0.01, "Gain percent mismatch")
		})
	}
}

func TestPortfolioSummaryEdgeCases(t *testing.T) {
	t.Run("Zero cost portfolio", func(t *testing.T) {
		portfolio := []models.Portfolio{
			{
				Code:          "FREE",
				Name:          "Free Stock",
				Shares:        100,
				PurchasePrice: 0.0, // Free stock
				PurchaseDate:  time.Now(),
			},
		}
		currentPrices := map[string]float64{
			"FREE": 10.0,
		}

		summary := CalculatePortfolioSummary(portfolio, currentPrices)

		assert.Equal(t, 1000.0, summary.TotalValue, "Total value should be 1000")
		assert.Equal(t, 0.0, summary.TotalCost, "Total cost should be 0")
		assert.Equal(t, 1000.0, summary.TotalGain, "Total gain should be 1000")
		assert.Equal(t, 0.0, summary.TotalGainPercent, "Gain percent should be 0 (division by zero)")
	})

	t.Run("Negative price scenario", func(t *testing.T) {
		portfolio := []models.Portfolio{
			{
				Code:          "NEG",
				Name:          "Negative Stock",
				Shares:        100,
				PurchasePrice: client.FloatToDecimal(1000.0),
				PurchaseDate:  time.Now(),
			},
		}
		currentPrices := map[string]float64{
			"NEG": -10.0, // Unusual but possible in some derivative instruments
		}

		summary := CalculatePortfolioSummary(portfolio, currentPrices)

		assert.Equal(t, -1000.0, summary.TotalValue, "Total value should be -1000")
		assert.Equal(t, 100000.0, summary.TotalCost, "Total cost should be 100000")
		assert.Equal(t, -101000.0, summary.TotalGain, "Total gain should be -101000")
		assert.Equal(t, -101.0, summary.TotalGainPercent, "Gain percent should be -101")
	})
}
