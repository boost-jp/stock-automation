# Google Apps Script実装ガイド

## 概要
Google Apps Scriptを使用した株価データ自動取得・更新システムの実装方法

## Apps Scriptの基本設定

### Step 1: Apps Scriptエディタを開く
1. Google Sheetsで「拡張機能」→「Apps Script」
2. 新しいプロジェクトが作成される
3. プロジェクト名を「株価自動化システム」に変更

### Step 2: 基本コードの実装

#### メインの株価更新関数
```javascript
function updateStockPrices() {
  const sheet = SpreadsheetApp.getActiveSheet();
  const lastRow = sheet.getLastRow();
  
  console.log('株価更新開始: ' + new Date());
  
  for (let i = 2; i <= lastRow; i++) {
    const stockCode = sheet.getRange(i, 1).getValue();
    if (stockCode) {
      try {
        const price = getStockPrice(stockCode);
        if (price) {
          sheet.getRange(i, 5).setValue(price);
          console.log(`${stockCode}: ${price}円`);
        }
      } catch (error) {
        console.error(`${stockCode}のデータ取得エラー:`, error);
      }
      
      // API制限対策で1秒待機
      Utilities.sleep(1000);
    }
  }
  
  console.log('株価更新完了: ' + new Date());
}
```

#### Yahoo Finance APIから株価取得
```javascript
function getStockPrice(stockCode) {
  try {
    const url = `https://query1.finance.yahoo.com/v8/finance/chart/${stockCode}.T`;
    const response = UrlFetchApp.fetch(url);
    const data = JSON.parse(response.getContentText());
    
    if (data.chart.result && data.chart.result.length > 0) {
      const result = data.chart.result[0];
      const currentPrice = result.meta.regularMarketPrice;
      return currentPrice;
    }
    return null;
  } catch (error) {
    console.error('株価取得エラー:', error);
    return null;
  }
}
```

#### 出来高データ取得
```javascript
function getCurrentVolume(stockCode) {
  try {
    const url = `https://query1.finance.yahoo.com/v8/finance/chart/${stockCode}.T`;
    const response = UrlFetchApp.fetch(url);
    const data = JSON.parse(response.getContentText());
    
    if (data.chart.result && data.chart.result.length > 0) {
      const result = data.chart.result[0];
      const volume = result.meta.regularMarketVolume;
      return volume;
    }
    return null;
  } catch (error) {
    console.error('出来高取得エラー:', error);
    return null;
  }
}
```

## トリガー設定

### 定期実行トリガーの作成
```javascript
function createTriggers() {
  // 既存のトリガーを削除
  const triggers = ScriptApp.getProjectTriggers();
  triggers.forEach(trigger => ScriptApp.deleteTrigger(trigger));
  
  // 平日の9:00に実行（寄り付き前）
  ScriptApp.newTrigger('updateStockPrices')
    .timeBased()
    .everyDays(1)
    .atHour(9)
    .create();
  
  // 平日の15:30に実行（大引け後）
  ScriptApp.newTrigger('updateStockPrices')
    .timeBased()
    .everyDays(1)
    .atHour(15)
    .nearMinute(30)
    .create();
  
  // 平日の18:00に日次レポート送信
  ScriptApp.newTrigger('generateDailyReport')
    .timeBased()
    .everyDays(1)
    .atHour(18)
    .create();
  
  console.log('トリガー設定完了');
}
```

### 手動でトリガー作成（UI使用）
1. Apps Scriptエディタで「トリガー」アイコンをクリック
2. 「トリガーを追加」をクリック
3. 以下を設定：
   - 実行する関数: `updateStockPrices`
   - イベントソース: `時間主導型`
   - 時間ベースのトリガー: `日タイマー`
   - 時刻: `午前9時〜10時`

## LINE通知システム

### LINE Notify設定
```javascript
// LINE Notify設定（トークンは実際の値に置き換え）
const LINE_NOTIFY_TOKEN = 'YOUR_LINE_NOTIFY_TOKEN_HERE';

function sendLineNotify(message) {
  const url = 'https://notify-api.line.me/api/notify';
  const payload = {
    'message': message
  };
  
  const options = {
    'method': 'POST',
    'headers': {
      'Authorization': 'Bearer ' + LINE_NOTIFY_TOKEN,
    },
    'payload': payload
  };
  
  try {
    const response = UrlFetchApp.fetch(url, options);
    if (response.getResponseCode() === 200) {
      console.log('LINE通知送信成功');
    } else {
      console.error('LINE通知送信失敗:', response.getContentText());
    }
  } catch (error) {
    console.error('LINE通知送信エラー:', error);
  }
}
```

### 価格アラート機能
```javascript
function checkPriceAlerts() {
  const sheet = SpreadsheetApp.getActiveSpreadsheet().getSheetByName('監視銘柄');
  const lastRow = sheet.getLastRow();
  
  for (let i = 2; i <= lastRow; i++) {
    const stockCode = sheet.getRange(i, 1).getValue();
    const stockName = sheet.getRange(i, 2).getValue();
    const targetBuy = sheet.getRange(i, 3).getValue();
    const targetSell = sheet.getRange(i, 4).getValue();
    const currentPrice = sheet.getRange(i, 5).getValue();
    
    // 買いシグナル
    if (currentPrice && targetBuy && currentPrice <= targetBuy) {
      const message = `🔔 買いシグナル\n${stockName}(${stockCode})\n現在価格: ${currentPrice.toLocaleString()}円\n目標買い価格: ${targetBuy.toLocaleString()}円`;
      sendLineNotify(message);
    }
    
    // 売りシグナル
    if (currentPrice && targetSell && currentPrice >= targetSell) {
      const message = `🔔 売りシグナル\n${stockName}(${stockCode})\n現在価格: ${currentPrice.toLocaleString()}円\n目標売り価格: ${targetSell.toLocaleString()}円`;
      sendLineNotify(message);
    }
  }
}
```

## 日次レポート生成

### レポート生成関数
```javascript
function generateDailyReport() {
  const sheet = SpreadsheetApp.getActiveSpreadsheet().getSheetByName('ポートフォリオ');
  const lastRow = sheet.getLastRow();
  
  let totalValue = 0;
  let totalProfit = 0;
  let reportText = `📊 ${new Date().toLocaleDateString('ja-JP')} 投資レポート\n\n`;
  
  // 各銘柄の情報を集計
  for (let i = 2; i <= lastRow; i++) {
    const stockName = sheet.getRange(i, 2).getValue();
    const shares = sheet.getRange(i, 3).getValue();
    const currentPrice = sheet.getRange(i, 5).getValue();
    const value = sheet.getRange(i, 6).getValue();
    const profit = sheet.getRange(i, 7).getValue();
    const profitRate = sheet.getRange(i, 8).getValue();
    
    if (stockName && shares && currentPrice) {
      totalValue += value;
      totalProfit += profit;
      
      const profitIndicator = profit > 0 ? '📈' : profit < 0 ? '📉' : '➡️';
      reportText += `${profitIndicator} ${stockName}\n`;
      reportText += `価格: ${currentPrice.toLocaleString()}円\n`;
      reportText += `評価額: ${value.toLocaleString()}円\n`;
      reportText += `損益: ${profit > 0 ? '+' : ''}${profit.toLocaleString()}円 (${(profitRate * 100).toFixed(1)}%)\n\n`;
    }
  }
  
  // 合計サマリー
  const totalProfitRate = ((totalProfit / (totalValue - totalProfit)) * 100).toFixed(1);
  const profitIndicator = totalProfit > 0 ? '🎉' : totalProfit < 0 ? '😢' : '😐';
  
  reportText += `${profitIndicator} 合計サマリー\n`;
  reportText += `総評価額: ${totalValue.toLocaleString()}円\n`;
  reportText += `総損益: ${totalProfit > 0 ? '+' : ''}${totalProfit.toLocaleString()}円\n`;
  reportText += `総損益率: ${totalProfitRate}%`;
  
  // LINE通知送信
  sendLineNotify(reportText);
  
  // 価格アラートもチェック
  checkPriceAlerts();
}
```

## テクニカル指標計算

### 移動平均線計算
```javascript
function calculateMovingAverage(stockCode, period) {
  try {
    const url = `https://query1.finance.yahoo.com/v8/finance/chart/${stockCode}.T?interval=1d&range=${period + 5}d`;
    const response = UrlFetchApp.fetch(url);
    const data = JSON.parse(response.getContentText());
    
    if (data.chart.result && data.chart.result.length > 0) {
      const prices = data.chart.result[0].indicators.quote[0].close;
      const validPrices = prices.filter(price => price !== null);
      
      if (validPrices.length >= period) {
        const recentPrices = validPrices.slice(-period);
        const sum = recentPrices.reduce((a, b) => a + b, 0);
        return sum / period;
      }
    }
    return null;
  } catch (error) {
    console.error('移動平均線計算エラー:', error);
    return null;
  }
}
```

### RSI計算
```javascript
function calculateRSI(stockCode, period = 14) {
  try {
    const url = `https://query1.finance.yahoo.com/v8/finance/chart/${stockCode}.T?interval=1d&range=${period + 10}d`;
    const response = UrlFetchApp.fetch(url);
    const data = JSON.parse(response.getContentText());
    
    if (data.chart.result && data.chart.result.length > 0) {
      const prices = data.chart.result[0].indicators.quote[0].close;
      const validPrices = prices.filter(price => price !== null);
      
      if (validPrices.length >= period + 1) {
        let gains = [];
        let losses = [];
        
        for (let i = 1; i < validPrices.length; i++) {
          const change = validPrices[i] - validPrices[i-1];
          if (change > 0) {
            gains.push(change);
            losses.push(0);
          } else {
            gains.push(0);
            losses.push(Math.abs(change));
          }
        }
        
        const recentGains = gains.slice(-period);
        const recentLosses = losses.slice(-period);
        
        const avgGain = recentGains.reduce((a, b) => a + b, 0) / period;
        const avgLoss = recentLosses.reduce((a, b) => a + b, 0) / period;
        
        if (avgLoss === 0) return 100;
        
        const rs = avgGain / avgLoss;
        const rsi = 100 - (100 / (1 + rs));
        
        return rsi;
      }
    }
    return null;
  } catch (error) {
    console.error('RSI計算エラー:', error);
    return null;
  }
}
```

## エラーハンドリング

### 堅牢性向上のための実装
```javascript
function updateStockPricesWithErrorHandling() {
  try {
    const sheet = SpreadsheetApp.getActiveSheet();
    const lastRow = sheet.getLastRow();
    
    console.log('株価更新開始: ' + new Date());
    
    for (let i = 2; i <= lastRow; i++) {
      const stockCode = sheet.getRange(i, 1).getValue();
      if (stockCode) {
        let retryCount = 0;
        let success = false;
        
        while (retryCount < 3 && !success) {
          try {
            const price = getStockPrice(stockCode);
            if (price) {
              sheet.getRange(i, 5).setValue(price);
              success = true;
              console.log(`${stockCode}: ${price}円`);
            }
          } catch (error) {
            retryCount++;
            console.warn(`${stockCode}のデータ取得エラー (試行${retryCount}/3):`, error);
            Utilities.sleep(2000); // 2秒待機してリトライ
          }
        }
        
        if (!success) {
          console.error(`${stockCode}のデータ取得に3回失敗しました`);
        }
        
        Utilities.sleep(1000); // API制限対策
      }
    }
    
    console.log('株価更新完了: ' + new Date());
  } catch (error) {
    console.error('updateStockPrices実行エラー:', error);
    sendLineNotify(`⚠️ システムエラー\n株価更新処理でエラーが発生しました。\n${error.toString()}`);
  }
}
```

## デバッグとテスト

### テスト用関数
```javascript
function testStockPriceAPI() {
  const testCodes = ['7203', '6758', '9984'];
  
  testCodes.forEach(code => {
    console.log(`=== ${code} テスト ===`);
    const price = getStockPrice(code);
    const volume = getCurrentVolume(code);
    const ma5 = calculateMovingAverage(code, 5);
    const rsi = calculateRSI(code);
    
    console.log(`価格: ${price}`);
    console.log(`出来高: ${volume}`);
    console.log(`MA5: ${ma5}`);
    console.log(`RSI: ${rsi}`);
  });
}

function testLineNotify() {
  sendLineNotify('📊 テスト通知\nApps Scriptからの通知テストです。');
}
```

## セキュリティ設定

### プロパティストアの使用（トークン管理）
```javascript
function setLineNotifyToken() {
  const token = 'YOUR_ACTUAL_TOKEN_HERE';
  PropertiesService.getScriptProperties().setProperty('LINE_NOTIFY_TOKEN', token);
  console.log('LINE Notifyトークンを設定しました');
}

function getLineNotifyToken() {
  return PropertiesService.getScriptProperties().getProperty('LINE_NOTIFY_TOKEN');
}

// 修正版のLINE通知関数
function sendLineNotifySecure(message) {
  const token = getLineNotifyToken();
  if (!token) {
    console.error('LINE Notifyトークンが設定されていません');
    return;
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
    console.log('LINE通知送信成功');
  } catch (error) {
    console.error('LINE通知送信エラー:', error);
  }
}
```

## 導入手順まとめ

### 1. コードの貼り付け
1. Apps Scriptエディタで上記のコードを貼り付け
2. `YOUR_LINE_NOTIFY_TOKEN_HERE`を実際のトークンに置き換え
3. 保存（Ctrl+S）

### 2. 権限の承認
1. 初回実行時に権限確認ダイアログが表示
2. 「権限を確認」→「安全でないページ」→「許可」

### 3. トリガーの設定
1. `createTriggers()`関数を実行
2. または手動でトリガーを作成

### 4. テスト実行
1. `testStockPriceAPI()`でデータ取得テスト
2. `testLineNotify()`で通知テスト
3. `updateStockPrices()`で全体テスト

これでApps Scriptの実装が完了します。次は[LINE通知設定](line-notify.md)の詳細設定に進んでください。