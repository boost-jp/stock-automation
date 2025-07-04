package main

import (
	"fmt"
	"log"
	"stock-automation/internal/api"
)

func main() {
	// Initialize Yahoo Finance client
	client := api.NewYahooFinanceClient()

	// Test current price fetch
	fmt.Println("📊 Yahoo Finance API統合テスト")
	fmt.Println("================================")

	testCode := "7203" // Toyota
	fmt.Printf("🔍 %s の現在価格を取得中...\n", testCode)

	currentPrice, err := client.GetCurrentPrice(testCode)
	if err != nil {
		log.Printf("❌ 現在価格取得エラー: %v", err)
	} else {
		fmt.Printf("✅ 現在価格: ¥%.2f\n", currentPrice.Price)
		fmt.Printf("   出来高: %d\n", currentPrice.Volume)
		fmt.Printf("   高値: ¥%.2f\n", currentPrice.High)
		fmt.Printf("   安値: ¥%.2f\n", currentPrice.Low)
	}

	fmt.Println()
	fmt.Printf("📈 %s の過去30日分のデータを取得中...\n", testCode)

	historicalData, err := client.GetHistoricalData(testCode, 30)
	if err != nil {
		log.Printf("❌ 履歴データ取得エラー: %v", err)
	} else {
		fmt.Printf("✅ 履歴データ取得成功: %d件\n", len(historicalData))
		if len(historicalData) > 0 {
			fmt.Printf("   最古: %s (¥%.2f)\n",
				historicalData[0].Timestamp.Format("2006-01-02"),
				historicalData[0].Close)
			fmt.Printf("   最新: %s (¥%.2f)\n",
				historicalData[len(historicalData)-1].Timestamp.Format("2006-01-02"),
				historicalData[len(historicalData)-1].Close)
		}
	}

	fmt.Println()
	fmt.Println("🎉 Yahoo Finance API統合テスト完了")
}
