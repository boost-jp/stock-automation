# Slack通知システム

## 概要
Slack Webhook APIを使用した株価アラート・投資レポートの自動通知システム

## Slack Webhook設定

### Step 1: Incoming Webhook作成
1. Slackワークスペースで「App」→「App Directory」
2. 「Incoming WebHooks」を検索・インストール
3. 通知先チャンネル選択（例: #stock-alerts）
4. Webhook URLをコピー保存

### Step 2: 環境変数設定
```bash
# .env ファイルに追加
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX
```

## Go実装

### 基本的な通知システム

#### `internal/notification/slack.go`
```go
package notification

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "time"
    
    "github.com/sirupsen/logrus"
)

type SlackNotifier struct {
    webhookURL string
    client     *http.Client
}

type SlackMessage struct {
    Text        string       `json:"text,omitempty"`
    Username    string       `json:"username,omitempty"`
    Channel     string       `json:"channel,omitempty"`
    IconEmoji   string       `json:"icon_emoji,omitempty"`
    Attachments []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
    Color    string  `json:"color,omitempty"`
    Title    string  `json:"title,omitempty"`
    Text     string  `json:"text,omitempty"`
    Fields   []Field `json:"fields,omitempty"`
    Footer   string  `json:"footer,omitempty"`
    Ts       int64   `json:"ts,omitempty"`
}

type Field struct {
    Title string `json:"title"`
    Value string `json:"value"`
    Short bool   `json:"short"`
}

func NewSlackNotifier() *SlackNotifier {
    return &SlackNotifier{
        webhookURL: os.Getenv("SLACK_WEBHOOK_URL"),
        client: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

func (s *SlackNotifier) SendMessage(text string) error {
    message := SlackMessage{
        Text:      text,
        Username:  "Stock Bot",
        IconEmoji: ":chart_with_upwards_trend:",
    }
    
    return s.sendMessage(message)
}

func (s *SlackNotifier) sendMessage(message SlackMessage) error {
    if s.webhookURL == "" {
        return fmt.Errorf("Slack webhook URL not configured")
    }
    
    jsonData, err := json.Marshal(message)
    if err != nil {
        return fmt.Errorf("failed to marshal message: %w", err)
    }
    
    resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return fmt.Errorf("failed to send message: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }
    
    logrus.Debug("Slack message sent successfully")
    return nil
}
```

### 価格アラート通知

```go
func (s *SlackNotifier) SendPriceAlert(stockCode, stockName string, currentPrice, targetPrice float64, alertType string) error {
    var color, emoji string
    
    switch alertType {
    case "buy":
        color = "good"  // 緑
        emoji = ":arrow_down:"
    case "sell":
        color = "danger"  // 赤
        emoji = ":arrow_up:"
    default:
        color = "warning"  // 黄
        emoji = ":bell:"
    }
    
    message := SlackMessage{
        Username:  "Stock Alert Bot",
        IconEmoji: ":bell:",
        Attachments: []Attachment{
            {
                Color: color,
                Title: fmt.Sprintf("%s %sシグナル", emoji, alertType),
                Fields: []Field{
                    {
                        Title: "銘柄",
                        Value: fmt.Sprintf("%s (%s)", stockName, stockCode),
                        Short: true,
                    },
                    {
                        Title: "現在価格",
                        Value: fmt.Sprintf("¥%,.0f", currentPrice),
                        Short: true,
                    },
                    {
                        Title: "目標価格",
                        Value: fmt.Sprintf("¥%,.0f", targetPrice),
                        Short: true,
                    },
                    {
                        Title: "差額",
                        Value: fmt.Sprintf("¥%,.0f", currentPrice-targetPrice),
                        Short: true,
                    },
                },
                Footer: "Stock Automation System",
                Ts:     time.Now().Unix(),
            },
        },
    }
    
    return s.sendMessage(message)
}
```

### 日次レポート通知

```go
type PortfolioSummary struct {
    TotalValue     float64
    TotalProfit    float64
    TotalProfitRate float64
    Stocks         []StockSummary
}

type StockSummary struct {
    Code         string
    Name         string
    CurrentPrice float64
    Shares       int
    Value        float64
    Profit       float64
    ProfitRate   float64
}

func (s *SlackNotifier) SendDailyReport(summary PortfolioSummary) error {
    var color string
    var emoji string
    
    if summary.TotalProfit > 0 {
        color = "good"
        emoji = ":chart_with_upwards_trend:"
    } else if summary.TotalProfit < 0 {
        color = "danger" 
        emoji = ":chart_with_downwards_trend:"
    } else {
        color = "warning"
        emoji = ":bar_chart:"
    }
    
    // 各銘柄の詳細
    var stockDetails string
    for _, stock := range summary.Stocks {
        profitEmoji := ":small_red_triangle_down:"
        if stock.Profit > 0 {
            profitEmoji = ":small_red_triangle:"
        } else if stock.Profit == 0 {
            profitEmoji = ":small_blue_diamond:"
        }
        
        stockDetails += fmt.Sprintf("%s *%s*: ¥%,.0f (%+.1f%%)\n", 
            profitEmoji, stock.Name, stock.CurrentPrice, stock.ProfitRate*100)
    }
    
    message := SlackMessage{
        Username:  "Portfolio Report Bot",
        IconEmoji: ":bar_chart:",
        Attachments: []Attachment{
            {
                Color: color,
                Title: fmt.Sprintf("%s 本日のポートフォリオレポート", emoji),
                Text:  stockDetails,
                Fields: []Field{
                    {
                        Title: "総評価額",
                        Value: fmt.Sprintf("¥%,.0f", summary.TotalValue),
                        Short: true,
                    },
                    {
                        Title: "総損益",
                        Value: fmt.Sprintf("%+.0f円", summary.TotalProfit),
                        Short: true,
                    },
                    {
                        Title: "総損益率",
                        Value: fmt.Sprintf("%+.2f%%", summary.TotalProfitRate*100),
                        Short: true,
                    },
                    {
                        Title: "銘柄数",
                        Value: fmt.Sprintf("%d銘柄", len(summary.Stocks)),
                        Short: true,
                    },
                },
                Footer: fmt.Sprintf("更新日時: %s", time.Now().Format("2006/01/02 15:04")),
                Ts:     time.Now().Unix(),
            },
        },
    }
    
    return s.sendMessage(message)
}
```

### 投資判断アラート

```go
type InvestmentAnalysis struct {
    StockCode      string
    StockName      string
    Recommendation string
    Confidence     int
    BuySignals     []string
    SellSignals    []string
    CurrentPrice   float64
    TargetPrice    float64
}

func (s *SlackNotifier) SendInvestmentAlert(analysis InvestmentAnalysis) error {
    var color, emoji string
    
    switch analysis.Recommendation {
    case "強い買い":
        color = "good"
        emoji = ":rocket:"
    case "買い":
        color = "good"
        emoji = ":arrow_up:"
    case "強い売り":
        color = "danger"
        emoji = ":exclamation:"
    case "売り":
        color = "warning"
        emoji = ":arrow_down:"
    default:
        color = "#439FE0"
        emoji = ":information_source:"
    }
    
    // シグナルの整理
    allSignals := append(analysis.BuySignals, analysis.SellSignals...)
    signalsText := ""
    for _, signal := range allSignals {
        signalsText += fmt.Sprintf("• %s\n", signal)
    }
    
    message := SlackMessage{
        Username:  "Investment Analysis Bot",
        IconEmoji: ":brain:",
        Attachments: []Attachment{
            {
                Color: color,
                Title: fmt.Sprintf("%s 投資判断アラート", emoji),
                Fields: []Field{
                    {
                        Title: "銘柄",
                        Value: fmt.Sprintf("%s (%s)", analysis.StockName, analysis.StockCode),
                        Short: true,
                    },
                    {
                        Title: "現在価格",
                        Value: fmt.Sprintf("¥%,.0f", analysis.CurrentPrice),
                        Short: true,
                    },
                    {
                        Title: "判断",
                        Value: analysis.Recommendation,
                        Short: true,
                    },
                    {
                        Title: "信頼度",
                        Value: fmt.Sprintf("%d%%", analysis.Confidence),
                        Short: true,
                    },
                    {
                        Title: "根拠",
                        Value: signalsText,
                        Short: false,
                    },
                },
                Footer: "AI Investment Analysis",
                Ts:     time.Now().Unix(),
            },
        },
    }
    
    return s.sendMessage(message)
}
```

### システムアラート

```go
func (s *SlackNotifier) SendSystemAlert(alertType, message string, err error) error {
    var color, emoji string
    
    switch alertType {
    case "ERROR":
        color = "danger"
        emoji = ":x:"
    case "WARNING":
        color = "warning"
        emoji = ":warning:"
    case "INFO":
        color = "good"
        emoji = ":information_source:"
    default:
        color = "#439FE0"
        emoji = ":gear:"
    }
    
    var errorText string
    if err != nil {
        errorText = fmt.Sprintf("\nエラー詳細: `%s`", err.Error())
    }
    
    slackMessage := SlackMessage{
        Username:  "System Monitor",
        IconEmoji: ":robot_face:",
        Attachments: []Attachment{
            {
                Color: color,
                Title: fmt.Sprintf("%s システム通知", emoji),
                Text:  fmt.Sprintf("%s%s", message, errorText),
                Fields: []Field{
                    {
                        Title: "種類",
                        Value: alertType,
                        Short: true,
                    },
                    {
                        Title: "時刻",
                        Value: time.Now().Format("2006/01/02 15:04:05"),
                        Short: true,
                    },
                },
                Footer: "Stock Automation System",
                Ts:     time.Now().Unix(),
            },
        },
    }
    
    return s.sendMessage(slackMessage)
}
```

## 通知管理システム

### 通知頻度制限

```go
type NotificationManager struct {
    notifier     *SlackNotifier
    lastSent     map[string]time.Time
    cooldownTime time.Duration
    mu           sync.RWMutex
}

func NewNotificationManager(notifier *SlackNotifier) *NotificationManager {
    return &NotificationManager{
        notifier:     notifier,
        lastSent:     make(map[string]time.Time),
        cooldownTime: 5 * time.Minute, // 5分のクールダウン
    }
}

func (nm *NotificationManager) SendPriceAlertWithCooldown(stockCode string, alert PriceAlert) error {
    nm.mu.Lock()
    defer nm.mu.Unlock()
    
    key := fmt.Sprintf("%s_%s", stockCode, alert.Type)
    
    if lastSent, exists := nm.lastSent[key]; exists {
        if time.Since(lastSent) < nm.cooldownTime {
            logrus.Debug("Skipping notification due to cooldown:", key)
            return nil
        }
    }
    
    err := nm.notifier.SendPriceAlert(stockCode, alert.Name, alert.CurrentPrice, alert.TargetPrice, alert.Type)
    if err == nil {
        nm.lastSent[key] = time.Now()
    }
    
    return err
}
```

### バッチ通知

```go
func (s *SlackNotifier) SendBatchAlerts(alerts []PriceAlert) error {
    if len(alerts) == 0 {
        return nil
    }
    
    var buyAlerts, sellAlerts []PriceAlert
    for _, alert := range alerts {
        if alert.Type == "buy" {
            buyAlerts = append(buyAlerts, alert)
        } else {
            sellAlerts = append(sellAlerts, alert)
        }
    }
    
    var fields []Field
    
    if len(buyAlerts) > 0 {
        var buyText string
        for _, alert := range buyAlerts {
            buyText += fmt.Sprintf("• %s: ¥%,.0f\n", alert.Name, alert.CurrentPrice)
        }
        fields = append(fields, Field{
            Title: ":arrow_down: 買いシグナル",
            Value: buyText,
            Short: true,
        })
    }
    
    if len(sellAlerts) > 0 {
        var sellText string
        for _, alert := range sellAlerts {
            sellText += fmt.Sprintf("• %s: ¥%,.0f\n", alert.Name, alert.CurrentPrice)
        }
        fields = append(fields, Field{
            Title: ":arrow_up: 売りシグナル",
            Value: sellText,
            Short: true,
        })
    }
    
    message := SlackMessage{
        Username:  "Batch Alert Bot",
        IconEmoji: ":bell:",
        Attachments: []Attachment{
            {
                Color:  "warning",
                Title:  fmt.Sprintf(":bell: %d件のアラートが発生しました", len(alerts)),
                Fields: fields,
                Footer: fmt.Sprintf("発生時刻: %s", time.Now().Format("15:04")),
                Ts:     time.Now().Unix(),
            },
        },
    }
    
    return s.sendMessage(message)
}
```

## テスト用機能

### 通知テスト

```go
func (s *SlackNotifier) TestNotifications() error {
    // 基本通知テスト
    if err := s.SendMessage("📊 Slack通知システムのテストです"); err != nil {
        return fmt.Errorf("basic notification test failed: %w", err)
    }
    
    time.Sleep(1 * time.Second)
    
    // アラートテスト
    if err := s.SendPriceAlert("7203", "トヨタ自動車", 2650, 2600, "buy"); err != nil {
        return fmt.Errorf("price alert test failed: %w", err)
    }
    
    time.Sleep(1 * time.Second)
    
    // システムアラートテスト
    if err := s.SendSystemAlert("INFO", "システム通知のテストです", nil); err != nil {
        return fmt.Errorf("system alert test failed: %w", err)
    }
    
    return nil
}
```

### 設定検証

```go
func (s *SlackNotifier) ValidateConfiguration() error {
    if s.webhookURL == "" {
        return fmt.Errorf("Slack webhook URL is not configured")
    }
    
    // 接続テスト
    testMessage := SlackMessage{
        Text:      "Configuration validation test",
        Username:  "Config Test",
        IconEmoji: ":test_tube:",
    }
    
    if err := s.sendMessage(testMessage); err != nil {
        return fmt.Errorf("failed to validate Slack configuration: %w", err)
    }
    
    return nil
}
```

## 使用例

### メイン関数での初期化

```go
func main() {
    // 通知システム初期化
    notifier := notification.NewSlackNotifier()
    
    // 設定検証
    if err := notifier.ValidateConfiguration(); err != nil {
        logrus.Fatal("Slack configuration error:", err)
    }
    
    // 通知マネージャー初期化
    notificationManager := notification.NewNotificationManager(notifier)
    
    // 起動通知
    notifier.SendSystemAlert("INFO", "Stock Automation System Started", nil)
    
    // ... メインロジック
}
```

### 定期的な通知

```go
func scheduledReports(notifier *notification.SlackNotifier) {
    // 毎日18:00にレポート送信
    c := cron.New()
    c.AddFunc("0 18 * * *", func() {
        summary := generatePortfolioSummary()
        notifier.SendDailyReport(summary)
    })
    c.Start()
}
```

この実装により、Slackを通じて効果的な株価監視・投資判断の通知システムが構築できます。次は[データ収集システム](data-collection.md)の実装に進んでください。