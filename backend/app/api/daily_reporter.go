package api

import (
	"context"
	"fmt"
	"time"

	"github.com/boost-jp/stock-automation/app/analysis"
	"github.com/boost-jp/stock-automation/app/infrastructure/client"
	"github.com/boost-jp/stock-automation/app/notification"
	"github.com/boost-jp/stock-automation/internal/repository"
	"github.com/sirupsen/logrus"
)

type DailyReporter struct {
	repositories *repository.Repositories
	stockClient  client.StockDataClient
	notifier     *notification.SlackNotifier
}

func NewDailyReporter(repos *repository.Repositories, notifier *notification.SlackNotifier) *DailyReporter {
	return &DailyReporter{
		repositories: repos,
		stockClient:  client.NewYahooFinanceClient(),
		notifier:     notifier,
	}
}

func (dr *DailyReporter) GenerateAndSendDailyReport() error {
	logrus.Info("Generating daily portfolio report...")

	ctx := context.Background()
	// ポートフォリオ取得
	portfolio, err := dr.repositories.Portfolio.GetAll(ctx)
	if err != nil {
		return err
	}

	if len(portfolio) == 0 {
		logrus.Info("No portfolio data found, skipping daily report")

		return nil
	}

	// 現在価格取得
	currentPrices := make(map[string]float64)

	for _, holding := range portfolio {
		price, err := dr.repositories.Stock.GetLatestPrice(ctx, holding.Code)
		if err != nil {
			logrus.Warnf("Failed to get price for %s: %v", holding.Code, err)

			continue
		}

		currentPrices[holding.Code] = client.DecimalToFloat(price.ClosePrice)
	}

	// ポートフォリオサマリー計算
	summary := analysis.CalculatePortfolioSummary(portfolio, currentPrices)

	// Slack通知送信
	if err := dr.notifier.SendDailyReport(summary.TotalValue, summary.TotalGain, summary.TotalGainPercent); err != nil {
		return err
	}

	logrus.Infof("Daily report sent: Total Value=¥%.0f, Gain=¥%.0f (%.2f%%)",
		summary.TotalValue, summary.TotalGain, summary.TotalGainPercent)

	return nil
}

func (dr *DailyReporter) SendPortfolioAnalysis() error {
	ctx := context.Background()
	// ポートフォリオ取得
	portfolio, err := dr.repositories.Portfolio.GetAll(ctx)
	if err != nil {
		return err
	}

	// 現在価格取得
	currentPrices := make(map[string]float64)

	for _, holding := range portfolio {
		price, err := dr.repositories.Stock.GetLatestPrice(ctx, holding.Code)
		if err != nil {
			continue
		}

		currentPrices[holding.Code] = client.DecimalToFloat(price.ClosePrice)
	}

	// 詳細レポート生成
	summary := analysis.CalculatePortfolioSummary(portfolio, currentPrices)
	report := analysis.GeneratePortfolioReport(summary)

	// Slack送信
	return dr.notifier.SendMessage(report)
}

// GenerateComprehensiveDailyReport generates a comprehensive daily report with enhanced error handling.
func (dr *DailyReporter) GenerateComprehensiveDailyReport() (string, error) {
	logrus.Info("Generating comprehensive daily portfolio report...")

	ctx := context.Background()
	// ポートフォリオ取得
	portfolio, err := dr.repositories.Portfolio.GetAll(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get portfolio: %w", err)
	}

	if len(portfolio) == 0 {
		return "📊 ポートフォリオレポート\n\n💡 現在ポートフォリオにデータがありません", nil
	}

	// 現在価格取得（エラーハンドリング強化）
	currentPrices := make(map[string]float64)

	var priceErrors []string

	for _, holding := range portfolio {
		price, err := dr.repositories.Stock.GetLatestPrice(ctx, holding.Code)
		if err != nil {
			errorMsg := fmt.Sprintf("%s (%s): 価格取得エラー", holding.Name, holding.Code)
			priceErrors = append(priceErrors, errorMsg)

			logrus.Warnf("Failed to get price for %s: %v", holding.Code, err)

			continue
		}

		currentPrices[holding.Code] = client.DecimalToFloat(price.ClosePrice)
	}

	// レポート生成
	summary := analysis.CalculatePortfolioSummary(portfolio, currentPrices)
	report := analysis.GeneratePortfolioReport(summary)

	// エラーがあった場合は警告を追加
	if len(priceErrors) > 0 {
		report += "\n⚠️ 価格取得エラー:\n"
		for _, errorMsg := range priceErrors {
			report += fmt.Sprintf("   - %s\n", errorMsg)
		}
	}

	// タイムスタンプ追加
	report += fmt.Sprintf("\n🕐 生成時刻: %s", time.Now().Format("2006-01-02 15:04:05"))

	return report, nil
}

// SendComprehensiveDailyReport sends comprehensive daily report via notification.
func (dr *DailyReporter) SendComprehensiveDailyReport() error {
	report, err := dr.GenerateComprehensiveDailyReport()
	if err != nil {
		return fmt.Errorf("failed to generate comprehensive report: %w", err)
	}

	// Slack送信
	if err := dr.notifier.SendMessage(report); err != nil {
		return fmt.Errorf("failed to send comprehensive report: %w", err)
	}

	logrus.Info("Comprehensive daily report sent successfully")

	return nil
}

// GetPortfolioStatistics returns detailed portfolio statistics.
func (dr *DailyReporter) GetPortfolioStatistics() (*analysis.PortfolioSummary, error) {
	ctx := context.Background()
	// ポートフォリオ取得
	portfolio, err := dr.repositories.Portfolio.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}

	if len(portfolio) == 0 {
		return &analysis.PortfolioSummary{}, nil
	}

	// 現在価格取得
	currentPrices := make(map[string]float64)

	for _, holding := range portfolio {
		price, err := dr.repositories.Stock.GetLatestPrice(ctx, holding.Code)
		if err != nil {
			logrus.Warnf("Failed to get price for %s: %v", holding.Code, err)

			continue
		}

		currentPrices[holding.Code] = client.DecimalToFloat(price.ClosePrice)
	}

	// 統計計算
	summary := analysis.CalculatePortfolioSummary(portfolio, currentPrices)

	return summary, nil
}
