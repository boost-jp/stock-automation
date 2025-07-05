package models

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestPortfolio_CalculateCurrentValue(t *testing.T) {
	portfolio := &Portfolio{
		Code:   "1234",
		Name:   "Test Stock",
		Shares: 100,
	}

	tests := []struct {
		name         string
		currentPrice float64
		expected     float64
	}{
		{
			name:         "Basic calculation",
			currentPrice: 1100.0,
			expected:     110000.0,
		},
		{
			name:         "Zero price",
			currentPrice: 0.0,
			expected:     0.0,
		},
		{
			name:         "High price",
			currentPrice: 5000.0,
			expected:     500000.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := portfolio.CalculateCurrentValue(tt.currentPrice)
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("CalculateCurrentValue mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPortfolio_CalculatePurchaseCost(t *testing.T) {
	portfolio := &Portfolio{
		Code:   "1234",
		Name:   "Test Stock",
		Shares: 100,
		// PurchasePrice will return 1000.0 from GetPurchasePrice()
	}

	expected := 100000.0 // 100 shares * 1000.0 price
	result := portfolio.CalculatePurchaseCost()

	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("CalculatePurchaseCost mismatch (-want +got):\n%s", diff)
	}
}

func TestPortfolio_CalculateGain(t *testing.T) {
	portfolio := &Portfolio{
		Code:   "1234",
		Name:   "Test Stock",
		Shares: 100,
	}

	tests := []struct {
		name         string
		currentPrice float64
		expected     float64
	}{
		{
			name:         "Profit scenario",
			currentPrice: 1100.0,
			expected:     10000.0, // (100 * 1100) - (100 * 1000)
		},
		{
			name:         "Loss scenario",
			currentPrice: 900.0,
			expected:     -10000.0, // (100 * 900) - (100 * 1000)
		},
		{
			name:         "No change",
			currentPrice: 1000.0,
			expected:     0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := portfolio.CalculateGain(tt.currentPrice)
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("CalculateGain mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPortfolio_CalculateGainPercent(t *testing.T) {
	portfolio := &Portfolio{
		Code:   "1234",
		Name:   "Test Stock",
		Shares: 100,
	}

	tests := []struct {
		name         string
		currentPrice float64
		expected     float64
	}{
		{
			name:         "10% profit",
			currentPrice: 1100.0,
			expected:     10.0,
		},
		{
			name:         "10% loss",
			currentPrice: 900.0,
			expected:     -10.0,
		},
		{
			name:         "No change",
			currentPrice: 1000.0,
			expected:     0.0,
		},
		{
			name:         "50% profit",
			currentPrice: 1500.0,
			expected:     50.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := portfolio.CalculateGainPercent(tt.currentPrice)
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("CalculateGainPercent mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPortfolio_Validate(t *testing.T) {
	tests := []struct {
		name      string
		portfolio *Portfolio
		wantError bool
	}{
		{
			name: "Valid portfolio",
			portfolio: &Portfolio{
				Code:         "1234",
				Name:         "Test Stock",
				Shares:       100,
				PurchaseDate: time.Now(),
			},
			wantError: false,
		},
		{
			name: "Empty code",
			portfolio: &Portfolio{
				Code:         "",
				Name:         "Test Stock",
				Shares:       100,
				PurchaseDate: time.Now(),
			},
			wantError: true,
		},
		{
			name: "Empty name",
			portfolio: &Portfolio{
				Code:         "1234",
				Name:         "",
				Shares:       100,
				PurchaseDate: time.Now(),
			},
			wantError: true,
		},
		{
			name: "Zero shares",
			portfolio: &Portfolio{
				Code:         "1234",
				Name:         "Test Stock",
				Shares:       0,
				PurchaseDate: time.Now(),
			},
			wantError: true,
		},
		{
			name: "Negative shares",
			portfolio: &Portfolio{
				Code:         "1234",
				Name:         "Test Stock",
				Shares:       -100,
				PurchaseDate: time.Now(),
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.portfolio.Validate()

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
