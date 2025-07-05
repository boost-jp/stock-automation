package commands

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/boost-jp/stock-automation/app/domain/models"
	"github.com/boost-jp/stock-automation/app/infrastructure/database"
	"github.com/boost-jp/stock-automation/app/infrastructure/repository"
	"github.com/boost-jp/stock-automation/internal/ulid"
	"github.com/ericlagergren/decimal"
)

// RunWatchListManager runs the watchlist manager command.
func RunWatchListManager(connMgr database.ConnectionManager, args []string) {
	// Command line flags
	watchlistCmd := flag.NewFlagSet("watchlist", flag.ExitOnError)
	var (
		action          = watchlistCmd.String("action", "list", "Action to perform: list, add, update, delete, toggle")
		code            = watchlistCmd.String("code", "", "Stock code")
		name            = watchlistCmd.String("name", "", "Stock name")
		targetBuyPrice  = watchlistCmd.Float64("buy", 0, "Target buy price")
		targetSellPrice = watchlistCmd.Float64("sell", 0, "Target sell price")
		watchListID     = watchlistCmd.String("id", "", "Watch list ID (for update/delete/toggle)")
	)

	watchlistCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: stock-automation watchlist [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Manage watch list items\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		watchlistCmd.PrintDefaults()
	}

	if err := watchlistCmd.Parse(args); err != nil {
		log.Fatal(err)
	}

	// Initialize repository
	db := connMgr.GetDB()
	stockRepo := repository.NewStockRepository(db)

	ctx := context.Background()

	switch *action {
	case "list":
		listWatchList(ctx, stockRepo)
	case "add":
		if *code == "" || *name == "" {
			log.Fatal("Stock code and name are required for add action")
		}
		addToWatchList(ctx, stockRepo, *code, *name, *targetBuyPrice, *targetSellPrice)
	case "update":
		if *watchListID == "" {
			log.Fatal("Watch list ID is required for update action")
		}
		updateWatchList(ctx, stockRepo, *watchListID, *targetBuyPrice, *targetSellPrice)
	case "delete":
		if *watchListID == "" {
			log.Fatal("Watch list ID is required for delete action")
		}
		deleteFromWatchList(ctx, stockRepo, *watchListID)
	case "toggle":
		if *watchListID == "" {
			log.Fatal("Watch list ID is required for toggle action")
		}
		toggleWatchList(ctx, stockRepo, *watchListID)
	default:
		log.Fatalf("Unknown action: %s", *action)
	}
}

func listWatchList(ctx context.Context, repo repository.StockRepository) {
	watchList, err := repo.GetActiveWatchList(ctx)
	if err != nil {
		log.Fatal("Failed to get watch list:", err)
	}

	fmt.Println("📊 Active Watch List:")
	fmt.Println("=====================================")
	fmt.Printf("%-36s %-6s %-20s %-10s %-10s\n", "ID", "Code", "Name", "Buy Target", "Sell Target")
	fmt.Println("-------------------------------------")

	for _, item := range watchList {
		buyPrice := "-"
		if item.TargetBuyPrice.Big != nil {
			f, _ := item.TargetBuyPrice.Big.Float64()
			buyPrice = fmt.Sprintf("¥%.0f", f)
		}
		sellPrice := "-"
		if item.TargetSellPrice.Big != nil {
			f, _ := item.TargetSellPrice.Big.Float64()
			sellPrice = fmt.Sprintf("¥%.0f", f)
		}
		fmt.Printf("%-36s %-6s %-20s %-10s %-10s\n", item.ID, item.Code, item.Name, buyPrice, sellPrice)
	}

	fmt.Printf("\nTotal: %d stocks\n", len(watchList))
}

func addToWatchList(ctx context.Context, repo repository.StockRepository, code, name string, buyPrice, sellPrice float64) {
	// Generate new ULID
	id := ulid.NewULID()

	watchItem := &models.WatchList{
		ID:       id,
		Code:     code,
		Name:     name,
		IsActive: null.BoolFrom(true),
	}

	if buyPrice > 0 {
		watchItem.TargetBuyPrice = floatToNullDecimal(buyPrice)
	}
	if sellPrice > 0 {
		watchItem.TargetSellPrice = floatToNullDecimal(sellPrice)
	}

	// Save to database
	err := repo.AddToWatchList(ctx, watchItem)
	if err != nil {
		log.Fatal("Failed to add to watch list:", err)
	}

	fmt.Printf("✅ Added to watch list: %s (%s)\n", name, code)
	if buyPrice > 0 {
		fmt.Printf("   Buy target: ¥%.0f\n", buyPrice)
	}
	if sellPrice > 0 {
		fmt.Printf("   Sell target: ¥%.0f\n", sellPrice)
	}
}

func updateWatchList(ctx context.Context, repo repository.StockRepository, id string, buyPrice, sellPrice float64) {
	// Get existing item
	watchItem, err := repo.GetWatchListItem(ctx, id)
	if err != nil {
		log.Fatal("Failed to get watch list item:", err)
	}
	if watchItem == nil {
		log.Fatal("Watch list item not found")
	}

	// Update prices
	if buyPrice > 0 {
		watchItem.TargetBuyPrice = floatToNullDecimal(buyPrice)
	}
	if sellPrice > 0 {
		watchItem.TargetSellPrice = floatToNullDecimal(sellPrice)
	}

	// Save updates
	err = repo.UpdateWatchList(ctx, watchItem)
	if err != nil {
		log.Fatal("Failed to update watch list:", err)
	}

	fmt.Printf("✅ Updated watch list item: %s (%s)\n", watchItem.Name, watchItem.Code)
	if buyPrice > 0 {
		fmt.Printf("   New buy target: ¥%.0f\n", buyPrice)
	}
	if sellPrice > 0 {
		fmt.Printf("   New sell target: ¥%.0f\n", sellPrice)
	}
}

func deleteFromWatchList(ctx context.Context, repo repository.StockRepository, id string) {
	// Get item details first
	watchItem, err := repo.GetWatchListItem(ctx, id)
	if err != nil {
		log.Fatal("Failed to get watch list item:", err)
	}
	if watchItem == nil {
		log.Fatal("Watch list item not found")
	}

	// Delete the item
	err = repo.DeleteFromWatchList(ctx, id)
	if err != nil {
		log.Fatal("Failed to delete from watch list:", err)
	}

	fmt.Printf("✅ Deleted from watch list: %s (%s)\n", watchItem.Name, watchItem.Code)
}

func toggleWatchList(ctx context.Context, repo repository.StockRepository, id string) {
	// Get existing item
	watchItem, err := repo.GetWatchListItem(ctx, id)
	if err != nil {
		log.Fatal("Failed to get watch list item:", err)
	}
	if watchItem == nil {
		log.Fatal("Watch list item not found")
	}

	// Toggle active status
	watchItem.IsActive = null.BoolFrom(!watchItem.IsActive.Bool)

	// Save updates
	err = repo.UpdateWatchList(ctx, watchItem)
	if err != nil {
		log.Fatal("Failed to update watch list:", err)
	}

	status := "deactivated"
	if watchItem.IsActive.Bool {
		status = "activated"
	}
	fmt.Printf("✅ Watch list item %s: %s (%s)\n", status, watchItem.Name, watchItem.Code)
}

func floatToNullDecimal(value float64) types.NullDecimal {
	d := new(decimal.Big)
	d.SetFloat64(value)
	return types.NullDecimal{Big: d}
}

