package domain

import (
	"fmt"
	"math"
	"time"

	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/boost-jp/stock-automation/app/domain/models"
)

// TechnicalAnalysisService handles technical analysis business logic.
type TechnicalAnalysisService struct{}

// NewTechnicalAnalysisService creates a new technical analysis service.
func NewTechnicalAnalysisService() *TechnicalAnalysisService {
	return &TechnicalAnalysisService{}
}

// StockPriceData represents simplified stock price data for calculations.
type StockPriceData struct {
	Code      string
	Date      time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    int64
	Timestamp time.Time
}

// TechnicalIndicatorData represents calculated technical indicators.
type TechnicalIndicatorData struct {
	Code      string
	MA5       float64
	MA25      float64
	MA75      float64
	RSI       float64
	MACD      float64
	Signal    float64
	Histogram float64
	Timestamp time.Time
}

// TradingSignal represents buy/sell/hold signal.
type TradingSignal struct {
	Action     string  // "buy", "sell", "hold"
	Confidence float64 // 0.0 to 1.0
	Reason     string
	Score      float64
}

// ConvertStockPrices converts SQLBoiler models to domain service format.
func (s *TechnicalAnalysisService) ConvertStockPrices(prices []*models.StockPrice) []StockPriceData {
	result := make([]StockPriceData, len(prices))
	for i, p := range prices {
		result[i] = StockPriceData{
			Code:      p.Code,
			Date:      p.Date,
			Open:      s.decimalToFloat(p.OpenPrice),
			High:      s.decimalToFloat(p.HighPrice),
			Low:       s.decimalToFloat(p.LowPrice),
			Close:     s.decimalToFloat(p.ClosePrice),
			Volume:    p.Volume,
			Timestamp: p.Date,
		}
	}

	return result
}

// MovingAverage calculates simple moving average.
func (s *TechnicalAnalysisService) MovingAverage(prices []StockPriceData, period int) float64 {
	if len(prices) < period {
		return 0
	}

	sum := 0.0
	for i := len(prices) - period; i < len(prices); i++ {
		sum += prices[i].Close
	}

	return sum / float64(period)
}

// RSI calculates Relative Strength Index.
func (s *TechnicalAnalysisService) RSI(prices []StockPriceData, period int) float64 {
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

// MACD calculates Moving Average Convergence Divergence.
func (s *TechnicalAnalysisService) MACD(prices []StockPriceData, fastPeriod, slowPeriod, signalPeriod int) (macd, signal, histogram float64) {
	if len(prices) < slowPeriod {
		return 0, 0, 0
	}

	// EMA計算のヘルパー関数
	calculateEMA := func(data []StockPriceData, period int) float64 {
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

// CalculateAllIndicators calculates all technical indicators for a stock.
func (s *TechnicalAnalysisService) CalculateAllIndicators(prices []StockPriceData) *TechnicalIndicatorData {
	if len(prices) == 0 {
		return nil
	}

	lastPrice := prices[len(prices)-1]
	macd, signal, histogram := s.MACD(prices, 12, 26, 9)

	indicator := &TechnicalIndicatorData{
		Code:      lastPrice.Code,
		MA5:       s.MovingAverage(prices, 5),
		MA25:      s.MovingAverage(prices, 25),
		MA75:      s.MovingAverage(prices, 75),
		RSI:       s.RSI(prices, 14),
		MACD:      macd,
		Signal:    signal,
		Histogram: histogram,
		Timestamp: lastPrice.Timestamp,
	}

	return indicator
}

// GenerateTradingSignal generates trading signal based on technical indicators.
func (s *TechnicalAnalysisService) GenerateTradingSignal(indicator *TechnicalIndicatorData, currentPrice float64) *TradingSignal {
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

// ConvertToModelIndicator converts domain indicator to SQLBoiler model.
func (s *TechnicalAnalysisService) ConvertToModelIndicator(data *TechnicalIndicatorData) *models.TechnicalIndicator {
	return &models.TechnicalIndicator{
		Code: data.Code,
		// Note: You'll need to convert float64 back to types.Decimal for storage
		// This is left as an exercise based on your specific decimal library
		Rsi14:         s.floatToDecimal(data.RSI),
		Macd:          s.floatToDecimal(data.MACD),
		MacdSignal:    s.floatToDecimal(data.Signal),
		MacdHistogram: s.floatToDecimal(data.Histogram),
		Sma5:          s.floatToDecimal(data.MA5),
		Sma25:         s.floatToDecimal(data.MA25),
		Sma75:         s.floatToDecimal(data.MA75),
		Date:          data.Timestamp,
	}
}

// decimalToFloat converts types.Decimal to float64.
func (s *TechnicalAnalysisService) decimalToFloat(d any) float64 {
	// This is a simplified conversion for testing
	// In a real implementation, you would use the actual decimal library methods
	switch v := d.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case string:
		var f float64
		fmt.Sscanf(v, "%f", &f)
		return f
	default:
		// For types.Decimal, we'll use string conversion as fallback
		str := fmt.Sprintf("%v", d)
		var f float64
		fmt.Sscanf(str, "%f", &f)
		return f
	}
}

// floatToDecimal converts float64 to types.NullDecimal.
func (s *TechnicalAnalysisService) floatToDecimal(f float64) types.NullDecimal {
	// Simplified placeholder conversion - returns empty NullDecimal
	return types.NullDecimal{}
}

// ValidateIndicator validates technical indicator values.
func (s *TechnicalAnalysisService) ValidateIndicator(indicator *TechnicalIndicatorData) error {
	if indicator.Code == "" {
		return fmt.Errorf("銘柄コードは必須です")
	}

	if indicator.RSI < 0 || indicator.RSI > 100 {
		return fmt.Errorf("RSIは0から100の範囲である必要があります")
	}

	return nil
}

// GetSignalStrength returns signal strength description.
func (s *TechnicalAnalysisService) GetSignalStrength(indicator *TechnicalIndicatorData) string {
	buySignals := 0
	sellSignals := 0

	// RSI signals
	if indicator.RSI < 30 {
		buySignals++
	} else if indicator.RSI > 70 {
		sellSignals++
	}

	// MACD signals
	if indicator.MACD > indicator.Signal && indicator.Histogram > 0 {
		buySignals++
	} else if indicator.MACD < indicator.Signal && indicator.Histogram < 0 {
		sellSignals++
	}

	if buySignals > sellSignals {
		return "Strong Buy"
	} else if sellSignals > buySignals {
		return "Strong Sell"
	}

	return "Neutral"
}
