package main

import (
	"log"
	"time"

	"github.com/boost-jp/stock-automation/internal/database"
	"github.com/boost-jp/stock-automation/internal/models"
)

func main() {
	log.Println("💼 サンプルポートフォリオデータ追加")

	// データベース接続
	db, err := database.NewDB()
	if err != nil {
		log.Fatal("データベース接続エラー:", err)
	}
	defer db.Close()

	// サンプルポートフォリオデータ
	samplePortfolio := []models.Portfolio{
		{
			Code:          "7203",
			Name:          "トヨタ自動車",
			Shares:        100,
			PurchasePrice: 2800.0,
			PurchaseDate:  time.Now().AddDate(0, -2, 0), // 2ヶ月前
		},
		{
			Code:          "6758",
			Name:          "ソニーグループ",
			Shares:        50,
			PurchasePrice: 12000.0,
			PurchaseDate:  time.Now().AddDate(0, -1, -15), // 1ヶ月15日前
		},
		{
			Code:          "9984",
			Name:          "ソフトバンクグループ",
			Shares:        80,
			PurchasePrice: 5500.0,
			PurchaseDate:  time.Now().AddDate(0, -3, -10), // 3ヶ月10日前
		},
	}

	// 既存のポートフォリオデータをクリア
	err = db.GetDB().Where("1 = 1").Delete(&models.Portfolio{}).Error
	if err != nil {
		log.Printf("⚠️  既存データクリアエラー: %v", err)
	} else {
		log.Println("🗑️  既存ポートフォリオデータをクリアしました")
	}

	// サンプルデータ挿入
	for _, portfolio := range samplePortfolio {
		err := db.GetDB().Create(&portfolio).Error
		if err != nil {
			log.Printf("❌ %s (%s) 追加エラー: %v", portfolio.Name, portfolio.Code, err)
		} else {
			log.Printf("✅ %s (%s): %d株 @ ¥%.0f",
				portfolio.Name, portfolio.Code, portfolio.Shares, portfolio.PurchasePrice)
		}
	}

	log.Println("\n🎉 サンプルポートフォリオデータ追加完了")
	log.Println("💡 daily_report_testerを実行して詳細レポートを確認してください")
}
