package main

import (
	"fmt"
	"log"

	"stock-automation/internal/analysis"
	"stock-automation/internal/database"
)

func main() {
	// データベース接続
	db, err := database.NewDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// トヨタ自動車(7203)の過去30日の株価データを取得
	stockCode := "7203"
	fmt.Printf("Testing technical analysis for %s...\n", stockCode)
	
	prices, err := db.GetPriceHistory(stockCode, 30)
	if err != nil {
		log.Fatal("Failed to get price history:", err)
	}
	
	if len(prices) == 0 {
		log.Fatal("No price data found for", stockCode)
	}
	
	fmt.Printf("Found %d price records\n", len(prices))
	
	// テクニカル指標を計算
	indicator := analysis.CalculateAllIndicators(prices)
	if indicator == nil {
		log.Fatal("Failed to calculate technical indicators")
	}
	
	// 結果を表示
	fmt.Println("\n📊 Technical Indicators:")
	fmt.Printf("MA5:        %.2f\n", indicator.MA5)
	fmt.Printf("MA25:       %.2f\n", indicator.MA25)
	fmt.Printf("MA75:       %.2f\n", indicator.MA75)
	fmt.Printf("RSI:        %.2f\n", indicator.RSI)
	fmt.Printf("MACD:       %.4f\n", indicator.MACD)
	fmt.Printf("Signal:     %.4f\n", indicator.Signal)
	fmt.Printf("Histogram:  %.4f\n", indicator.Histogram)
	
	// 現在価格を取得
	latestPrice, err := db.GetLatestPrice(stockCode)
	if err != nil {
		log.Fatal("Failed to get latest price:", err)
	}
	
	// 売買シグナルを生成
	signal := analysis.GenerateTradingSignal(indicator, latestPrice.Price)
	
	fmt.Println("\n🎯 Trading Signal:")
	fmt.Printf("Action:     %s\n", signal.Action)
	fmt.Printf("Confidence: %.2f%%\n", signal.Confidence*100)
	fmt.Printf("Score:      %.2f\n", signal.Score)
	fmt.Printf("Reason:     %s\n", signal.Reason)
	
	// アクションに応じた絵文字を表示
	var emoji string
	switch signal.Action {
	case "buy":
		emoji = "🟢"
	case "sell":
		emoji = "🔴"
	default:
		emoji = "🟡"
	}
	
	fmt.Printf("\n%s %s信号 (信頼度: %.1f%%)\n", emoji, signal.Action, signal.Confidence*100)
	
	// データベースにテクニカル指標を保存
	if err := db.SaveTechnicalIndicator(indicator); err != nil {
		log.Fatal("Failed to save technical indicator:", err)
	}
	
	fmt.Println("\n✅ Technical indicators saved to database successfully!")
}