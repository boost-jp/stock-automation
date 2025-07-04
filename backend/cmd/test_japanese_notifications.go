package main

import (
	"fmt"
	"log"
	"os"

	"stock-automation/internal/notification"
)

func main() {
	// Slack Webhook URLを設定
	webhookURL := "https://hooks.slack.com/services/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"
	os.Setenv("SLACK_WEBHOOK_URL", webhookURL)

	// Slack通知サービス初期化
	notifier := notification.NewSlackNotifier()

	fmt.Println("🧪 日本語文字化けテスト開始")

	// テスト1: 基本的な日本語メッセージ
	fmt.Println("\n📱 基本日本語メッセージテスト")
	message1 := `🇯🇵 日本語文字エンコーディングテスト
株価情報システムの通知テストです。

📊 監視銘柄:
• トヨタ自動車 (7203): ¥2,484.50
• ソニーグループ (6758): ¥3,688.00  
• 任天堂 (7974): ¥13,200.00
• キーエンス (6861): ¥57,000.00

✅ 漢字・ひらがな・カタカナ・記号が正しく表示されているかチェック
🔔 アラート機能動作確認中...`

	if err := notifier.SendMessage(message1); err != nil {
		log.Fatal("Failed to send Japanese test message:", err)
	}
	fmt.Println("✅ 基本日本語メッセージ送信完了")

	// テスト2: 株価アラート（日本語）
	fmt.Println("\n🔔 日本語株価アラートテスト")
	if err := notifier.SendStockAlert("7203", "トヨタ自動車", 2484.50, 2500.00, "買い推奨"); err != nil {
		log.Fatal("Failed to send Japanese stock alert:", err)
	}
	fmt.Println("✅ 日本語株価アラート送信完了")

	// テスト3: 複雑な日本語文字列
	fmt.Println("\n🚀 複雑な日本語文字列テスト")
	message3 := `📈 投資判断システム稼働状況報告

🏆 本日のベストパフォーマンス:
1位: ファーストリテイリング (+2.5%)
2位: 三井住友フィナンシャルグループ (+1.8%)
3位: 武田薬品工業 (+1.2%)

⚠️ 注意銘柄:
• みずほフィナンシャルグループ: RSI買われすぎ警告
• 日立製作所: デッドクロス発生

💡 テクニカル分析結果:
移動平均線: ゴールデンクロス形成中
相対力指数: 中立圏内（45-55）
出来高: 平均比120%で活発

📊 次回分析予定: 2025年7月5日 15:30`

	if err := notifier.SendMessage(message3); err != nil {
		log.Fatal("Failed to send complex Japanese message:", err)
	}
	fmt.Println("✅ 複雑な日本語文字列送信完了")

	// テスト4: 特殊文字・記号
	fmt.Println("\n🎯 特殊文字・記号テスト")
	message4 := `🔢 特殊文字・記号テスト

💰 金額表記: ¥1,234,567
📊 パーセント: +12.34%、-5.67%
📈 矢印: ↗️↘️⬆️⬇️
🎌 日本語記号: 【重要】※注意※
🔘 円記号: ○×△▲▼
📱 絵文字: 🚗🏦💻📡🏭💊🛒

テスト完了予定時刻: 2025/07/05 00:30:00`

	if err := notifier.SendMessage(message4); err != nil {
		log.Fatal("Failed to send special characters message:", err)
	}
	fmt.Println("✅ 特殊文字・記号送信完了")

	fmt.Println("\n🎉 すべての日本語文字化けテストが完了しました！")
	fmt.Println("Slackで通知内容を確認してください。")
}