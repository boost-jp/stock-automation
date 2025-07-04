# LINE通知設定ガイド

## 概要
LINE Notifyを使用した株価アラート・投資レポートの自動通知システム設定方法

## LINE Notifyとは
LINEが提供する通知サービスで、外部サービスからLINEに通知を送信できる無料サービス

## Step 1: LINE Notifyアカウント設定

### 1.1 LINE Notifyにアクセス
1. ブラウザで https://notify-bot.line.me/ にアクセス
2. 「ログイン」をクリック
3. LINEアカウントでログイン

### 1.2 トークンの発行
1. 右上のアカウント名をクリック→「マイページ」
2. 「トークンを発行する」をクリック
3. 以下の項目を設定：
   - **トークン名**: 株価通知システム
   - **通知を送信するトークルーム**: 「1:1でLINE Notifyから通知を受け取る」
4. 「発行する」をクリック
5. **重要**: 表示されたトークンをコピー（再表示されません）

### 1.3 トークンの保存
```
例: abcdefg123456789hijklmnopqrstuvwxyz
```
このトークンは後でApps Scriptで使用します。

## Step 2: Apps Scriptでの実装

### 2.1 基本的な通知関数
```javascript
// LINE Notify設定
const LINE_NOTIFY_TOKEN = 'YOUR_TOKEN_HERE'; // 実際のトークンに置き換え

function sendLineNotify(message) {
  const url = 'https://notify-api.line.me/api/notify';
  const payload = {
    'message': message
  };
  
  const options = {
    'method': 'POST',
    'headers': {
      'Authorization': 'Bearer ' + LINE_NOTIFY_TOKEN,
      'Content-Type': 'application/x-www-form-urlencoded'
    },
    'payload': payload
  };
  
  try {
    const response = UrlFetchApp.fetch(url, options);
    const responseCode = response.getResponseCode();
    
    if (responseCode === 200) {
      console.log('LINE通知送信成功');
      return true;
    } else {
      console.error('LINE通知送信失敗:', response.getContentText());
      return false;
    }
  } catch (error) {
    console.error('LINE通知送信エラー:', error);
    return false;
  }
}
```

### 2.2 セキュアなトークン管理
```javascript
// トークンをプロパティストアに保存
function setLineNotifyToken() {
  const token = 'YOUR_ACTUAL_TOKEN_HERE'; // 実際のトークン
  PropertiesService.getScriptProperties().setProperty('LINE_NOTIFY_TOKEN', token);
  console.log('LINE Notifyトークンを設定しました');
}

// トークンを取得
function getLineNotifyToken() {
  return PropertiesService.getScriptProperties().getProperty('LINE_NOTIFY_TOKEN');
}

// セキュアな通知関数
function sendLineNotifySecure(message) {
  const token = getLineNotifyToken();
  if (!token) {
    console.error('LINE Notifyトークンが設定されていません');
    return false;
  }
  
  const url = 'https://notify-api.line.me/api/notify';
  const payload = { 'message': message };
  
  const options = {
    'method': 'POST',
    'headers': { 'Authorization': 'Bearer ' + token },
    'payload': payload
  };
  
  try {
    const response = UrlFetchApp.fetch(url, options);
    return response.getResponseCode() === 200;
  } catch (error) {
    console.error('LINE通知送信エラー:', error);
    return false;
  }
}
```

## Step 3: 通知タイプ別の実装

### 3.1 価格アラート通知
```javascript
function sendPriceAlert(stockName, stockCode, currentPrice, targetPrice, alertType) {
  const emoji = alertType === '買い' ? '🔔' : '🔴';
  const arrow = alertType === '買い' ? '⬇️' : '⬆️';
  
  let message = `${emoji} ${alertType}シグナル\n`;
  message += `${stockName}(${stockCode})\n`;
  message += `現在価格: ${currentPrice.toLocaleString()}円 ${arrow}\n`;
  message += `目標価格: ${targetPrice.toLocaleString()}円\n`;
  message += `確認してください！`;
  
  return sendLineNotifySecure(message);
}
```

### 3.2 日次レポート通知
```javascript
function sendDailyReport(portfolioData) {
  const today = new Date().toLocaleDateString('ja-JP');
  let message = `📊 ${today} 投資レポート\n\n`;
  
  let totalValue = 0;
  let totalProfit = 0;
  
  portfolioData.forEach(stock => {
    totalValue += stock.value;
    totalProfit += stock.profit;
    
    const profitIndicator = stock.profit > 0 ? '📈' : stock.profit < 0 ? '📉' : '➡️';
    message += `${profitIndicator} ${stock.name}\n`;
    message += `${stock.currentPrice.toLocaleString()}円 (${stock.profitRate > 0 ? '+' : ''}${stock.profitRate.toFixed(1)}%)\n\n`;
  });
  
  const totalProfitRate = ((totalProfit / (totalValue - totalProfit)) * 100).toFixed(1);
  const profitEmoji = totalProfit > 0 ? '🎉' : totalProfit < 0 ? '😢' : '😐';
  
  message += `${profitEmoji} 合計\n`;
  message += `評価額: ${totalValue.toLocaleString()}円\n`;
  message += `損益: ${totalProfit > 0 ? '+' : ''}${totalProfit.toLocaleString()}円 (${totalProfitRate}%)`;
  
  return sendLineNotifySecure(message);
}
```

### 3.3 投資判断アラート
```javascript
function sendInvestmentAlert(analysis) {
  const confidenceEmoji = analysis.confidence >= 90 ? '🎯' : 
                         analysis.confidence >= 80 ? '⚡' : '💡';
  
  let message = `${confidenceEmoji} 投資判断アラート\n`;
  message += `${analysis.stockName}(${analysis.stockCode})\n\n`;
  message += `判断: ${analysis.recommendation}\n`;
  message += `信頼度: ${analysis.confidence}%\n\n`;
  message += `根拠:\n`;
  
  const allSignals = analysis.buySignals.concat(analysis.sellSignals);
  allSignals.forEach(signal => {
    message += `• ${signal}\n`;
  });
  
  return sendLineNotifySecure(message);
}
```

### 3.4 エラー・システム通知
```javascript
function sendSystemAlert(errorType, details) {
  const errorEmojis = {
    'API_ERROR': '⚠️',
    'SYSTEM_ERROR': '🚨',
    'WARNING': '⚡',
    'INFO': 'ℹ️'
  };
  
  const emoji = errorEmojis[errorType] || '❓';
  let message = `${emoji} システム通知\n`;
  message += `種類: ${errorType}\n`;
  message += `詳細: ${details}\n`;
  message += `時刻: ${new Date().toLocaleString('ja-JP')}`;
  
  return sendLineNotifySecure(message);
}
```

## Step 4: 通知スケジュール設定

### 4.1 定期通知の設定
```javascript
function setupNotificationSchedule() {
  // 毎日18:00に日次レポート
  ScriptApp.newTrigger('sendDailyReportTrigger')
    .timeBased()
    .everyDays(1)
    .atHour(18)
    .create();
  
  // 平日9:00に市場開始通知
  ScriptApp.newTrigger('sendMarketOpenAlert')
    .timeBased()
    .everyDays(1)
    .atHour(9)
    .create();
  
  // 平日15:30に価格アラートチェック
  ScriptApp.newTrigger('checkPriceAlerts')
    .timeBased()
    .everyDays(1)
    .atHour(15)
    .nearMinute(30)
    .create();
}
```

### 4.2 条件付き通知
```javascript
function checkAndNotify() {
  // 土日は通知しない
  const today = new Date();
  const dayOfWeek = today.getDay();
  if (dayOfWeek === 0 || dayOfWeek === 6) {
    console.log('土日のため通知をスキップ');
    return;
  }
  
  // 市場時間中のみアラート
  const hour = today.getHours();
  if (hour >= 9 && hour <= 15) {
    checkPriceAlerts();
  }
  
  // 18:00にのみ日次レポート
  if (hour === 18) {
    generateAndSendDailyReport();
  }
}
```

## Step 5: 高度な通知機能

### 5.1 画像付き通知（グラフなど）
```javascript
function sendNotificationWithImage(message, imageBlob) {
  const url = 'https://notify-api.line.me/api/notify';
  const token = getLineNotifyToken();
  
  const payload = {
    'message': message
  };
  
  const options = {
    'method': 'POST',
    'headers': {
      'Authorization': 'Bearer ' + token
    },
    'payload': payload
  };
  
  // 画像がある場合
  if (imageBlob) {
    options.payload['imageFile'] = imageBlob;
  }
  
  try {
    const response = UrlFetchApp.fetch(url, options);
    return response.getResponseCode() === 200;
  } catch (error) {
    console.error('画像付き通知送信エラー:', error);
    return false;
  }
}
```

### 5.2 通知の重複防止
```javascript
function sendNotificationWithDuplicateCheck(message, cacheKey) {
  const cache = CacheService.getScriptCache();
  const lastMessage = cache.get(cacheKey);
  
  // 同じメッセージは1時間以内は送信しない
  if (lastMessage === message) {
    console.log('重複通知のため送信をスキップ');
    return false;
  }
  
  const success = sendLineNotifySecure(message);
  if (success) {
    cache.put(cacheKey, message, 3600); // 1時間キャッシュ
  }
  
  return success;
}
```

### 5.3 通知の優先度制御
```javascript
function sendPriorityNotification(message, priority) {
  const priorities = {
    'HIGH': '🚨',    // 損切りライン到達など
    'MEDIUM': '⚡',  // 強い売買シグナル
    'LOW': 'ℹ️'      // 一般的な情報
  };
  
  const emoji = priorities[priority] || '';
  const finalMessage = `${emoji} ${message}`;
  
  // 高優先度は即座に送信、低優先度は制限あり
  if (priority === 'HIGH') {
    return sendLineNotifySecure(finalMessage);
  } else {
    return sendNotificationWithDuplicateCheck(finalMessage, `priority_${priority}`);
  }
}
```

## Step 6: 通知内容のカスタマイズ

### 6.1 通知設定の管理
```javascript
function getNotificationSettings() {
  const settings = PropertiesService.getScriptProperties().getProperties();
  
  return {
    dailyReport: settings.DAILY_REPORT_ENABLED === 'true',
    priceAlerts: settings.PRICE_ALERTS_ENABLED === 'true',
    investmentAlerts: settings.INVESTMENT_ALERTS_ENABLED === 'true',
    minConfidence: parseInt(settings.MIN_CONFIDENCE) || 80,
    reportTime: parseInt(settings.REPORT_TIME) || 18
  };
}

function updateNotificationSettings(newSettings) {
  const properties = PropertiesService.getScriptProperties();
  
  Object.keys(newSettings).forEach(key => {
    properties.setProperty(key.toUpperCase(), newSettings[key].toString());
  });
  
  console.log('通知設定を更新しました');
}
```

### 6.2 メッセージテンプレート
```javascript
const MESSAGE_TEMPLATES = {
  PRICE_ALERT: (data) => `
🔔 ${data.alertType}シグナル
${data.stockName}(${data.stockCode})
現在価格: ${data.currentPrice.toLocaleString()}円
目標価格: ${data.targetPrice.toLocaleString()}円
確認してください！
  `.trim(),
  
  DAILY_REPORT: (data) => `
📊 ${data.date} 投資レポート

${data.stocks.map(stock => 
  `${stock.profitIndicator} ${stock.name}: ${stock.currentPrice.toLocaleString()}円 (${stock.profitRate}%)`
).join('\n')}

💰 合計
評価額: ${data.totalValue.toLocaleString()}円
損益: ${data.totalProfit > 0 ? '+' : ''}${data.totalProfit.toLocaleString()}円 (${data.totalProfitRate}%)
  `.trim(),
  
  INVESTMENT_ALERT: (data) => `
🎯 投資判断アラート
${data.stockName}(${data.stockCode})

判断: ${data.recommendation}
信頼度: ${data.confidence}%

根拠:
${data.signals.map(signal => `• ${signal}`).join('\n')}
  `.trim()
};
```

## Step 7: テストと運用

### 7.1 通知テスト関数
```javascript
function testLineNotifications() {
  console.log('=== LINE通知テスト開始 ===');
  
  // 基本通知テスト
  const basicTest = sendLineNotifySecure('📊 テスト通知: システムが正常に動作しています');
  console.log('基本通知:', basicTest ? '成功' : '失敗');
  
  // 価格アラートテスト
  const alertTest = sendPriceAlert('テスト銘柄', 'TEST', 1000, 950, '買い');
  console.log('価格アラート:', alertTest ? '成功' : '失敗');
  
  // システムアラートテスト
  const systemTest = sendSystemAlert('INFO', 'システムテストが実行されました');
  console.log('システムアラート:', systemTest ? '成功' : '失敗');
  
  console.log('=== LINE通知テスト完了 ===');
}
```

### 7.2 エラーハンドリング
```javascript
function robustLineNotify(message, retryCount = 3) {
  for (let i = 0; i < retryCount; i++) {
    try {
      const success = sendLineNotifySecure(message);
      if (success) {
        return true;
      }
    } catch (error) {
      console.warn(`通知送信失敗 (試行${i + 1}/${retryCount}):`, error);
      if (i < retryCount - 1) {
        Utilities.sleep(2000); // 2秒待機してリトライ
      }
    }
  }
  
  console.error('LINE通知の送信に失敗しました');
  return false;
}
```

## Step 8: 使用制限と注意事項

### API制限
- **送信制限**: 1時間に1000回まで
- **文字数制限**: 1回の送信で1000文字まで
- **画像サイズ**: 最大1MB

### 推奨される使用方法
1. **重要な通知のみ**: 高頻度の通知は避ける
2. **バッチ処理**: 複数の情報をまとめて送信
3. **エラーハンドリング**: 送信失敗に対する適切な処理

### セキュリティ対策
1. **トークンの管理**: プロパティストアで安全に保存
2. **アクセス制御**: スクリプトの共有範囲を制限
3. **ログ管理**: 送信履歴の適切な記録

これでLINE通知システムの設定が完了します。次は[IFTTT連携](ifttt-setup.md)で更に高度な通知システムを構築できます。