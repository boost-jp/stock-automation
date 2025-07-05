package commands

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/boost-jp/stock-automation/app/infrastructure/client"
)

// RunTestYahooAPI runs the Yahoo Finance API test command.
func RunTestYahooAPI(args []string) {
	// Command line flags
	testCmd := flag.NewFlagSet("test-yahoo", flag.ExitOnError)
	var (
		stockCode = testCmd.String("code", "7203", "Stock code to test")
		days      = testCmd.Int("days", 30, "Number of days of historical data to fetch")
	)

	testCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: stock-automation test-yahoo [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Test Yahoo Finance API integration\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		testCmd.PrintDefaults()
	}

	if err := testCmd.Parse(args); err != nil {
		log.Fatal(err)
	}

	// Initialize Yahoo Finance client
	client := client.NewYahooFinanceClient()

	// Test current price fetch
	log.Println("📊 Yahoo Finance API統合テスト")
	log.Println("================================")

	log.Printf("🔍 %s の現在価格を取得中...\n", *stockCode)

	currentPrice, err := client.GetCurrentPrice(*stockCode)
	if err != nil {
		log.Printf("❌ 現在価格取得エラー: %v", err)
	} else {
		log.Printf("✅ 現在価格取得成功:")
		log.Printf("   コード: %s", currentPrice.Code)
		log.Printf("   日付: %s", currentPrice.Date.Format("2006-01-02"))
		closePrice, _ := currentPrice.ClosePrice.Float64()
		highPrice, _ := currentPrice.HighPrice.Float64()
		lowPrice, _ := currentPrice.LowPrice.Float64()
		openPrice, _ := currentPrice.OpenPrice.Float64()

		log.Printf("   終値: ¥%.2f", closePrice)
		log.Printf("   出来高: %d", currentPrice.Volume)
		log.Printf("   高値: ¥%.2f", highPrice)
		log.Printf("   安値: ¥%.2f", lowPrice)
		log.Printf("   始値: ¥%.2f", openPrice)
	}

	log.Println()
	log.Printf("📈 %s の過去%d日分のデータを取得中...\n", *stockCode, *days)

	historicalData, err := client.GetHistoricalData(*stockCode, *days)
	if err != nil {
		log.Printf("❌ 履歴データ取得エラー: %v", err)
	} else {
		log.Printf("✅ 履歴データ取得成功: %d件\n", len(historicalData))
		if len(historicalData) > 0 {
			firstClose, _ := historicalData[0].ClosePrice.Float64()
			lastClose, _ := historicalData[len(historicalData)-1].ClosePrice.Float64()

			log.Printf("   最古: %s (¥%.2f)\n",
				historicalData[0].Date.Format("2006-01-02"),
				firstClose)
			log.Printf("   最新: %s (¥%.2f)\n",
				historicalData[len(historicalData)-1].Date.Format("2006-01-02"),
				lastClose)

			// 最初の5件を表示
			log.Println("\n   最初の5件:")
			for i := 0; i < 5 && i < len(historicalData); i++ {
				data := historicalData[i]
				open, _ := data.OpenPrice.Float64()
				high, _ := data.HighPrice.Float64()
				low, _ := data.LowPrice.Float64()
				close, _ := data.ClosePrice.Float64()

				log.Printf("   %s: O:%.2f H:%.2f L:%.2f C:%.2f V:%d",
					data.Date.Format("2006-01-02"),
					open, high, low, close,
					data.Volume)
			}
		}
	}

	log.Println()
	log.Println("🎉 Yahoo Finance API統合テスト完了！")
}

