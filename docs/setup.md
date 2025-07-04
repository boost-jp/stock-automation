# 環境構築ガイド

## 概要
Go+自宅PC+Slack版の株式投資自動化システムの環境構築手順

## 前提条件

### システム要件
- **OS**: Windows 10/11, macOS 10.15+, Linux (Ubuntu 20.04+)
- **CPU**: 2コア以上
- **メモリ**: 4GB以上
- **ストレージ**: 10GB以上の空き容量
- **ネットワーク**: 常時インターネット接続

## Step 1: Go開発環境のセットアップ

### 1.1 Goのインストール

#### Windows
```bash
# 1. https://golang.org/dl/ からWindows用インストーラーをダウンロード
# 2. msiファイルを実行してインストール
# 3. コマンドプロンプトで確認
go version
```

#### macOS
```bash
# Homebrewを使用（推奨）
brew install go

# または公式インストーラー
# https://golang.org/dl/ からpkgファイルをダウンロード・実行

# 確認
go version
```

#### Linux (Ubuntu)
```bash
# 公式バイナリのダウンロード・インストール
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# パスの設定
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# 確認
go version
```

### 1.2 Go環境の設定

```bash
# GOPATH・GOROOT確認
go env GOPATH
go env GOROOT

# プロキシ設定（必要に応じて）
go env -w GOPROXY=https://proxy.golang.org,direct
go env -w GOSUMDB=sum.golang.org
```

## Step 2: プロジェクトセットアップ

### 2.1 プロジェクト構造の作成

```bash
# プロジェクトディレクトリ作成
mkdir stock-automation
cd stock-automation

# Goモジュール初期化
go mod init stock-automation

# ディレクトリ構造作成
mkdir -p cmd
mkdir -p internal/{api,database,analysis,notification,models}
mkdir -p configs
mkdir -p data
mkdir -p logs
mkdir -p scripts

# 基本ファイル作成
touch cmd/main.go
touch configs/config.yaml
touch configs/stocks.yaml
touch scripts/setup.sh
touch scripts/deploy.sh
touch README.md
```

### 2.2 必要ライブラリのインストール

```bash
# HTTPクライアント・JSON処理
go get github.com/go-resty/resty/v2

# MySQLドライバー
go get github.com/go-sql-driver/mysql

# データベースORM
go get gorm.io/gorm
go get gorm.io/driver/mysql

# 設定ファイル読み込み
go get gopkg.in/yaml.v3

# ログライブラリ
go get github.com/sirupsen/logrus

# スケジューラー
go get github.com/go-co-op/gocron

# 環境変数管理
go get github.com/joho/godotenv

# HTTP router（WebUI用、オプション）
go get github.com/gin-gonic/gin

# Redis クライアント（キャッシュ用）
go get github.com/go-redis/redis/v8
```

### 2.3 設定ファイルの作成

#### `configs/config.yaml`
```yaml
# アプリケーション設定
app:
  name: "Stock Automation"
  version: "1.0.0"
  log_level: "info"
  log_file: "logs/app.log"

# データベース設定
database:
  host: "localhost"
  port: 3306
  user: "stock_user"
  password: "stock_password_456"
  database: "stock_automation"
  charset: "utf8mb4"
  parse_time: true
  loc: "Asia%2FTokyo"
  
# API設定
apis:
  yahoo_finance:
    base_url: "https://query1.finance.yahoo.com"
    timeout: 30
    retry_count: 3
    
# Slack設定
slack:
  webhook_url: "${SLACK_WEBHOOK_URL}"
  channel: "#stock-alerts"
  username: "Stock Bot"
  
# スケジューラー設定
scheduler:
  data_update_interval: "5m"  # 5分毎
  analysis_interval: "15m"    # 15分毎
  daily_report_time: "18:00"  # 18時
  
# 投資設定
investment:
  max_position_size: 0.05     # 総資産の5%
  stop_loss_rate: 0.10        # 10%で損切り
  take_profit_rate: 0.20      # 20%で利確
```

#### `configs/stocks.yaml`
```yaml
# 監視銘柄リスト
watch_list:
  - code: "7203"
    name: "トヨタ自動車"
    target_buy_price: 2500
    target_sell_price: 2800
    
  - code: "6758"
    name: "ソニーグループ"
    target_buy_price: 12000
    target_sell_price: 15000
    
  - code: "9984"
    name: "ソフトバンクグループ"
    target_buy_price: 5000
    target_sell_price: 6000

# ポートフォリオ（保有銘柄）
portfolio:
  - code: "7203"
    name: "トヨタ自動車"
    shares: 100
    purchase_price: 2550
    purchase_date: "2024-01-15"
```

## Step 3: Slack設定

### 3.1 Slackワークスペースの作成（既存がない場合）
1. https://slack.com/create にアクセス
2. 「ワークスペースを作成する」をクリック
3. メールアドレスを入力・認証
4. ワークスペース名・チャンネル名を設定

### 3.2 Incoming Webhookの設定
1. Slackワークスペースにログイン
2. 左下の「App」→「App Directory」
3. 「Incoming WebHooks」を検索・追加
4. 通知先チャンネルを選択（例: #stock-alerts）
5. Webhook URLをコピー（後で使用）

### 3.3 環境変数の設定

#### `.env`ファイル作成
```bash
# プロジェクトルートに.envファイル作成
touch .env

# 以下を追加
echo "SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL" >> .env
```

#### システム環境変数設定（オプション）

**Windows:**
```cmd
setx SLACK_WEBHOOK_URL "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
```

**macOS/Linux:**
```bash
# ~/.bashrc または ~/.zshrc に追加
echo 'export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"' >> ~/.bashrc
source ~/.bashrc
```

## Step 4: 基本コードの実装

### 4.1 メインプログラム (`cmd/main.go`)
```go
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "stock-automation/internal/database"
    "stock-automation/internal/notification"
    
    "github.com/joho/godotenv"
    "github.com/sirupsen/logrus"
)

func main() {
    // 環境変数読み込み
    if err := godotenv.Load(); err != nil {
        log.Println("Warning: .env file not found")
    }
    
    // ログ設定
    logrus.SetLevel(logrus.InfoLevel)
    logrus.SetFormatter(&logrus.JSONFormatter{})
    
    // データベース接続
    dbConfig := database.Config{
        Host:      "localhost",
        Port:      3306,
        User:      "stock_user",
        Password:  "stock_password_456",
        Database:  "stock_automation",
        Charset:   "utf8mb4",
        ParseTime: true,
        Loc:       "Asia%2FTokyo",
    }
    
    db, err := database.NewConnection(dbConfig)
    if err != nil {
        logrus.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()
    
    // 通知システム初期化
    notifier := notification.NewSlackNotifier()
    
    // テスト通知
    if err := notifier.SendMessage("🚀 Stock Automation System Started!"); err != nil {
        logrus.Error("Failed to send startup notification:", err)
    }
    
    logrus.Info("Stock Automation System is running...")
    
    // Graceful shutdown
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    <-c
    
    logrus.Info("Stock Automation System is shutting down...")
}
```

### 4.2 設定管理 (`internal/models/config.go`)
```go
package models

import (
    "os"
    "gopkg.in/yaml.v3"
)

type Config struct {
    App        AppConfig        `yaml:"app"`
    Database   DatabaseConfig   `yaml:"database"`
    APIs       APIConfig        `yaml:"apis"`
    Slack      SlackConfig      `yaml:"slack"`
    Scheduler  SchedulerConfig  `yaml:"scheduler"`
    Investment InvestmentConfig `yaml:"investment"`
}

type AppConfig struct {
    Name     string `yaml:"name"`
    Version  string `yaml:"version"`
    LogLevel string `yaml:"log_level"`
    LogFile  string `yaml:"log_file"`
}

type SlackConfig struct {
    WebhookURL string `yaml:"webhook_url"`
    Channel    string `yaml:"channel"`
    Username   string `yaml:"username"`
}

// 設定ファイル読み込み
func LoadConfig(path string) (*Config, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    
    var config Config
    decoder := yaml.NewDecoder(file)
    if err := decoder.Decode(&config); err != nil {
        return nil, err
    }
    
    // 環境変数の展開
    config.Slack.WebhookURL = os.ExpandEnv(config.Slack.WebhookURL)
    
    return &config, nil
}
```

## Step 5: 動作確認

### 5.1 基本テスト
```bash
# 依存関係の確認
go mod tidy

# ビルドテスト
go build -o stock-automation cmd/main.go

# 実行テスト
./stock-automation
```

### 5.2 Slack通知テスト
```go
// test/notification_test.go
package main

import (
    "stock-automation/internal/notification"
    "github.com/joho/godotenv"
)

func main() {
    godotenv.Load()
    
    notifier := notification.NewSlackNotifier()
    err := notifier.SendMessage("📊 Test notification from Go application!")
    
    if err != nil {
        panic(err)
    }
    
    println("Notification sent successfully!")
}
```

## Step 6: 自動起動設定

### 6.1 systemd設定（Linux）
```bash
# サービスファイル作成
sudo vim /etc/systemd/system/stock-automation.service
```

```ini
[Unit]
Description=Stock Automation Service
After=network.target

[Service]
Type=simple
User=your-username
WorkingDirectory=/path/to/stock-automation
ExecStart=/path/to/stock-automation/stock-automation
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
# サービス有効化・開始
sudo systemctl daemon-reload
sudo systemctl enable stock-automation
sudo systemctl start stock-automation
sudo systemctl status stock-automation
```

### 6.2 Windows Serviceとして登録
```bash
# nssm (Non-Sucking Service Manager) を使用
# https://nssm.cc/download からダウンロード

nssm install StockAutomation
# パス設定: C:\path\to\stock-automation.exe
# 作業ディレクトリ: C:\path\to\project\

nssm start StockAutomation
```

### 6.3 macOS LaunchDaemon
```xml
<!-- ~/Library/LaunchAgents/com.stockautomation.plist -->
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.stockautomation</string>
    <key>ProgramArguments</key>
    <array>
        <string>/path/to/stock-automation</string>
    </array>
    <key>WorkingDirectory</key>
    <string>/path/to/project</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
```

```bash
# 登録・開始
launchctl load ~/Library/LaunchAgents/com.stockautomation.plist
launchctl start com.stockautomation
```

## Step 7: 開発環境の最適化

### 7.1 VSCode設定
```bash
# Go拡張機能インストール
code --install-extension golang.go

# settings.json設定
mkdir .vscode
cat << EOF > .vscode/settings.json
{
    "go.lintTool": "golangci-lint",
    "go.formatTool": "goimports",
    "go.useLanguageServer": true,
    "[go]": {
        "editor.formatOnSave": true,
        "editor.codeActionsOnSave": {
            "source.organizeImports": true
        }
    }
}
EOF
```

### 7.2 Git設定
```bash
# .gitignore作成
cat << EOF > .gitignore
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
stock-automation

# Test binary
*.test

# Output of the go coverage tool
*.out

# Dependency directories
vendor/

# Go workspace file
go.work

# Environment variables
.env

# Data files
data/*.db
logs/*.log

# IDE
.vscode/
.idea/
EOF

# 初期コミット
git init
git add .
git commit -m "Initial commit: Project setup"
```

## トラブルシューティング

### 一般的な問題と解決策

#### 1. Go Moduleエラー
```bash
# 依存関係を整理
go mod tidy
go mod download
```

#### 2. SQLiteビルドエラー（Windows）
```bash
# CGOを有効にしてGCCをインストール
# TDM-GCC をインストールするか
# go get github.com/mattn/go-sqlite3 の代わりに
go get modernc.org/sqlite
```

#### 3. 権限エラー（Linux）
```bash
# 実行権限付与
chmod +x stock-automation
chmod +x scripts/*.sh
```

#### 4. ファイアウォール問題
- 外部API（Yahoo Finance）への接続許可
- Slackへのアウトバウンド通信許可

これで基本的な環境構築が完了です。次は[データベース設計](database.md)に進んでください。