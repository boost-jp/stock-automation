# プロジェクト管理（1人プロジェクト版）

## 基本的な開発フロー

1. **アイデア・要求** → GitHub Issue作成
2. **実装** → 機能ブランチで開発
3. **品質チェック** → テスト・lint実行
4. **PR作成** → セルフレビュー後マージ
5. **デプロイ・確認** → 動作確認

## Issue管理

### Issue種別
- `[FEATURE]` - 新機能・改善
- `[BUG]` - バグ・不具合
- `[TASK]` - 開発・運用作業

### ラベル
- `enhancement`, `bug`, `task` - 種別
- `priority/high`, `priority/low` - 優先度

### Issue作成時のポイント
- **何を・なぜ・どうやって** を簡潔に記載
- 完了条件をチェックボックスで明確化
- 工数見積を記載

## 開発規約（要点）

### ブランチ命名
```bash
feature/123-description  # 機能開発
fix/456-bug-name        # バグ修正
```

### コミットメッセージ
```bash
feat: add daily report generation
fix: resolve API timeout issue
docs: update README
```

### コーディングスタイル
```go
// Go規約準拠
type StockPrice struct {}     // PascalCase
func GetStock() {}            // PascalCase
var stockCode string          // camelCase

// エラーハンドリング
if err != nil {
    return fmt.Errorf("failed to get stock: %w", err)
}
```

## 品質管理

### 実装時チェック
```bash
make test      # テスト実行
make lint      # 静的解析
make security  # セキュリティチェック
```

### PR作成前チェック
- [ ] 動作確認完了
- [ ] テスト通過
- [ ] lint エラーなし
- [ ] 不要コード削除
- [ ] コミットメッセージ適切

### AIレビュー活用
ChatGPT/Claudeに以下を依頼：
- コードレビュー（バグ・パフォーマンス・セキュリティ）
- テストケース作成支援
- ドキュメント生成支援

## ファイル構成

```
docs/project/
├── README.md              # このファイル（概要）
└── development-workflow.md # 詳細な開発手順

.github/
├── ISSUE_TEMPLATE/        # Issue テンプレート
└── PULL_REQUEST_TEMPLATE/ # PR テンプレート

backend/
├── Makefile              # ビルド・テストコマンド
├── .golangci.yml         # lint設定
├── .gosec.json           # セキュリティ設定
└── .pre-commit-config.yaml # commit前チェック
```

## 参考リンク

- [開発ワークフロー詳細](./development-workflow.md)
- [コーディング規約](../dev/coding-rules.md)
- [開発者ガイド](../dev/development.md)