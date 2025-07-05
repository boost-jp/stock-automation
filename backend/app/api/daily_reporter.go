package api

import (
	"context"

	"github.com/boost-jp/stock-automation/app/analysis"
	"github.com/boost-jp/stock-automation/app/usecase"
)

// DailyReporter is a thin wrapper around PortfolioReportUseCase for backward compatibility.
type DailyReporter struct {
	useCase *usecase.PortfolioReportUseCase
}

// NewDailyReporter creates a new daily reporter.
func NewDailyReporter(useCase *usecase.PortfolioReportUseCase) *DailyReporter {
	return &DailyReporter{
		useCase: useCase,
	}
}

// GenerateAndSendDailyReport generates and sends daily report.
func (dr *DailyReporter) GenerateAndSendDailyReport() error {
	return dr.useCase.GenerateAndSendDailyReport(context.Background())
}

// SendPortfolioAnalysis sends portfolio analysis.
func (dr *DailyReporter) SendPortfolioAnalysis() error {
	return dr.useCase.SendPortfolioAnalysis(context.Background())
}

// GenerateComprehensiveDailyReport generates a comprehensive daily report.
func (dr *DailyReporter) GenerateComprehensiveDailyReport() (string, error) {
	return dr.useCase.GenerateComprehensiveDailyReport(context.Background())
}

// SendComprehensiveDailyReport sends comprehensive daily report.
func (dr *DailyReporter) SendComprehensiveDailyReport() error {
	return dr.useCase.SendComprehensiveDailyReport(context.Background())
}

// GetPortfolioStatistics returns portfolio statistics.
func (dr *DailyReporter) GetPortfolioStatistics() (*analysis.PortfolioSummary, error) {
	return dr.useCase.GetPortfolioStatistics(context.Background())
}
