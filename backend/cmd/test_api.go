package main

import (
	"fmt"
	"log"

	"stock-automation/internal/api"
	"stock-automation/internal/database"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	// データベース接続
	db, err := database.NewDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// データコレクター初期化
	collector := api.NewDataCollector(db)

	// 監視銘柄とポートフォリオの読み込み
	if err := collector.UpdateWatchList(); err != nil {
		log.Fatal("Failed to update watch list:", err)
	}
	if err := collector.UpdatePortfolio(); err != nil {
		log.Fatal("Failed to update portfolio:", err)
	}

	// Yahoo Finance APIクライアント作成
	yahooClient := api.NewYahooFinanceClient()

	// トヨタ自動車(7203)の株価を取得してテスト
	fmt.Println("Testing Yahoo Finance API...")
	price, err := yahooClient.GetCurrentPrice("7203")
	if err != nil {
		log.Fatal("Failed to get current price:", err)
	}

	fmt.Printf("Stock: %s\n", price.Code)
	fmt.Printf("Price: ¥%.2f\n", price.Price)
	fmt.Printf("Volume: %d\n", price.Volume)
	fmt.Printf("High: ¥%.2f\n", price.High)
	fmt.Printf("Low: ¥%.2f\n", price.Low)
	fmt.Printf("Open: ¥%.2f\n", price.Open)

	// データベースに保存
	price.Name = "トヨタ自動車"
	if err := db.SaveStockPrice(price); err != nil {
		log.Fatal("Failed to save stock price:", err)
	}

	fmt.Println("✅ Stock price saved to database successfully!")

	// 全銘柄の価格更新をテスト
	fmt.Println("\nTesting price update for all stocks...")
	if err := collector.UpdateAllPrices(); err != nil {
		log.Fatal("Failed to update all prices:", err)
	}

	fmt.Println("✅ All prices updated successfully!")

	// 履歴データ取得テスト
	fmt.Println("\nTesting historical data collection...")
	if err := collector.CollectHistoricalData("7203", 30); err != nil {
		log.Fatal("Failed to collect historical data:", err)
	}

	fmt.Println("✅ Historical data collection completed!")
}