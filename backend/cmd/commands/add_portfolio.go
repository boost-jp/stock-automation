package commands

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/boost-jp/stock-automation/app/infrastructure/client"
	"github.com/boost-jp/stock-automation/app/infrastructure/dao"
	"github.com/boost-jp/stock-automation/app/infrastructure/database"
	"github.com/boost-jp/stock-automation/internal/ulid"
)

// RunAddSamplePortfolio runs the add sample portfolio command.
func RunAddSamplePortfolio(connMgr database.ConnectionManager, args []string) {
	// Command line flags
	portfolioCmd := flag.NewFlagSet("add-portfolio", flag.ExitOnError)
	var (
		clearExisting = portfolioCmd.Bool("clear", false, "Clear existing portfolio data before adding")
	)

	portfolioCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: stock-automation add-portfolio [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Add sample portfolio data for testing\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		portfolioCmd.PrintDefaults()
	}

	if err := portfolioCmd.Parse(args); err != nil {
		log.Fatal(err)
	}

	log.Println("💼 サンプルポートフォリオデータ追加")

	db := connMgr.GetDB()
	ctx := context.Background()

	// サンプルポートフォリオデータ
	samplePortfolio := []dao.Portfolio{
		{
			ID:            ulid.NewULID(),
			Code:          "7203",
			Name:          "トヨタ自動車",
			Shares:        100,
			PurchasePrice: client.FloatToDecimal(2800.0),
			PurchaseDate:  time.Now().AddDate(0, -2, 0), // 2ヶ月前
		},
		{
			ID:            ulid.NewULID(),
			Code:          "6758",
			Name:          "ソニーグループ",
			Shares:        50,
			PurchasePrice: client.FloatToDecimal(12000.0),
			PurchaseDate:  time.Now().AddDate(0, -1, -15), // 1ヶ月15日前
		},
		{
			ID:            ulid.NewULID(),
			Code:          "9984",
			Name:          "ソフトバンクグループ",
			Shares:        80,
			PurchasePrice: client.FloatToDecimal(5500.0),
			PurchaseDate:  time.Now().AddDate(0, -3, -10), // 3ヶ月10日前
		},
	}

	// 既存のポートフォリオデータをクリア（オプション）
	if *clearExisting {
		_, err := dao.Portfolios(qm.Where("1 = 1")).DeleteAll(ctx, db)
		if err != nil {
			log.Printf("⚠️  既存データクリアエラー: %v", err)
		} else {
			log.Println("🗑️  既存ポートフォリオデータをクリアしました")
		}
	}

	// サンプルデータ挿入
	for _, portfolio := range samplePortfolio {
		err := portfolio.Insert(ctx, db, boil.Infer())
		if err != nil {
			log.Printf("❌ %s (%s) 追加エラー: %v", portfolio.Name, portfolio.Code, err)
		} else {
			purchasePrice, _ := portfolio.PurchasePrice.Float64()
			log.Printf("✅ %s (%s): %d株 @ ¥%.0f",
				portfolio.Name, portfolio.Code, portfolio.Shares, purchasePrice)
		}
	}

	// 統計情報を表示
	count, err := dao.Portfolios().Count(ctx, db)
	if err != nil {
		log.Printf("⚠️  統計取得エラー: %v", err)
	} else {
		log.Printf("\n📊 ポートフォリオ銘柄数: %d", count)
	}

	log.Println("\n🎉 サンプルポートフォリオデータ追加完了")
	log.Println("💡 daily-report コマンドを実行して詳細レポートを確認してください")
}

