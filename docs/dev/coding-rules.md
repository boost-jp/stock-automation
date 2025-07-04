# コーディングルール

## 📋 基本原則

### 1. 可読性優先
- コードは書く時間より読む時間の方が長い
- 明確で理解しやすいコードを書く
- 適切なコメントを記述する

### 2. 軽量アーキテクチャ準拠
- オニオンアーキテクチャ + CQRSパターンに従う
- 各層の責任を明確に分離する
- 依存性の逆転原則を守る

### 3. テスト容易性
- テストしやすい設計を心がける
- 依存性注入を活用する
- 純粋関数を優先する

## 🏗️ アーキテクチャルール

### ディレクトリ構造
```
backend/
├── internal/
│   ├── domain/          # ドメイン層（ビジネスロジック）
│   ├── usecase/         # ユースケース層（アプリケーションロジック）
│   ├── repository/      # リポジトリ層（データアクセス）
│   ├── infra/           # インフラ層
│   └── interfaces/      # インターフェース層
├── pkg/                 # 共通パッケージ
└── tests/               # テスト
```

### 層間依存ルール
```go
// ✅ 良い例: 内側から外側への依存なし
// domain層 → 外部依存なし
func (s *PortfolioService) CalculatePerformance(portfolio Portfolio, prices map[string]float64) PortfolioSummary {
    // 純粋なビジネスロジック
}

// usecase層 → domain層、repository層のみ
func (u *PortfolioReportUsecase) GenerateAndSendDailyReport(ctx context.Context) error {
    portfolio, err := u.portfolioRepo.GetPortfolio(ctx)
    if err != nil {
        return err
    }
    summary := u.portfolioSvc.CalculatePerformance(portfolio, prices)
    return nil
}

// ❌ 悪い例: 内側から外側への依存
func (s *PortfolioService) CalculatePerformance(db *gorm.DB) PortfolioSummary {
    // domain層がinfra層に依存してはいけない
}
```

## 🔄 ユースケース実装ルール

### ユースケースの構造
```go
// ✅ 良い例: ユースケースの依存性注入
type PortfolioReportUsecase struct {
    stockRepo       StockRepository
    portfolioRepo   PortfolioRepository
    notificationSvc NotificationService
    portfolioSvc    *domain.PortfolioService
}

func NewPortfolioReportUsecase(
    stockRepo StockRepository,
    portfolioRepo PortfolioRepository,
    notificationSvc NotificationService,
) *PortfolioReportUsecase {
    return &PortfolioReportUsecase{
        stockRepo:       stockRepo,
        portfolioRepo:   portfolioRepo,
        notificationSvc: notificationSvc,
        portfolioSvc:    &domain.PortfolioService{},
    }
}
```

### ユースケースの実装パターン
```go
// ✅ 良い例: シンプルなユースケースフロー
func (u *PortfolioReportUsecase) GenerateAndSendDailyReport(ctx context.Context) error {
    // 1. データ取得（集約単位）
    portfolio, err := u.portfolioRepo.GetPortfolio(ctx)
    if err != nil {
        return err
    }
    
    // 2. 関連データ取得
    codes := extractStockCodes(portfolio)
    prices, err := u.stockRepo.GetLatestPrices(ctx, codes)
    if err != nil {
        return err
    }
    
    // 3. ドメインロジック実行
    summary := u.portfolioSvc.CalculatePerformance(portfolio, prices)
    report := u.portfolioSvc.GenerateReport(summary)
    
    // 4. 外部サービス呼び出し
    return u.notificationSvc.SendReport(ctx, report)
}
```

## 🚨 エラーハンドリングルール

### 基本方針
- **エラーの発生源**: カスタムエラーでメッセージをWrap
- **中間層**: エラーをそのままバブリング
- **必要に応じて**: 追加コンテキストをWrap

### 実装例
```go
// ✅ 良い例: ドメイン層でのエラー生成
func (s *PortfolioService) ValidateHolding(holding Holding) error {
    if holding.Shares <= 0 {
        return errors.NewInvalidArgument("shares must be positive")
    }
    return nil
}

// ✅ 良い例: アプリケーション層でのバブリング
func (h *CommandHandler) UpdatePortfolio(ctx context.Context, cmd UpdatePortfolioCommand) error {
    for _, holding := range cmd.Holdings {
        if err := h.portfolioSvc.ValidateHolding(holding); err != nil {
            return err // そのままバブリング
        }
    }
    return nil
}

// ✅ 良い例: 追加コンテキストが必要な場合
func (h *CommandHandler) CollectStockData(ctx context.Context, cmd CollectStockDataCommand) error {
    prices, err := h.externalSvc.FetchStockPrices(ctx, cmd.StockCodes, cmd.Days)
    if err != nil {
        return errors.Wrap(err, "failed to fetch stock prices from external API")
    }
    return nil
}
```

## 🧪 テストルール

### テスト方針
- **ドメインロジック**: 純粋関数のユニットテスト
- **アプリケーション層**: リアルDBを使った統合テスト
- **モックは最小限**: 外部API呼び出しのみモック

### テスト実装例
```go
// ✅ 良い例: ドメインロジックのユニットテスト
func TestPortfolioService_CalculatePerformance(t *testing.T) {
    tests := []struct {
        name      string
        portfolio domain.Portfolio
        prices    map[string]float64
        expected  float64
    }{
        {
            name: "利益が出ている場合",
            portfolio: domain.Portfolio{
                Holdings: []domain.Holding{
                    {StockCode: "7203", Shares: 100, PurchasePrice: 2800},
                },
            },
            prices:   map[string]float64{"7203": 3000},
            expected: 20000,
        },
    }

    service := &domain.PortfolioService{}
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            summary := service.CalculatePerformance(tt.portfolio, tt.prices)
            assert.Equal(t, tt.expected, summary.TotalGain)
        })
    }
}

// ✅ 良い例: 統合テスト（リアルDB使用）
func TestQueryHandler_GetPortfolioReport(t *testing.T) {
    db := testutil.SetupTestDB(t)
    defer testutil.CleanupTestDB(t, db)
    
    testutil.SeedTestData(t, db)
    
    stockRepo := infra.NewGormStockRepository(db)
    handler := app.NewQueryHandler(stockRepo)
    
    result, err := handler.GetPortfolioReport(context.Background(), query)
    
    assert.NoError(t, err)
    assert.NotEmpty(t, result.Report)
}
```

## 🏷️ 命名規則

### パッケージ名
```go
// ✅ 良い例
package domain
package app
package infra
package interfaces

// ❌ 悪い例
package utils
package helpers
package common
```

### 構造体・インターフェース名
```go
// ✅ 良い例: ドメインエンティティ
type Stock struct {
    Code   string
    Name   string
    Market string
}

type Portfolio struct {
    Holdings []Holding
}

// ✅ 良い例: サービスインターフェース
type StockRepository interface {
    Save(ctx context.Context, stock Stock) error
    FindByCode(ctx context.Context, code string) (*Stock, error)
}

type ExternalDataService interface {
    FetchStockPrices(ctx context.Context, codes []string, days int) ([]StockPrice, error)
}
```

### 関数名
```go
// ✅ 良い例: ドメインサービス
func (s *PortfolioService) CalculatePerformance(portfolio Portfolio, prices map[string]float64) PortfolioSummary

// ✅ 良い例: ハンドラーメソッド
func (h *CommandHandler) CollectStockData(ctx context.Context, cmd CollectStockDataCommand) error
func (h *QueryHandler) GetPortfolioReport(ctx context.Context, query GetPortfolioReportQuery) (*PortfolioReportResponse, error)

// ✅ 良い例: リポジトリメソッド
func (r *GormStockRepository) Save(ctx context.Context, stock domain.Stock) error
func (r *GormStockRepository) FindByCode(ctx context.Context, code string) (*domain.Stock, error)
```

## 🔧 依存性注入ルール

### コンストラクタパターン
```go
// ✅ 良い例: 依存性を明示的にコンストラクタで受け取る
type CommandHandler struct {
    stockRepo    StockRepository
    externalSvc  ExternalDataService
    notifySvc    NotificationService
    portfolioSvc *domain.PortfolioService
}

func NewCommandHandler(stockRepo StockRepository, externalSvc ExternalDataService, notifySvc NotificationService) *CommandHandler {
    return &CommandHandler{
        stockRepo:    stockRepo,
        externalSvc:  externalSvc,
        notifySvc:    notifySvc,
        portfolioSvc: &domain.PortfolioService{},
    }
}

// ❌ 悪い例: グローバル変数や内部での初期化
var globalDB *gorm.DB

func (h *CommandHandler) CollectStockData(ctx context.Context, cmd CollectStockDataCommand) error {
    // globalDBを直接使用するのは避ける
    globalDB.Save(data)
}
```

## 📝 コメントルール

### ドメインロジックのコメント
```go
// ✅ 良い例: ビジネスルールの説明
// CalculatePerformance calculates portfolio performance metrics.
// It computes total value, gain/loss, and percentage returns based on current market prices.
func (s *PortfolioService) CalculatePerformance(portfolio Portfolio, prices map[string]float64) PortfolioSummary {
    // RSI計算: 14日間の上昇・下降の平均を使用
    // Yahoo Finance APIの制限: 1秒間に最大5回までのリクエスト
}
```

### インターフェースのコメント
```go
// ✅ 良い例: インターフェースの役割を明確に
// StockRepository provides methods to persist and retrieve stock data.
// Implementations should handle database operations and ensure data consistency.
type StockRepository interface {
    // Save persists a stock entity to the repository
    Save(ctx context.Context, stock Stock) error
    
    // FindByCode retrieves a stock by its symbol code
    FindByCode(ctx context.Context, code string) (*Stock, error)
}
```

## 🔒 セキュリティルール

### 機密情報の取り扱い
```go
// ✅ 良い例: 環境変数から取得
func NewYahooFinanceService() *YahooFinanceService {
    apiKey := os.Getenv("YAHOO_API_KEY")
    if apiKey == "" {
        log.Fatal("YAHOO_API_KEY environment variable is required")
    }
    return &YahooFinanceService{apiKey: apiKey}
}

// ❌ 悪い例: ハードコード
const apiKey = "abc123xyz" // 絶対に禁止
```

## ✅ コードレビューチェックリスト

### アーキテクチャ観点
- [ ] 各層の責任が適切に分離されている
- [ ] 依存性の方向が正しい（内側→外側への依存なし）
- [ ] ユースケースが集約単位でリポジトリを呼び出している

### テスト観点
- [ ] ドメインロジックがユニットテストされている
- [ ] 統合テストが適切にデータベースを使用している
- [ ] モックの使用が最小限に抑えられている

### エラーハンドリング観点
- [ ] エラーの発生源でカスタムエラーが作成されている
- [ ] 中間層でエラーが適切にバブリングされている
- [ ] 必要に応じて追加コンテキストが付与されている

### コード品質観点
- [ ] 命名規則が守られている
- [ ] 依存性注入が適切に実装されている
- [ ] セキュリティルールが守られている