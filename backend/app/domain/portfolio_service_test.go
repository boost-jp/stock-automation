package domain

import (
	"testing"
	"time"

	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/boost-jp/stock-automation/app/domain/models"
	"github.com/google/go-cmp/cmp"
)

func TestNewPortfolioService(t *testing.T) {
	service := NewPortfolioService()
	if service == nil {
		t.Error("Expected non-nil service")
	}
}

func TestPortfolioService_CalculatePortfolioSummary(t *testing.T) {
	service := NewPortfolioService()

	tests := []struct {
		name                string
		portfolios          []*models.Portfolio
		currentPrices       map[string]float64
		expectedValue       float64
		expectedCost        float64
		expectedGain        float64
		expectedGainPercent float64
	}{
		{
			name: "Single holding with profit",
			portfolios: []*models.Portfolio{
				{
					Code:          "1234",
					Name:          "Test Stock",
					Shares:        100,
					PurchasePrice: createDecimal(1000.0),
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
			portfolios: []*models.Portfolio{
				{
					Code:          "5678",
					Name:          "Test Stock 2",
					Shares:        50,
					PurchasePrice: createDecimal(2000.0),
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
			portfolios: []*models.Portfolio{
				{
					Code:          "1234",
					Name:          "Test Stock 1",
					Shares:        100,
					PurchasePrice: createDecimal(1000.0),
					PurchaseDate:  time.Now(),
				},
				{
					Code:          "5678",
					Name:          "Test Stock 2",
					Shares:        50,
					PurchasePrice: createDecimal(2000.0),
					PurchaseDate:  time.Now(),
				},
			},
			currentPrices: map[string]float64{
				"1234": 1200.0,
				"5678": 1800.0,
			},
			expectedValue:       210000.0,
			expectedCost:        200000.0,
			expectedGain:        10000.0,
			expectedGainPercent: 5.0,
		},
		{
			name:                "Empty portfolio",
			portfolios:          []*models.Portfolio{},
			currentPrices:       map[string]float64{},
			expectedValue:       0.0,
			expectedCost:        0.0,
			expectedGain:        0.0,
			expectedGainPercent: 0.0,
		},
		{
			name: "Portfolio with missing price data",
			portfolios: []*models.Portfolio{
				{
					Code:          "1234",
					Name:          "Test Stock",
					Shares:        100,
					PurchasePrice: createDecimal(1000.0),
					PurchaseDate:  time.Now(),
				},
			},
			currentPrices: map[string]float64{
				"5678": 1100.0,
			},
			expectedValue:       0.0,
			expectedCost:        0.0,
			expectedGain:        0.0,
			expectedGainPercent: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := service.CalculatePortfolioSummary(tt.portfolios, tt.currentPrices)

			if diff := cmp.Diff(tt.expectedValue, summary.TotalValue); diff != "" {
				t.Errorf("Total value mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expectedCost, summary.TotalCost); diff != "" {
				t.Errorf("Total cost mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expectedGain, summary.TotalGain); diff != "" {
				t.Errorf("Total gain mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expectedGainPercent, summary.TotalGainPercent); diff != "" {
				t.Errorf("Total gain percent mismatch (-want +got):\n%s", diff)
			}
			if summary.UpdatedAt.IsZero() {
				t.Error("UpdatedAt should be set")
			}
		})
	}
}

func TestPortfolioService_CalculatePortfolioSummary_HoldingDetails(t *testing.T) {
	service := NewPortfolioService()

	portfolios := []*models.Portfolio{
		{
			Code:          "1234",
			Name:          "Test Stock",
			Shares:        100,
			PurchasePrice: createDecimal(1000.0),
			PurchaseDate:  time.Now(),
		},
	}
	currentPrices := map[string]float64{
		"1234": 1100.0,
	}

	summary := service.CalculatePortfolioSummary(portfolios, currentPrices)

	if len(summary.Holdings) != 1 {
		t.Fatalf("Expected 1 holding, got %d", len(summary.Holdings))
	}

	holding := summary.Holdings[0]
	expected := HoldingSummary{
		Code:          "1234",
		Name:          "Test Stock",
		Shares:        100,
		CurrentPrice:  1100.0,
		PurchasePrice: 1000.0,
		CurrentValue:  110000.0,
		PurchaseCost:  100000.0,
		Gain:          10000.0,
		GainPercent:   10.0,
		LastUpdated:   holding.LastUpdated, // Use actual value for comparison
	}

	if diff := cmp.Diff(expected, holding); diff != "" {
		t.Errorf("Holding details mismatch (-want +got):\n%s", diff)
	}
	if holding.LastUpdated.IsZero() {
		t.Error("LastUpdated should be set")
	}
}

func TestPortfolioService_GeneratePortfolioReport(t *testing.T) {
	service := NewPortfolioService()

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
				"📈",
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
				"📉",
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
			report := service.GeneratePortfolioReport(tt.summary)

			if len(report) == 0 {
				t.Error("Report should not be empty")
			}

			for _, expectedString := range tt.expectedContains {
				if !containsString(report, expectedString) {
					t.Errorf("Report should contain expected string: %s", expectedString)
				}
			}
		})
	}
}

func TestPortfolioService_ValidatePortfolio(t *testing.T) {
	service := NewPortfolioService()

	tests := []struct {
		name      string
		portfolio *models.Portfolio
		wantError bool
	}{
		{
			name: "Valid portfolio",
			portfolio: &models.Portfolio{
				Code:          "1234",
				Name:          "Test Stock",
				Shares:        100,
				PurchasePrice: createDecimal(1000.0),
				PurchaseDate:  time.Now(),
			},
			wantError: false,
		},
		{
			name: "Empty code",
			portfolio: &models.Portfolio{
				Code:          "",
				Name:          "Test Stock",
				Shares:        100,
				PurchasePrice: createDecimal(1000.0),
				PurchaseDate:  time.Now(),
			},
			wantError: true,
		},
		{
			name: "Empty name",
			portfolio: &models.Portfolio{
				Code:          "1234",
				Name:          "",
				Shares:        100,
				PurchasePrice: createDecimal(1000.0),
				PurchaseDate:  time.Now(),
			},
			wantError: true,
		},
		{
			name: "Zero shares",
			portfolio: &models.Portfolio{
				Code:          "1234",
				Name:          "Test Stock",
				Shares:        0,
				PurchasePrice: createDecimal(1000.0),
				PurchaseDate:  time.Now(),
			},
			wantError: true,
		},
		{
			name: "Negative shares",
			portfolio: &models.Portfolio{
				Code:          "1234",
				Name:          "Test Stock",
				Shares:        -100,
				PurchasePrice: createDecimal(1000.0),
				PurchaseDate:  time.Now(),
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidatePortfolio(tt.portfolio)

			if tt.wantError {
				if err == nil {
					t.Error("Expected validation error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error, got: %v", err)
				}
			}
		})
	}
}

func TestPortfolioService_CalculateHoldingValue(t *testing.T) {
	service := NewPortfolioService()

	portfolio := &models.Portfolio{
		Code:          "1234",
		Name:          "Test Stock",
		Shares:        100,
		PurchasePrice: createDecimal(1000.0),
		PurchaseDate:  time.Now(),
	}

	currentPrice := 1100.0
	expectedValue := 110000.0

	actualValue := service.CalculateHoldingValue(portfolio, currentPrice)

	if diff := cmp.Diff(expectedValue, actualValue); diff != "" {
		t.Errorf("Holding value calculation mismatch (-want +got):\n%s", diff)
	}
}

func TestPortfolioService_CalculateHoldingReturn(t *testing.T) {
	service := NewPortfolioService()

	tests := []struct {
		name           string
		portfolio      *models.Portfolio
		currentPrice   float64
		expectedReturn float64
	}{
		{
			name: "Positive return",
			portfolio: &models.Portfolio{
				Code:          "1234",
				Name:          "Test Stock",
				Shares:        100,
				PurchasePrice: createDecimal(1000.0),
				PurchaseDate:  time.Now(),
			},
			currentPrice:   1100.0,
			expectedReturn: 10.0,
		},
		{
			name: "Negative return",
			portfolio: &models.Portfolio{
				Code:          "5678",
				Name:          "Test Stock 2",
				Shares:        50,
				PurchasePrice: createDecimal(2000.0),
				PurchaseDate:  time.Now(),
			},
			currentPrice:   1800.0,
			expectedReturn: -10.0,
		},
		{
			name: "Zero return",
			portfolio: &models.Portfolio{
				Code:          "9999",
				Name:          "Test Stock 3",
				Shares:        200,
				PurchasePrice: createDecimal(500.0),
				PurchaseDate:  time.Now(),
			},
			currentPrice:   500.0,
			expectedReturn: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualReturn := service.CalculateHoldingReturn(tt.portfolio, tt.currentPrice)

			if diff := cmp.Diff(tt.expectedReturn, actualReturn); diff != "" {
				t.Errorf("Holding return calculation mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPortfolioService_DecimalConversion(t *testing.T) {
	service := NewPortfolioService()

	tests := []struct {
		name     string
		input    any
		expected float64
	}{
		{
			name:     "String number",
			input:    "1234.56",
			expected: 1234.56,
		},
		{
			name:     "Integer",
			input:    1000,
			expected: 1000.0,
		},
		{
			name:     "Float",
			input:    1500.75,
			expected: 1500.75,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.decimalToFloat(tt.input)
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("Decimal conversion mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// Helper function to create a decimal for testing
func createDecimal(_ float64) types.Decimal {
	// This is a simplified approach for testing
	// In a real implementation, you would use proper decimal creation
	return types.Decimal{}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || contains(s, substr))
}

func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
