package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/boost-jp/stock-automation/app/domain/models"
	"github.com/boost-jp/stock-automation/app/utility"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// StockDataClient defines the interface for stock data providers.
type StockDataClient interface {
	GetCurrentPrice(stockCode string) (*models.StockPrice, error)
	GetHistoricalData(stockCode string, days int) ([]*models.StockPrice, error)
	GetIntradayData(stockCode string, interval string) ([]*models.StockPrice, error)
}

// YahooFinanceClient implements StockDataClient using Yahoo Finance API.
type YahooFinanceClient struct {
	client      *resty.Client
	baseURL     string
	rateLimiter *RateLimiter
}

// Yahoo Finance APIレスポンス構造.
type YahooFinanceResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Symbol               string  `json:"symbol"`
				RegularMarketPrice   float64 `json:"regularMarketPrice"`
				PreviousClose        float64 `json:"previousClose"`
				RegularMarketOpen    float64 `json:"regularMarketOpen"`
				RegularMarketDayLow  float64 `json:"regularMarketDayLow"`
				RegularMarketDayHigh float64 `json:"regularMarketDayHigh"`
				RegularMarketVolume  int64   `json:"regularMarketVolume"`
				Currency             string  `json:"currency"`
				ExchangeName         string  `json:"exchangeName"`
			} `json:"meta"`
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []float64 `json:"open"`
					High   []float64 `json:"high"`
					Low    []float64 `json:"low"`
					Close  []float64 `json:"close"`
					Volume []int64   `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
		Error interface{} `json:"error"`
	} `json:"chart"`
}

// YahooFinanceConfig holds Yahoo Finance client configuration.
type YahooFinanceConfig struct {
	BaseURL       string
	Timeout       time.Duration
	RetryCount    int
	RetryWaitTime time.Duration
	RetryMaxWait  time.Duration
	UserAgent     string
	RateLimitRPS  int
}

// NewYahooFinanceClient creates a new Yahoo Finance client.
func NewYahooFinanceClient() *YahooFinanceClient {
	return NewYahooFinanceClientWithConfig(DefaultYahooFinanceConfig())
}

// NewYahooFinanceClientWithConfig creates a new Yahoo Finance client with custom configuration.
func NewYahooFinanceClientWithConfig(config YahooFinanceConfig) *YahooFinanceClient {
	client := resty.New()
	client.SetTimeout(config.Timeout)
	client.SetRetryCount(config.RetryCount)
	client.SetRetryWaitTime(config.RetryWaitTime)
	client.SetRetryMaxWaitTime(config.RetryMaxWait)

	// Add exponential backoff for retries
	client.AddRetryCondition(func(r *resty.Response, err error) bool {
		// Retry on server errors or rate limit
		if r != nil && (r.StatusCode() >= 500 || r.StatusCode() == 429) {
			return true
		}
		// Retry on network errors
		if err != nil && IsRetryableError(err) {
			return true
		}
		return false
	})

	return &YahooFinanceClient{
		client:      client,
		baseURL:     config.BaseURL,
		rateLimiter: NewRateLimiter(config.RateLimitRPS),
	}
}

// DefaultYahooFinanceConfig returns default configuration for Yahoo Finance client.
func DefaultYahooFinanceConfig() YahooFinanceConfig {
	return YahooFinanceConfig{
		BaseURL:       "https://query1.finance.yahoo.com",
		Timeout:       30 * time.Second,
		RetryCount:    3,
		RetryWaitTime: 1 * time.Second,
		RetryMaxWait:  10 * time.Second,
		UserAgent:     "Mozilla/5.0 (compatible; StockAutomation/1.0)",
		RateLimitRPS:  10,
	}
}

// GetCurrentPrice retrieves real-time stock price.
func (y *YahooFinanceClient) GetCurrentPrice(stockCode string) (*models.StockPrice, error) {
	// Apply rate limiting
	if err := y.rateLimiter.Wait(context.Background()); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	url := fmt.Sprintf("%s/v8/finance/chart/%s.T", y.baseURL, stockCode)

	resp, err := y.client.R().
		SetHeader("User-Agent", "Mozilla/5.0 (compatible; StockAutomation/1.0)").
		Get(url)
	if err != nil {
		if IsRetryableError(err) {
			return nil, fmt.Errorf("temporary error fetching data for %s: %w", stockCode, err)
		}
		return nil, fmt.Errorf("failed to fetch data for %s: %w", stockCode, err)
	}

	if resp.StatusCode() != 200 {
		if httpErr := ClassifyHTTPError(resp.StatusCode()); httpErr != nil {
			return nil, fmt.Errorf("API error for %s: %w (status: %d)", stockCode, httpErr, resp.StatusCode())
		}
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode())
	}

	var response YahooFinanceResponse
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data found for stock code: %s", stockCode)
	}

	result := response.Chart.Result[0]
	meta := result.Meta

	stockPrice := &models.StockPrice{
		ID:         utility.NewULID(),
		Code:       stockCode,
		Date:       time.Now(),
		OpenPrice:  floatToDecimal(meta.RegularMarketOpen),
		HighPrice:  floatToDecimal(meta.RegularMarketDayHigh),
		LowPrice:   floatToDecimal(meta.RegularMarketDayLow),
		ClosePrice: floatToDecimal(meta.RegularMarketPrice),
		Volume:     meta.RegularMarketVolume,
	}

	logrus.WithFields(logrus.Fields{
		"code":  stockCode,
		"price": stockPrice.ClosePrice,
	}).Debug("Yahoo Finance current price fetched")

	return stockPrice, nil
}

// GetHistoricalData retrieves historical stock price data.
func (y *YahooFinanceClient) GetHistoricalData(stockCode string, days int) ([]*models.StockPrice, error) {
	// Apply rate limiting
	if err := y.rateLimiter.Wait(context.Background()); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	endTime := time.Now().Unix()
	startTime := time.Now().AddDate(0, 0, -days).Unix()

	url := fmt.Sprintf("%s/v8/finance/chart/%s.T", y.baseURL, stockCode)

	resp, err := y.client.R().
		SetQueryParams(map[string]string{
			"period1":  strconv.FormatInt(startTime, 10),
			"period2":  strconv.FormatInt(endTime, 10),
			"interval": "1d",
		}).
		SetHeader("User-Agent", "Mozilla/5.0 (compatible; StockAutomation/1.0)").
		Get(url)
	if err != nil {
		if IsRetryableError(err) {
			return nil, fmt.Errorf("temporary error fetching historical data for %s: %w", stockCode, err)
		}
		return nil, fmt.Errorf("failed to fetch historical data for %s: %w", stockCode, err)
	}

	var response YahooFinanceResponse
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.Chart.Result) == 0 {
		return nil, fmt.Errorf("no historical data found for: %s", stockCode)
	}

	result := response.Chart.Result[0]

	// Check for API errors
	if response.Chart.Error != nil {
		return nil, fmt.Errorf("Yahoo Finance API error for %s: %v", stockCode, response.Chart.Error)
	}

	timestamps := result.Timestamp

	if len(result.Indicators.Quote) == 0 {
		return nil, fmt.Errorf("no quote indicators found for: %s", stockCode)
	}

	quotes := result.Indicators.Quote[0]

	var prices []*models.StockPrice

	for i, ts := range timestamps {
		// Skip invalid or missing data points
		if i >= len(quotes.Close) || i >= len(quotes.Open) || i >= len(quotes.High) ||
			i >= len(quotes.Low) || i >= len(quotes.Volume) {
			continue
		}

		// Skip zero or negative prices (invalid data)
		if quotes.Close[i] <= 0 || quotes.Open[i] <= 0 || quotes.High[i] <= 0 || quotes.Low[i] <= 0 {
			continue
		}

		price := &models.StockPrice{
			ID:         utility.NewULID(),
			Code:       stockCode,
			Date:       time.Unix(ts, 0),
			OpenPrice:  floatToDecimal(quotes.Open[i]),
			HighPrice:  floatToDecimal(quotes.High[i]),
			LowPrice:   floatToDecimal(quotes.Low[i]),
			ClosePrice: floatToDecimal(quotes.Close[i]),
			Volume:     quotes.Volume[i],
		}

		prices = append(prices, price)
	}

	logrus.WithFields(logrus.Fields{
		"code":    stockCode,
		"records": len(prices),
	}).Debug("Yahoo Finance historical data fetched")

	return prices, nil
}

// GetIntradayData retrieves intraday stock price data.
func (y *YahooFinanceClient) GetIntradayData(stockCode string, interval string) ([]*models.StockPrice, error) {
	// Apply rate limiting
	if err := y.rateLimiter.Wait(context.Background()); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	url := fmt.Sprintf("%s/v8/finance/chart/%s.T", y.baseURL, stockCode)

	resp, err := y.client.R().
		SetQueryParams(map[string]string{
			"range":    "1d",
			"interval": interval,
		}).
		SetHeader("User-Agent", "Mozilla/5.0 (compatible; StockAutomation/1.0)").
		Get(url)
	if err != nil {
		if IsRetryableError(err) {
			return nil, fmt.Errorf("temporary error fetching intraday data for %s: %w", stockCode, err)
		}
		return nil, fmt.Errorf("failed to fetch intraday data for %s: %w", stockCode, err)
	}

	var response YahooFinanceResponse
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.Chart.Result) == 0 {
		return nil, fmt.Errorf("no intraday data found for: %s", stockCode)
	}

	result := response.Chart.Result[0]
	timestamps := result.Timestamp
	quotes := result.Indicators.Quote[0]

	var prices []*models.StockPrice

	for i, ts := range timestamps {
		if i >= len(quotes.Close) || quotes.Close[i] == 0 {
			continue
		}

		price := &models.StockPrice{
			ID:         utility.NewULID(),
			Code:       stockCode,
			Date:       time.Unix(ts, 0),
			OpenPrice:  floatToDecimal(quotes.Open[i]),
			HighPrice:  floatToDecimal(quotes.High[i]),
			LowPrice:   floatToDecimal(quotes.Low[i]),
			ClosePrice: floatToDecimal(quotes.Close[i]),
			Volume:     quotes.Volume[i],
		}

		prices = append(prices, price)
	}

	return prices, nil
}
