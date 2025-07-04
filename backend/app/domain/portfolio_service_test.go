package domain

import (
	"fmt"
	"testing"
	"time"

	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/boost-jp/stock-automation/app/domain/models"
	"github.com/ericlagergren/decimal"
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
				createTestPortfolio("1234", "Test Stock", 100, 1000.0),
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
				createTestPortfolio("5678", "Test Stock 2", 50, 2000.0),
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
				createTestPortfolio("1234", "Test Stock 1", 100, 1000.0),
				createTestPortfolio("5678", "Test Stock 2", 50, 2000.0),
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
				createTestPortfolio("1234", "Test Stock", 100, 1000.0),
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
		createTestPortfolio("1234", "Test Stock", 100, 1000.0),
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
			name:      "Valid portfolio",
			portfolio: createTestPortfolio("1234", "Test Stock", 100, 1000.0),
			wantError: false,
		},
		{
			name:      "Empty code",
			portfolio: createTestPortfolio("", "Test Stock", 100, 1000.0),
			wantError: true,
		},
		{
			name:      "Empty name",
			portfolio: createTestPortfolio("1234", "", 100, 1000.0),
			wantError: true,
		},
		{
			name:      "Zero shares",
			portfolio: createTestPortfolio("1234", "Test Stock", 0, 1000.0),
			wantError: true,
		},
		{
			name:      "Negative shares",
			portfolio: createTestPortfolio("1234", "Test Stock", -100, 1000.0),
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

	portfolio := createTestPortfolio("1234", "Test Stock", 100, 1000.0)

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
			name:           "Positive return",
			portfolio:      createTestPortfolio("1234", "Test Stock", 100, 1000.0),
			currentPrice:   1100.0,
			expectedReturn: 10.0,
		},
		{
			name:           "Negative return",
			portfolio:      createTestPortfolio("5678", "Test Stock 2", 50, 2000.0),
			currentPrice:   1800.0,
			expectedReturn: -10.0,
		},
		{
			name:           "Zero return",
			portfolio:      createTestPortfolio("9999", "Test Stock 3", 200, 500.0),
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

// TestPortfolio is a test version of Portfolio with float64 prices for easier testing
type TestPortfolio struct {
	Code          string
	Name          string
	Shares        int
	PurchasePrice float64
	PurchaseDate  time.Time
}

// Convert TestPortfolio to actual Portfolio with proper types.Decimal
func (tp TestPortfolio) ToPortfolio() *models.Portfolio {
	return &models.Portfolio{
		Code:          tp.Code,
		Name:          tp.Name,
		Shares:        tp.Shares,
		PurchasePrice: floatToDecimal(tp.PurchasePrice),
		PurchaseDate:  tp.PurchaseDate,
	}
}

// floatToDecimal converts float64 to types.Decimal for testing
func floatToDecimal(value float64) types.Decimal {
	decimalStr := fmt.Sprintf("%.6f", value)
	d := new(decimal.Big)
	d.SetString(decimalStr)
	return types.Decimal{Big: d}
}

// Helper function to create a portfolio with test price
func createTestPortfolio(code, name string, shares int, purchasePrice float64) *models.Portfolio {
	// Convert float to decimal using the infrastructure client helper
	decimalValue := floatToDecimal(purchasePrice)

	return &models.Portfolio{
		Code:          code,
		Name:          name,
		Shares:        shares,
		PurchasePrice: decimalValue,
		PurchaseDate:  time.Now(),
	}
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
