package notification

// NotificationService defines the interface for notification services.
type NotificationService interface {
	// SendMessage sends a plain text message
	SendMessage(message string) error

	// SendStockAlert sends a stock price alert
	SendStockAlert(stockCode, stockName string, currentPrice, targetPrice float64, alertType string) error

	// SendDailyReport sends a daily portfolio report
	SendDailyReport(totalValue, totalGain float64, gainPercent float64) error
}
