package analysis

import (
	"math"

	"github.com/boost-jp/stock-automation/app/models"
)

// MovingAverage calculates simple moving average
func MovingAverage(prices []models.StockPrice, period int) float64 {
	if len(prices) < period {
		return 0
	}

	sum := 0.0
	for i := len(prices) - period; i < len(prices); i++ {
		sum += prices[i].Close
	}

	return sum / float64(period)
}

// RSI calculates Relative Strength Index
func RSI(prices []models.StockPrice, period int) float64 {
	if len(prices) <= period {
		return 50.0 // 中立値
	}

	gains := 0.0
	losses := 0.0

	// 最初のperiod分を計算
	for i := 1; i <= period; i++ {
		change := prices[len(prices)-i].Close - prices[len(prices)-i-1].Close
		if change > 0 {
			gains += change
		} else {
			losses += math.Abs(change)
		}
	}

	if losses == 0 {
		return 100.0
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))

	return rsi
}

// MACD calculates Moving Average Convergence Divergence
func MACD(prices []models.StockPrice, fastPeriod, slowPeriod, signalPeriod int) (macd, signal, histogram float64) {
	if len(prices) < slowPeriod {
		return 0, 0, 0
	}

	// EMA計算のヘルパー関数
	calculateEMA := func(data []models.StockPrice, period int) float64 {
		if len(data) < period {
			return 0
		}

		multiplier := 2.0 / (float64(period) + 1.0)
		ema := data[0].Close

		for i := 1; i < len(data); i++ {
			ema = (data[i].Close * multiplier) + (ema * (1 - multiplier))
		}

		return ema
	}

	// Fast EMAとSlow EMAを計算
	fastEMA := calculateEMA(prices, fastPeriod)
	slowEMA := calculateEMA(prices, slowPeriod)

	// MACD線を計算
	macd = fastEMA - slowEMA

	// シグナル線は通常MACD線の9日EMA（ここでは簡易計算）
	signal = macd * 0.9 // 簡易計算

	// ヒストグラムはMACDからシグナルを引いた値
	histogram = macd - signal

	return macd, signal, histogram
}

// CalculateAllIndicators calculates all technical indicators for a stock
func CalculateAllIndicators(prices []models.StockPrice) *models.TechnicalIndicator {
	if len(prices) == 0 {
		return nil
	}

	lastPrice := prices[len(prices)-1]

	indicator := &models.TechnicalIndicator{
		Code:      lastPrice.Code,
		MA5:       MovingAverage(prices, 5),
		MA25:      MovingAverage(prices, 25),
		MA75:      MovingAverage(prices, 75),
		RSI:       RSI(prices, 14),
		Timestamp: lastPrice.Timestamp,
	}

	indicator.MACD, indicator.Signal, indicator.Histogram = MACD(prices, 12, 26, 9)

	return indicator
}

// TradingSignal represents buy/sell/hold signal
type TradingSignal struct {
	Action     string  // "buy", "sell", "hold"
	Confidence float64 // 0.0 to 1.0
	Reason     string
	Score      float64
}

// GenerateTradingSignal generates trading signal based on technical indicators
func GenerateTradingSignal(indicator *models.TechnicalIndicator, currentPrice float64) *TradingSignal {
	score := 0.0
	reasons := []string{}

	// RSI based signals
	if indicator.RSI < 30 {
		score += 2.0
		reasons = append(reasons, "RSI oversold")
	} else if indicator.RSI > 70 {
		score -= 2.0
		reasons = append(reasons, "RSI overbought")
	}

	// Moving Average signals
	if indicator.MA5 > indicator.MA25 && indicator.MA25 > indicator.MA75 {
		score += 1.5
		reasons = append(reasons, "Bullish MA alignment")
	} else if indicator.MA5 < indicator.MA25 && indicator.MA25 < indicator.MA75 {
		score -= 1.5
		reasons = append(reasons, "Bearish MA alignment")
	}

	// MACD signals
	if indicator.MACD > indicator.Signal && indicator.Histogram > 0 {
		score += 1.0
		reasons = append(reasons, "MACD bullish")
	} else if indicator.MACD < indicator.Signal && indicator.Histogram < 0 {
		score -= 1.0
		reasons = append(reasons, "MACD bearish")
	}

	// Price vs Moving Average
	if currentPrice > indicator.MA5 && currentPrice > indicator.MA25 {
		score += 0.5
		reasons = append(reasons, "Price above key MAs")
	} else if currentPrice < indicator.MA5 && currentPrice < indicator.MA25 {
		score -= 0.5
		reasons = append(reasons, "Price below key MAs")
	}

	// Determine action and confidence
	var action string
	confidence := math.Min(math.Abs(score)/5.0, 1.0) // Normalize to 0-1

	if score > 1.0 {
		action = "buy"
	} else if score < -1.0 {
		action = "sell"
	} else {
		action = "hold"
		confidence = 1.0 - confidence // High confidence in hold when score is near 0
	}

	reasonText := ""
	if len(reasons) > 0 {
		reasonText = reasons[0]
		if len(reasons) > 1 {
			reasonText += " and others"
		}
	}

	return &TradingSignal{
		Action:     action,
		Confidence: confidence,
		Reason:     reasonText,
		Score:      score,
	}
}
