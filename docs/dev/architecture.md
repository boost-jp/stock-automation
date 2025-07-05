# アーキテクチャ設計

## 📋 概要
株式自動化システムは、シンプルなオニオンアーキテクチャを採用し、テストしやすく保守性の高い設計を実現しています。

## 🏗️ アーキテクチャ構成

### 基本構造
```
Domain Layer (最内層)
  ↓
Usecase Layer (アプリケーションロジック)
  ↓
Repository Layer (データアクセス)
  ↓
Infrastructure Layer
  ↓
Interface Layer (最外層)
```

### ディレクトリ構造
```
backend/
├── app/
│   ├── domain/          # ドメイン層（ビジネスロジック）
│   │   ├── stock.go     # 株式関連ドメイン
│   │   ├── portfolio.go # ポートフォリオ関連
│   │   └── analysis.go  # 分析ロジック
│   ├── usecase/         # ユースケース層（アプリケーションロジック）
│   │   ├── collect_data.go      # データ収集ユースケース
│   │   ├── portfolio_report.go  # レポート生成ユースケース
│   │   └── portfolio_analysis.go # ポートフォリオ分析
│   ├── repository/      # リポジトリ層（データアクセス）
│   │   ├── stock.go     # 株式データリポジトリ
│   │   └── portfolio.go # ポートフォリオリポジトリ
│   ├── infra/           # インフラ層
│   │   ├── external.go  # 外部API
│   │   └── notification.go # 通知
│   └── interfaces/      # インターフェース層
│       ├── scheduler.go # スケジューラー
│       └── cli.go       # CLI
├── pkg/
│   ├── config/          # 設定
│   └── testutil/        # テストユーティリティ
└── tests/
    ├── unit/            # ユニットテスト
    └── integration/     # 統合テスト
```

## 🎯 各層の責務

### Domain Layer（ドメイン層）
- **目的**: ビジネスルールとドメインロジックの実装
- **特徴**: 
  - 外部依存なし
  - 純粋関数中心
  - テストしやすい設計
- **含まれるもの**:
  - エンティティ（Stock, Portfolio, Holding）
  - ドメインサービス（PortfolioService）
  - ビジネスルール検証

### Usecase Layer（ユースケース層）
- **目的**: アプリケーションロジックの実装
- **特徴**:
  - ビジネスフローの調整
  - 集約単位でのリポジトリ呼び出し
  - トランザクション管理
- **含まれるもの**:
  - ユースケース実装（データ収集、レポート生成）
  - リポジトリインターフェース定義
  - 外部サービスインターフェース定義

### Repository Layer（リポジトリ層）
- **目的**: データアクセスの抽象化
- **特徴**:
  - 集約単位での操作
  - データベース実装の隠蔽
  - テスト用実装の差し替え可能
- **含まれるもの**:
  - Repository実装
  - データベース操作
  - クエリ最適化

### Infrastructure Layer（インフラ層）
- **目的**: 外部システムとの接続
- **特徴**:
  - 外部API呼び出し
  - 通知機能
  - 設定管理
- **含まれるもの**:
  - 外部APIクライアント
  - 通知サービス
  - 設定読み込み

### Interface Layer（インターフェース層）
- **目的**: 外部からの入力受付
- **特徴**:
  - スケジューラー
  - CLI
  - 依存性注入
- **含まれるもの**:
  - Scheduler
  - CLI handlers
  - DI container

## 🔄 ユースケース実装パターン

### 基本的なユースケース
```go
// ポートフォリオレポート生成ユースケース
type PortfolioReportUsecase struct {
    stockRepo       StockRepository
    portfolioRepo   PortfolioRepository
    notificationSvc NotificationService
    portfolioSvc    *domain.PortfolioService
}

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

### データ収集ユースケース
```go
type CollectDataUsecase struct {
    stockRepo       StockRepository
    externalDataSvc ExternalDataService
}

func (u *CollectDataUsecase) CollectStockData(ctx context.Context, stockCodes []string, days int) error {
    // 1. 外部APIからデータ取得
    prices, err := u.externalDataSvc.FetchStockPrices(ctx, stockCodes, days)
    if err != nil {
        return err
    }
    
    // 2. データ保存（集約単位）
    return u.stockRepo.SaveStockPrices(ctx, prices)
}
```

## 🏪 リポジトリパターン

### 集約単位の操作
```go
// ポートフォリオリポジトリ
type PortfolioRepository interface {
    GetPortfolio(ctx context.Context) (domain.Portfolio, error)
    SavePortfolio(ctx context.Context, portfolio domain.Portfolio) error
}

// 株式データリポジトリ
type StockRepository interface {
    SaveStockPrices(ctx context.Context, prices []domain.StockPrice) error
    GetLatestPrices(ctx context.Context, codes []string) (map[string]float64, error)
    GetPriceHistory(ctx context.Context, code string, days int) ([]domain.StockPrice, error)
}
```

### リポジトリ実装
```go
type GormPortfolioRepository struct {
    db *gorm.DB
}

func (r *GormPortfolioRepository) GetPortfolio(ctx context.Context) (domain.Portfolio, error) {
    var holdings []domain.Holding
    err := r.db.WithContext(ctx).Find(&holdings).Error
    if err != nil {
        return domain.Portfolio{}, err
    }
    
    return domain.Portfolio{Holdings: holdings}, nil
}
```

## 🧪 テスト戦略

### 基本方針
- **ドメインロジック**: 純粋関数のユニットテスト
- **ユースケース**: リアルDBを使った統合テスト
- **モックは最小限**: 外部API呼び出しのみモック

### テスト構成
```
tests/
├── unit/
│   ├── domain_test.go      # ドメインロジックテスト
│   └── analysis_test.go    # 分析ロジックテスト
└── integration/
    ├── usecase_test.go     # ユースケース統合テスト
    └── repository_test.go  # リポジトリテスト
```

### テスト実装例
```go
// ドメインロジックテスト
func TestPortfolioService_CalculatePerformance(t *testing.T) {
    service := &domain.PortfolioService{}
    portfolio := createTestPortfolio()
    prices := createTestPrices()
    
    summary := service.CalculatePerformance(portfolio, prices)
    
    assert.Equal(t, expectedGain, summary.TotalGain)
}

// ユースケース統合テスト
func TestPortfolioReportUsecase_GenerateAndSendDailyReport(t *testing.T) {
    db := testutil.SetupTestDB(t)
    defer testutil.CleanupTestDB(t, db)
    
    // リアルリポジトリ使用
    stockRepo := repository.NewGormStockRepository(db)
    portfolioRepo := repository.NewGormPortfolioRepository(db)
    
    // モック通知サービス
    mockNotifier := &testutil.MockNotificationService{}
    
    usecase := usecase.NewPortfolioReportUsecase(stockRepo, portfolioRepo, mockNotifier)
    err := usecase.GenerateAndSendDailyReport(context.Background())
    
    assert.NoError(t, err)
    assert.True(t, mockNotifier.ReportSent)
}
```

## 🔧 依存性注入

### シンプルなDI実装
```go
// main.go
func main() {
    // インフラ層の構築
    db := setupDatabase()
    yahooSvc := infra.NewYahooFinanceService()
    notifySvc := infra.NewSlackNotificationService()
    
    // リポジトリ層の構築
    stockRepo := repository.NewGormStockRepository(db)
    portfolioRepo := repository.NewGormPortfolioRepository(db)
    
    // ユースケース層の構築
    collectDataUC := usecase.NewCollectDataUsecase(stockRepo, yahooSvc)
    portfolioReportUC := usecase.NewPortfolioReportUsecase(stockRepo, portfolioRepo, notifySvc)
    
    // インターフェース層の構築
    scheduler := interfaces.NewScheduler(collectDataUC, portfolioReportUC)
    
    // 開始
    scheduler.Start()
}
```

## 🚨 エラーハンドリング

### 基本方針
- **エラーの発生源**: カスタムエラーでメッセージをWrap
- **ユースケース層**: エラーをそのままバブリング
- **必要に応じて**: 追加コンテキストをWrap

### 実装例
```go
// ドメイン層でのエラー生成
func (s *PortfolioService) ValidateHolding(holding Holding) error {
    if holding.Shares <= 0 {
        return errors.NewInvalidArgument("shares must be positive")
    }
    return nil
}

// ユースケース層でのエラーバブリング
func (u *PortfolioReportUsecase) GenerateAndSendDailyReport(ctx context.Context) error {
    portfolio, err := u.portfolioRepo.GetPortfolio(ctx)
    if err != nil {
        return err // そのままバブリング
    }
    
    // 必要に応じて追加コンテキスト
    if err := u.notificationSvc.SendReport(ctx, report); err != nil {
        return errors.Wrap(err, "failed to send daily report notification")
    }
    return nil
}
```

## 📊 利点

1. **シンプルさ**: CQRSのような複雑性を避けた実用的な設計
2. **テストしやすさ**: 各層が独立してテスト可能
3. **保守性**: 責任の明確な分離
4. **拡張性**: 新機能の追加が容易
5. **理解しやすさ**: 一般的なパターンで学習コストが低い

## 🎯 設計原則

1. **依存性の逆転**: 内側の層は外側の層に依存しない
2. **単一責任**: 各層は明確な責任を持つ
3. **テスト容易性**: すべてのコンポーネントが独立してテスト可能
4. **実用性**: オーバーエンジニアリングを避ける
5. **集約指向**: リポジトリは集約単位で操作する

このシンプルなアーキテクチャにより、テストが書きやすく、保守しやすい、実用的なシステムを構築できます。