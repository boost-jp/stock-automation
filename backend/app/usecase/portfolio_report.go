package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/boost-jp/stock-automation/app/domain"
	"github.com/boost-jp/stock-automation/app/infrastructure/client"
	"github.com/boost-jp/stock-automation/app/infrastructure/notification"
	"github.com/boost-jp/stock-automation/app/infrastructure/repository"
	"github.com/sirupsen/logrus"
)

// PortfolioReportUseCase handles portfolio reporting business logic.
type PortfolioReportUseCase struct {
	stockRepo     repository.StockRepository
	portfolioRepo repository.PortfolioRepository
	stockClient   client.StockDataClient
	notifier      notification.NotificationService
}

// NewPortfolioReportUseCase creates a new portfolio report use case.
func NewPortfolioReportUseCase(
	stockRepo repository.StockRepository,
	portfolioRepo repository.PortfolioRepository,
	stockClient client.StockDataClient,
	notifier notification.NotificationService,
) *PortfolioReportUseCase {
	return &PortfolioReportUseCase{
		stockRepo:     stockRepo,
		portfolioRepo: portfolioRepo,
		stockClient:   stockClient,
		notifier:      notifier,
	}
}

// GenerateAndSendDailyReport generates and sends the daily portfolio report.
func (uc *PortfolioReportUseCase) GenerateAndSendDailyReport(ctx context.Context) error {
	logrus.Info("Generating daily portfolio report...")

	// Get portfolio
	portfolio, err := uc.portfolioRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	if len(portfolio) == 0 {
		logrus.Info("No portfolio data found, skipping daily report")
		return nil
	}

	// Get current prices
	currentPrices := make(map[string]float64)
	for _, holding := range portfolio {
		price, err := uc.stockRepo.GetLatestPrice(ctx, holding.Code)
		if err != nil {
			logrus.Warnf("Failed to get price for %s: %v", holding.Code, err)
			continue
		}
		currentPrices[holding.Code] = client.DecimalToFloat(price.ClosePrice)
	}

	// Calculate portfolio summary
	summary := domain.CalculatePortfolioSummary(portfolio, currentPrices)

	// Send notification
	if err := uc.notifier.SendDailyReport(summary.TotalValue, summary.TotalGain, summary.TotalGainPercent); err != nil {
		return err
	}

	logrus.Infof("Daily report sent: Total Value=¥%.0f, Gain=¥%.0f (%.2f%%)",
		summary.TotalValue, summary.TotalGain, summary.TotalGainPercent)

	return nil
}

// SendPortfolioAnalysis sends detailed portfolio domain.
func (uc *PortfolioReportUseCase) SendPortfolioAnalysis(ctx context.Context) error {
	// Get portfolio
	portfolio, err := uc.portfolioRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	// Get current prices
	currentPrices := make(map[string]float64)
	for _, holding := range portfolio {
		price, err := uc.stockRepo.GetLatestPrice(ctx, holding.Code)
		if err != nil {
			continue
		}
		currentPrices[holding.Code] = client.DecimalToFloat(price.ClosePrice)
	}

	// Generate detailed report
	summary := domain.CalculatePortfolioSummary(portfolio, currentPrices)
	report := domain.GeneratePortfolioReport(summary)

	// Send via notification
	return uc.notifier.SendMessage(report)
}

// GenerateComprehensiveDailyReport generates a comprehensive daily report with error handling.
func (uc *PortfolioReportUseCase) GenerateComprehensiveDailyReport(ctx context.Context) (string, error) {
	logrus.Info("Generating comprehensive daily portfolio report...")

	// Get portfolio
	portfolio, err := uc.portfolioRepo.GetAll(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get portfolio: %w", err)
	}

	if len(portfolio) == 0 {
		return "📊 ポートフォリオレポート\n\n💡 現在ポートフォリオにデータがありません", nil
	}

	// Get current prices with error tracking
	currentPrices := make(map[string]float64)
	var priceErrors []string

	for _, holding := range portfolio {
		price, err := uc.stockRepo.GetLatestPrice(ctx, holding.Code)
		if err != nil {
			errorMsg := fmt.Sprintf("%s (%s): 価格取得エラー", holding.Name, holding.Code)
			priceErrors = append(priceErrors, errorMsg)
			logrus.Warnf("Failed to get price for %s: %v", holding.Code, err)
			continue
		}
		currentPrices[holding.Code] = client.DecimalToFloat(price.ClosePrice)
	}

	// Generate report
	summary := domain.CalculatePortfolioSummary(portfolio, currentPrices)
	report := domain.GeneratePortfolioReport(summary)

	// Add errors if any
	if len(priceErrors) > 0 {
		report += "\n⚠️ 価格取得エラー:\n"
		for _, errorMsg := range priceErrors {
			report += fmt.Sprintf("   - %s\n", errorMsg)
		}
	}

	// Add timestamp
	report += fmt.Sprintf("\n🕐 生成時刻: %s", time.Now().Format("2006-01-02 15:04:05"))

	return report, nil
}

// SendComprehensiveDailyReport sends comprehensive daily report via notification.
func (uc *PortfolioReportUseCase) SendComprehensiveDailyReport(ctx context.Context) error {
	report, err := uc.GenerateComprehensiveDailyReport(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate comprehensive report: %w", err)
	}

	// Send via notification
	if err := uc.notifier.SendMessage(report); err != nil {
		return fmt.Errorf("failed to send comprehensive report: %w", err)
	}

	logrus.Info("Comprehensive daily report sent successfully")
	return nil
}

// GetPortfolioStatistics returns detailed portfolio statistics.
func (uc *PortfolioReportUseCase) GetPortfolioStatistics(ctx context.Context) (*domain.PortfolioSummary, error) {
	// Get portfolio
	portfolio, err := uc.portfolioRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}

	if len(portfolio) == 0 {
		return &domain.PortfolioSummary{}, nil
	}

	// Get current prices
	currentPrices := make(map[string]float64)
	for _, holding := range portfolio {
		price, err := uc.stockRepo.GetLatestPrice(ctx, holding.Code)
		if err != nil {
			logrus.Warnf("Failed to get price for %s: %v", holding.Code, err)
			continue
		}
		currentPrices[holding.Code] = client.DecimalToFloat(price.ClosePrice)
	}

	// Calculate statistics
	summary := domain.CalculatePortfolioSummary(portfolio, currentPrices)
	return summary, nil
}
