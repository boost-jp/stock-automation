package interfaces

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

// CLI represents the command line interface for the application
type CLI struct {
	container *Container
}

// NewCLI creates a new CLI instance
func NewCLI(container *Container) *CLI {
	return &CLI{
		container: container,
	}
}

// Run starts the CLI application
func (c *CLI) Run(args []string) error {
	if len(args) < 2 {
		return c.runScheduler()
	}

	command := args[1]
	switch command {
	case "scheduler", "run":
		return c.runScheduler()
	case "collect":
		return c.runDataCollection()
	case "report":
		return c.runDailyReport()
	case "portfolio":
		if len(args) < 3 {
			return fmt.Errorf("portfolio command requires subcommand: add, list, remove")
		}
		return c.runPortfolioCommand(args[2:])
	case "watchlist":
		if len(args) < 3 {
			return fmt.Errorf("watchlist command requires subcommand: add, list, remove")
		}
		return c.runWatchlistCommand(args[2:])
	case "help":
		c.printHelp()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// runScheduler starts the scheduler and waits for shutdown signal
func (c *CLI) runScheduler() error {
	logrus.Info("Starting stock automation scheduler...")

	// Start scheduler
	scheduler := c.container.GetScheduler()
	scheduler.StartScheduledCollection()

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	logrus.Info("Shutting down scheduler...")
	scheduler.Stop()

	return nil
}

// runDataCollection runs immediate data collection
func (c *CLI) runDataCollection() error {
	ctx := context.Background()
	useCase := c.container.GetCollectDataUseCase()

	logrus.Info("Running data collection...")

	// Update all data
	if err := useCase.UpdateAllPrices(ctx); err != nil {
		return fmt.Errorf("failed to update prices: %w", err)
	}

	if err := useCase.UpdateWatchList(ctx); err != nil {
		return fmt.Errorf("failed to update watch list: %w", err)
	}

	if err := useCase.UpdatePortfolio(ctx); err != nil {
		return fmt.Errorf("failed to update portfolio: %w", err)
	}

	logrus.Info("Data collection completed")
	return nil
}

// runDailyReport generates and sends the daily report immediately
func (c *CLI) runDailyReport() error {
	ctx := context.Background()
	useCase := c.container.GetPortfolioReportUseCase()

	logrus.Info("Generating daily report...")

	if err := useCase.GenerateAndSendDailyReport(ctx); err != nil {
		return fmt.Errorf("failed to generate daily report: %w", err)
	}

	logrus.Info("Daily report sent successfully")
	return nil
}

// runPortfolioCommand handles portfolio-related commands
func (c *CLI) runPortfolioCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("portfolio command requires subcommand: add, list, remove")
	}

	ctx := context.Background()
	subcommand := args[0]

	switch subcommand {
	case "add":
		if len(args) < 5 {
			return fmt.Errorf("usage: portfolio add <code> <name> <shares> <price>")
		}
		// TODO: Implement portfolio add functionality
		return fmt.Errorf("portfolio add not implemented yet")

	case "list":
		// Get portfolio statistics
		reportUseCase := c.container.GetPortfolioReportUseCase()
		summary, err := reportUseCase.GetPortfolioStatistics(ctx)
		if err != nil {
			return fmt.Errorf("failed to get portfolio statistics: %w", err)
		}

		// Display portfolio summary
		fmt.Printf("\n📊 Portfolio Summary\n")
		fmt.Printf("==================\n")
		fmt.Printf("Total Value:  ¥%.2f\n", summary.TotalValue)
		fmt.Printf("Total Cost:   ¥%.2f\n", summary.TotalCost)
		fmt.Printf("Total Gain:   ¥%.2f (%.2f%%)\n", summary.TotalGain, summary.TotalGainPercent)

		if len(summary.Holdings) > 0 {
			fmt.Printf("\n📈 Holdings\n")
			fmt.Printf("==================\n")
			for _, holding := range summary.Holdings {
				fmt.Printf("\n%s (%s)\n", holding.Name, holding.Code)
				fmt.Printf("  Shares:       %d\n", holding.Shares)
				fmt.Printf("  Price:        ¥%.2f\n", holding.CurrentPrice)
				fmt.Printf("  Value:        ¥%.2f\n", holding.CurrentValue)
				fmt.Printf("  Gain:         ¥%.2f (%.2f%%)\n", holding.Gain, holding.GainPercent)
			}
		}
		return nil

	case "remove":
		if len(args) < 2 {
			return fmt.Errorf("usage: portfolio remove <code>")
		}
		// TODO: Implement portfolio remove functionality
		return fmt.Errorf("portfolio remove not implemented yet")

	default:
		return fmt.Errorf("unknown portfolio subcommand: %s", subcommand)
	}
}

// runWatchlistCommand handles watchlist-related commands
func (c *CLI) runWatchlistCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("watchlist command requires subcommand: add, list, remove")
	}

	subcommand := args[0]

	switch subcommand {
	case "add":
		if len(args) < 3 {
			return fmt.Errorf("usage: watchlist add <code> <name>")
		}
		// TODO: Implement watchlist add functionality
		return fmt.Errorf("watchlist add not implemented yet")

	case "list":
		// TODO: Implement watchlist list functionality
		return fmt.Errorf("watchlist list not implemented yet")

	case "remove":
		if len(args) < 2 {
			return fmt.Errorf("usage: watchlist remove <code>")
		}
		// TODO: Implement watchlist remove functionality
		return fmt.Errorf("watchlist remove not implemented yet")

	default:
		return fmt.Errorf("unknown watchlist subcommand: %s", subcommand)
	}
}

// printHelp displays the help message
func (c *CLI) printHelp() {
	fmt.Println(`Stock Automation CLI

Usage:
  stock-automation [command]

Commands:
  scheduler, run    Start the scheduler (default)
  collect          Run immediate data collection
  report           Generate and send daily report
  portfolio        Manage portfolio
    add            Add a stock to portfolio
    list           List portfolio holdings
    remove         Remove a stock from portfolio
  watchlist        Manage watchlist
    add            Add a stock to watchlist
    list           List watchlist items
    remove         Remove a stock from watchlist
  help             Show this help message

Examples:
  stock-automation                                   # Start scheduler
  stock-automation collect                           # Run data collection
  stock-automation report                            # Send daily report
  stock-automation portfolio list                    # Show portfolio
  stock-automation portfolio add 7203 Toyota 100 2000  # Add to portfolio
  stock-automation watchlist add 9983 FastRetailing    # Add to watchlist`)
}