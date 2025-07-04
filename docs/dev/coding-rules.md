# コーディングルール

## 基本原則

### 1. 可読性優先
- コードは書く時間より読む時間の方が長い
- 明確で理解しやすいコードを書く
- 適切なコメントを記述する

### 2. 一貫性
- プロジェクト全体で統一されたスタイルを維持
- 既存のコードスタイルに合わせる
- チーム内でのルール遵守

### 3. 単純性
- 複雑な実装より単純で明確な実装を選ぶ
- YAGNI (You Aren't Gonna Need It) 原則
- 過度な抽象化を避ける

## Go言語固有のルール

### 1. 命名規則

#### パッケージ名
```go
// 良い例
package database
package notification
package analysis

// 悪い例
package db_helper
package utils
```

#### 変数・関数名
```go
// 良い例
var stockPrice float64
func GetDailyReport() *Report
func calculateMovingAverage(prices []float64, period int) float64

// 悪い例
var sp float64
func GetReport() *Report
func calc_ma(p []float64, n int) float64
```

#### 定数名
```go
// 良い例
const (
    DefaultTimeout = 30 * time.Second
    MaxRetryCount  = 3
    APIBaseURL     = "https://api.example.com"
)

// 悪い例
const (
    timeout = 30
    retries = 3
    url     = "https://api.example.com"
)
```

### 2. 構造体定義

#### フィールド名とタグ
```go
// 良い例
type Stock struct {
    Symbol      string    `json:"symbol" db:"symbol"`
    CompanyName string    `json:"company_name" db:"company_name"`
    Price       float64   `json:"price" db:"price"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// 悪い例
type Stock struct {
    s string    `json:"s"`
    n string    `json:"n"`
    p float64   `json:"p"`
    u time.Time `json:"u"`
}
```

#### コンストラクタ関数
```go
// 良い例
func NewStock(symbol, companyName string, price float64) *Stock {
    return &Stock{
        Symbol:      symbol,
        CompanyName: companyName,
        Price:       price,
        UpdatedAt:   time.Now(),
    }
}

// 構造体の初期化も可
stock := &Stock{
    Symbol:      "AAPL",
    CompanyName: "Apple Inc.",
    Price:       150.00,
    UpdatedAt:   time.Now(),
}
```

### 3. エラーハンドリング

#### エラーの即座チェック
```go
// 良い例
data, err := fetchStockData(symbol)
if err != nil {
    return fmt.Errorf("株価データの取得に失敗: %w", err)
}

// 悪い例
data, _ := fetchStockData(symbol) // エラー無視
```

#### カスタムエラー型
```go
// 良い例
type APIError struct {
    Code    int
    Message string
    Cause   error
}

func (e *APIError) Error() string {
    return fmt.Sprintf("API Error [%d]: %s", e.Code, e.Message)
}

func (e *APIError) Unwrap() error {
    return e.Cause
}
```

### 4. 並行処理

#### Goroutineの適切な使用
```go
// 良い例
func processStocks(stocks []string) error {
    var wg sync.WaitGroup
    errCh := make(chan error, len(stocks))
    
    for _, stock := range stocks {
        wg.Add(1)
        go func(symbol string) {
            defer wg.Done()
            if err := processStock(symbol); err != nil {
                errCh <- fmt.Errorf("処理失敗 %s: %w", symbol, err)
            }
        }(stock)
    }
    
    wg.Wait()
    close(errCh)
    
    for err := range errCh {
        if err != nil {
            return err
        }
    }
    return nil
}
```

#### コンテキストの使用
```go
// 良い例
func fetchStockData(ctx context.Context, symbol string) (*Stock, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
    if err != nil {
        return nil, err
    }
    
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    // レスポンス処理
    return parseStock(resp.Body)
}
```

## プロジェクト固有のルール

### 1. ディレクトリ構造
```
backend/
├── cmd/                 # エントリーポイント
├── internal/           # 内部パッケージ
│   ├── api/           # API関連
│   ├── database/      # データベース関連
│   ├── models/        # データモデル
│   ├── analysis/      # 分析ロジック
│   └── notification/  # 通知システム
├── pkg/               # 外部公開パッケージ
├── configs/           # 設定ファイル
├── docker/            # Docker関連
└── docs/              # ドキュメント
```

### 2. ログ出力

#### 構造化ログ
```go
// 良い例
log.Info("株価データ取得開始",
    zap.String("symbol", symbol),
    zap.String("date", date.Format("2006-01-02")),
)

log.Error("API呼び出し失敗",
    zap.String("symbol", symbol),
    zap.Error(err),
    zap.Duration("duration", time.Since(start)),
)

// 悪い例
log.Printf("Getting stock data for %s", symbol)
log.Printf("Error: %v", err)
```

### 3. 設定管理

#### 環境変数と設定ファイル
```go
// 良い例
type Config struct {
    Database struct {
        Host     string `yaml:"host" env:"DB_HOST"`
        Port     int    `yaml:"port" env:"DB_PORT"`
        User     string `yaml:"user" env:"DB_USER"`
        Password string `yaml:"password" env:"DB_PASSWORD"`
    } `yaml:"database"`
    
    API struct {
        Timeout time.Duration `yaml:"timeout" env:"API_TIMEOUT"`
        BaseURL string        `yaml:"base_url" env:"API_BASE_URL"`
    } `yaml:"api"`
}
```

### 4. テスト

#### テスト関数名
```go
// 良い例
func TestCalculateMovingAverage(t *testing.T) {}
func TestGetStockData_Success(t *testing.T) {}
func TestGetStockData_APIError(t *testing.T) {}

// 悪い例
func TestMA(t *testing.T) {}
func TestStock(t *testing.T) {}
```

#### テーブルドリブンテスト
```go
// 良い例
func TestCalculateMovingAverage(t *testing.T) {
    tests := []struct {
        name     string
        prices   []float64
        period   int
        expected float64
        hasError bool
    }{
        {
            name:     "正常ケース_5日移動平均",
            prices:   []float64{100, 110, 105, 115, 120},
            period:   5,
            expected: 110.0,
            hasError: false,
        },
        {
            name:     "データ不足エラー",
            prices:   []float64{100, 110},
            period:   5,
            expected: 0,
            hasError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := calculateMovingAverage(tt.prices, tt.period)
            if tt.hasError {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

## コメント規則

### 1. パッケージコメント
```go
// Package analysis provides technical analysis tools for stock data.
// It includes moving averages, RSI, and other technical indicators.
package analysis
```

### 2. 関数コメント
```go
// CalculateMovingAverage calculates the simple moving average for the given period.
// It returns an error if the data length is insufficient for the specified period.
//
// Parameters:
//   - prices: slice of stock prices
//   - period: number of periods for moving average calculation
//
// Returns:
//   - float64: calculated moving average
//   - error: error if calculation fails
func CalculateMovingAverage(prices []float64, period int) (float64, error) {
    // 実装
}
```

### 3. 複雑なロジックのコメント
```go
// Yahoo Finance APIのレート制限を考慮してリクエスト間隔を調整
// 1秒間に最大5回までのリクエストに制限
time.Sleep(200 * time.Millisecond)

// RSI計算: 14日間の上昇・下降の平均を使用
avgGain := calculateAverage(gains[:14])
avgLoss := calculateAverage(losses[:14])
```

## パフォーマンスガイドライン

### 1. メモリ効率
```go
// 良い例: スライスの事前確保
results := make([]Result, 0, len(stocks))

// 良い例: 文字列結合にstrings.Builder使用
var builder strings.Builder
for _, item := range items {
    builder.WriteString(item)
}
result := builder.String()

// 悪い例: 文字列の繰り返し結合
var result string
for _, item := range items {
    result += item // メモリ効率が悪い
}
```

### 2. データベース操作
```go
// 良い例: バッチ処理
func InsertStocks(stocks []Stock) error {
    query := "INSERT INTO stocks (symbol, price, updated_at) VALUES (?, ?, ?)"
    stmt, err := db.Prepare(query)
    if err != nil {
        return err
    }
    defer stmt.Close()
    
    for _, stock := range stocks {
        _, err := stmt.Exec(stock.Symbol, stock.Price, stock.UpdatedAt)
        if err != nil {
            return err
        }
    }
    return nil
}
```

## セキュリティルール

### 1. 機密情報の取り扱い
```go
// 良い例: 環境変数から取得
apiKey := os.Getenv("YAHOO_API_KEY")
if apiKey == "" {
    return errors.New("YAHOO_API_KEY環境変数が設定されていません")
}

// 悪い例: ハードコード
const apiKey = "abc123xyz" // 絶対に禁止
```

### 2. SQLインジェクション対策
```go
// 良い例: プリペアドステートメント使用
query := "SELECT * FROM stocks WHERE symbol = ?"
rows, err := db.Query(query, symbol)

// 悪い例: 文字列結合
query := fmt.Sprintf("SELECT * FROM stocks WHERE symbol = '%s'", symbol)
```

## コードレビューチェックリスト

### 提出前チェック
- [ ] gofmt でフォーマット済み
- [ ] golint でリント済み
- [ ] テストが全て通る
- [ ] 適切なコメントが記述されている
- [ ] エラーハンドリングが適切
- [ ] 機密情報がハードコードされていない

### レビュー観点
- [ ] ロジックの正確性
- [ ] パフォーマンスの問題
- [ ] セキュリティの問題
- [ ] テストカバレッジ
- [ ] ドキュメントの更新