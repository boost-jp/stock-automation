package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/boost-jp/stock-automation/app/infrastructure/config"
	"github.com/boost-jp/stock-automation/app/interfaces"
	"github.com/sirupsen/logrus"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// グローバルフラグ
	var (
		showVersion = flag.Bool("version", false, "Show version information")
		logLevel    = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		configPath  = flag.String("config", "configs/config.yaml", "Path to configuration file")
	)

	flag.Parse()

	// バージョン表示
	if *showVersion {
		fmt.Printf("Stock Automation System\n")
		fmt.Printf("  Version: %s\n", version)
		fmt.Printf("  Commit: %s\n", commit)
		fmt.Printf("  Built: %s\n", date)
		os.Exit(0)
	}

	// ログレベル設定
	level, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("Invalid log level: %s", *logLevel)
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   false,
	})

	// 設定ファイル読み込み
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 依存性注入コンテナの初期化
	container, err := interfaces.NewContainer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}
	defer container.Close()

	// Alert service status
	alertService := container.GetAlertService()
	if alertService != nil {
		logrus.Info("Alert service initialized successfully")
	} else {
		logrus.Warn("Alert service not available")
	}

	// CLIインターフェースの実行
	cli := interfaces.NewCLI(container)
	args := append([]string{os.Args[0]}, flag.Args()...)

	if err := cli.Run(args); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
