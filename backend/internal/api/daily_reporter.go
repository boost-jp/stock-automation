package api

import (
	"stock-automation/internal/analysis"
	"stock-automation/internal/database"
	"stock-automation/internal/notification"

	"github.com/sirupsen/logrus"
)

type DailyReporter struct {
	db       *database.DB
	notifier *notification.SlackNotifier
}

func NewDailyReporter(db *database.DB, notifier *notification.SlackNotifier) *DailyReporter {
	return &DailyReporter{
		db:       db,
		notifier: notifier,
	}
}

func (dr *DailyReporter) GenerateAndSendDailyReport() error {
	logrus.Info("Generating daily portfolio report...")

	// ポートフォリオ取得
	portfolio, err := dr.db.GetPortfolio()
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
		price, err := dr.db.GetLatestPrice(holding.Code)
		if err != nil {
			logrus.Warnf("Failed to get price for %s: %v", holding.Code, err)
			continue
		}
		currentPrices[holding.Code] = price.Price
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
	// ポートフォリオ取得
	portfolio, err := dr.db.GetPortfolio()
	if err != nil {
		return err
	}

	// 現在価格取得
	currentPrices := make(map[string]float64)
	for _, holding := range portfolio {
		price, err := dr.db.GetLatestPrice(holding.Code)
		if err != nil {
			continue
		}
		currentPrices[holding.Code] = price.Price
	}

	// 詳細レポート生成
	summary := analysis.CalculatePortfolioSummary(portfolio, currentPrices)
	report := analysis.GeneratePortfolioReport(summary)

	// Slack送信
	return dr.notifier.SendMessage(report)
}