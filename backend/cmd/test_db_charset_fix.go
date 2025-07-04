package main

import (
	"fmt"
	"log"
	"time"

	"stock-automation/internal/database"
	"stock-automation/internal/models"
)

func testDBCharsetFix() {
	fmt.Println("🔍 データベース文字エンコーディング修正テスト開始")

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
		Code:      "TEST01",
		Name:      "テスト株式会社",
		Price:     1000.00,
		Volume:    50000,
		High:      1100.00,
		Low:       950.00,
		Open:      980.00,
		Close:     1000.00,
		Timestamp: time.Now(),
	}

	// データベースに保存
	fmt.Println("📝 日本語データを保存中...")
	if err := db.SaveStockPrice(testPrice); err != nil {
		log.Fatal("Failed to save test data:", err)
	}

	// データベースから取得
	fmt.Println("📖 日本語データを取得中...")
	retrievedPrice, err := db.GetLatestPrice("TEST01")
	if err != nil {
		log.Fatal("Failed to retrieve test data:", err)
	}

	// 結果確認
	fmt.Printf("保存した名前: %s\n", testPrice.Name)
	fmt.Printf("取得した名前: %s\n", retrievedPrice.Name)
	fmt.Printf("名前が一致: %t\n", testPrice.Name == retrievedPrice.Name)

	// バイト数比較
	fmt.Printf("保存した名前（バイト長）: %d bytes\n", len(testPrice.Name))
	fmt.Printf("取得した名前（バイト長）: %d bytes\n", len(retrievedPrice.Name))

	// 複雑な日本語文字列でテスト
	complexName := "株式会社ソニーグループ 🇯🇵 ※東証一部上場※"
	testPrice2 := &models.StockPrice{
		Code:      "TEST02",
		Name:      complexName,
		Price:     2000.00,
		Volume:    30000,
		High:      2100.00,
		Low:       1900.00,
		Open:      1980.00,
		Close:     2000.00,
		Timestamp: time.Now(),
	}

	fmt.Println("\n🚀 複雑な日本語文字列テスト...")
	if err := db.SaveStockPrice(testPrice2); err != nil {
		log.Fatal("Failed to save complex test data:", err)
	}

	retrievedPrice2, err := db.GetLatestPrice("TEST02")
	if err != nil {
		log.Fatal("Failed to retrieve complex test data:", err)
	}

	fmt.Printf("保存した複雑な名前: %s\n", complexName)
	fmt.Printf("取得した複雑な名前: %s\n", retrievedPrice2.Name)
	fmt.Printf("複雑な名前が一致: %t\n", complexName == retrievedPrice2.Name)

	if testPrice.Name == retrievedPrice.Name && complexName == retrievedPrice2.Name {
		fmt.Println("\n✅ データベース文字エンコーディング修正成功!")
	} else {
		fmt.Println("\n❌ データベース文字エンコーディングにまだ問題があります")
	}

	fmt.Println("\n🎉 テスト完了")
}