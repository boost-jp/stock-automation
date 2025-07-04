package main

import (
	"fmt"
	"log"
	"time"

	"stock-automation/internal/database"
	"stock-automation/internal/models"
)

func testDBEncoding() {
	fmt.Println("🔍 データベース文字エンコーディングテスト開始")

	// データベース接続
	db, err := database.NewDB()
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	// マイグレーション実行
	if err := db.AutoMigrate(); err != nil {
		log.Fatal("Migration failed:", err)
	}

	// テストデータ作成
	testPrice := &models.StockPrice{
		Code:      "7203",
		Name:      "トヨタ自動車",
		Price:     2484.50,
		Volume:    1000000,
		High:      2500.00,
		Low:       2450.00,
		Open:      2470.00,
		Close:     2484.50,
		Timestamp: time.Now(),
	}

	// データベースに保存
	fmt.Println("📝 日本語データを保存中...")
	if err := db.SaveStockPrice(testPrice); err != nil {
		log.Fatal("Failed to save test data:", err)
	}

	// データベースから取得
	fmt.Println("📖 日本語データを取得中...")
	retrievedPrice, err := db.GetLatestPrice("7203")
	if err != nil {
		log.Fatal("Failed to retrieve test data:", err)
	}

	// 結果確認
	fmt.Printf("🔍 保存したデータ: %+v\n", testPrice)
	fmt.Printf("📋 取得したデータ: %+v\n", retrievedPrice)

	// 文字化け確認
	fmt.Println("\n🧪 文字化けチェック:")
	fmt.Printf("保存した名前: %s\n", testPrice.Name)
	fmt.Printf("取得した名前: %s\n", retrievedPrice.Name)
	fmt.Printf("名前が一致: %t\n", testPrice.Name == retrievedPrice.Name)

	// 文字列比較（バイト単位）
	fmt.Printf("保存した名前（バイト長）: %d bytes\n", len(testPrice.Name))
	fmt.Printf("取得した名前（バイト長）: %d bytes\n", len(retrievedPrice.Name))

	// 追加テスト: 複雑な日本語文字列
	complexName := "🇯🇵 ソニーグループ㈱【技術革新企業】※東証一部上場※"
	testPrice2 := &models.StockPrice{
		Code:      "6758",
		Name:      complexName,
		Price:     3688.00,
		Volume:    500000,
		High:      3700.00,
		Low:       3650.00,
		Open:      3680.00,
		Close:     3688.00,
		Timestamp: time.Now(),
	}

	fmt.Println("\n🚀 複雑な日本語文字列テスト...")
	if err := db.SaveStockPrice(testPrice2); err != nil {
		log.Fatal("Failed to save complex test data:", err)
	}

	retrievedPrice2, err := db.GetLatestPrice("6758")
	if err != nil {
		log.Fatal("Failed to retrieve complex test data:", err)
	}

	fmt.Printf("保存した複雑な名前: %s\n", complexName)
	fmt.Printf("取得した複雑な名前: %s\n", retrievedPrice2.Name)
	fmt.Printf("複雑な名前が一致: %t\n", complexName == retrievedPrice2.Name)

	if testPrice.Name == retrievedPrice.Name && complexName == retrievedPrice2.Name {
		fmt.Println("\n✅ データベース文字エンコーディング正常!")
	} else {
		fmt.Println("\n❌ データベース文字エンコーディングに問題があります")
		fmt.Println("DSN設定やデータベース設定を確認してください")
	}

	fmt.Println("\n🎉 テスト完了")
}

func main() {
	testDBEncoding()
}