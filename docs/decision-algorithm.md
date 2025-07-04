# 株式売買判断アルゴリズム

## 概要
テクニカル指標とファンダメンタル要素を組み合わせた客観的な売買判断システム

## 判断ロジックの構成

### スコアリングシステム
- **買いシグナル**: 最大8点
- **売りシグナル**: 最大11点
- **総合判断**: 信頼度とアクション推奨

## 買い判断アルゴリズム

### 評価条件と配点

#### 1. ゴールデンクロス（+3点）
```
条件: MA5 > MA25
説明: 短期移動平均が長期移動平均を上抜ける強いトレンド転換シグナル
```

#### 2. RSI売られすぎ（+2点）
```
条件: RSI ≤ 30
説明: 相対力指数が30以下で売られすぎ状態、反発の可能性
```

#### 3. MA5上抜け（+1点）
```
条件: 現在価格 > MA5
説明: 現在価格が5日移動平均を上回り、短期的な上昇トレンド
```

#### 4. 出来高急増（+1点）
```
条件: 出来高 > 20日平均出来高 × 1.5
説明: 平均の1.5倍以上の出来高で関心の高まり
```

#### 5. 安値からの反発（+1点）
```
条件: 現在価格 > 20日安値 × 1.05
説明: 直近安値から5%以上上昇で底打ちの可能性
```

### 買い判断の実装コード

```javascript
function getBuySignal(stockCode) {
  const currentPrice = getCurrentPrice(stockCode);
  const ma5 = calculateMovingAverage(stockCode, 5);
  const ma25 = calculateMovingAverage(stockCode, 25);
  const rsi = calculateRSI(stockCode);
  const volume = getCurrentVolume(stockCode);
  const avgVolume = getAverageVolume(stockCode, 20);
  const recentLow = getRecentLow(stockCode, 20);
  
  let buyScore = 0;
  let signals = [];
  
  // 条件1: ゴールデンクロス
  if (ma5 > ma25) {
    buyScore += 3;
    signals.push('ゴールデンクロス');
  }
  
  // 条件2: RSI売られすぎ
  if (rsi <= 30) {
    buyScore += 2;
    signals.push('RSI売られすぎ');
  }
  
  // 条件3: MA5上抜け
  if (currentPrice > ma5) {
    buyScore += 1;
    signals.push('MA5上抜け');
  }
  
  // 条件4: 出来高急増
  if (volume > avgVolume * 1.5) {
    buyScore += 1;
    signals.push('出来高急増');
  }
  
  // 条件5: 安値からの反発
  if (currentPrice > recentLow * 1.05) {
    buyScore += 1;
    signals.push('安値からの反発');
  }
  
  return {
    score: buyScore,
    signals: signals,
    recommendation: buyScore >= 4 ? '強い買い' : buyScore >= 2 ? '買い' : '様子見'
  };
}
```

## 売り判断アルゴリズム

### 評価条件と配点

#### 1. デッドクロス（+3点）
```
条件: MA5 < MA25
説明: 短期移動平均が長期移動平均を下抜ける弱いトレンド転換シグナル
```

#### 2. 損切りライン（+4点）★最重要
```
条件: (現在価格 - 取得価格) / 取得価格 ≤ -0.10
説明: 取得価格から10%以上下落で損失限定のための強制売り
```

#### 3. RSI買われすぎ（+2点）
```
条件: RSI ≥ 70
説明: 相対力指数が70以上で買われすぎ状態、調整の可能性
```

#### 4. 利確ライン（+2点）
```
条件: (現在価格 - 取得価格) / 取得価格 ≥ 0.20
説明: 取得価格から20%以上上昇で利益確定の検討
```

#### 5. MA5下抜け（+1点）
```
条件: 現在価格 < MA5
説明: 現在価格が5日移動平均を下回り、短期的な下降トレンド
```

#### 6. 高値からの下落（+1点）
```
条件: 現在価格 < 20日高値 × 0.95
説明: 直近高値から5%以上下落で調整局面入り
```

### 売り判断の実装コード

```javascript
function getSellSignal(stockCode, purchasePrice) {
  const currentPrice = getCurrentPrice(stockCode);
  const ma5 = calculateMovingAverage(stockCode, 5);
  const ma25 = calculateMovingAverage(stockCode, 25);
  const rsi = calculateRSI(stockCode);
  const recentHigh = getRecentHigh(stockCode, 20);
  
  let sellScore = 0;
  let signals = [];
  
  // 損益率計算
  const lossRate = (currentPrice - purchasePrice) / purchasePrice;
  
  // 条件1: デッドクロス
  if (ma5 < ma25) {
    sellScore += 3;
    signals.push('デッドクロス');
  }
  
  // 条件2: 損切りライン（最重要）
  if (lossRate <= -0.10) {
    sellScore += 4;
    signals.push('損切りライン');
  }
  
  // 条件3: RSI買われすぎ
  if (rsi >= 70) {
    sellScore += 2;
    signals.push('RSI買われすぎ');
  }
  
  // 条件4: 利確ライン
  if (lossRate >= 0.20) {
    sellScore += 2;
    signals.push('利確ライン');
  }
  
  // 条件5: MA5下抜け
  if (currentPrice < ma5) {
    sellScore += 1;
    signals.push('MA5下抜け');
  }
  
  // 条件6: 高値からの下落
  if (currentPrice < recentHigh * 0.95) {
    sellScore += 1;
    signals.push('高値からの下落');
  }
  
  return {
    score: sellScore,
    signals: signals,
    recommendation: sellScore >= 4 ? '強い売り' : sellScore >= 2 ? '売り' : '保持'
  };
}
```

## 総合判断システム

### 保有状況別の判断ロジック

```javascript
function comprehensiveAnalysis(stockCode, holdingData) {
  const buySignal = getBuySignal(stockCode);
  const sellSignal = getSellSignal(stockCode, holdingData.purchasePrice);
  
  let finalRecommendation = '';
  let confidence = 0;
  
  // 保有している場合の判断
  if (holdingData.shares > 0) {
    if (sellSignal.score >= 4) {
      finalRecommendation = '即座に売却';
      confidence = 90;
    } else if (sellSignal.score >= 2) {
      finalRecommendation = '売却検討';
      confidence = 70;
    } else if (buySignal.score >= 4) {
      finalRecommendation = '買い増し検討';
      confidence = 80;
    } else {
      finalRecommendation = '保持';
      confidence = 60;
    }
  } 
  // 保有していない場合の判断
  else {
    if (buySignal.score >= 4) {
      finalRecommendation = '新規買い';
      confidence = 85;
    } else if (buySignal.score >= 2) {
      finalRecommendation = '買い検討';
      confidence = 65;
    } else {
      finalRecommendation = '見送り';
      confidence = 70;
    }
  }
  
  return {
    recommendation: finalRecommendation,
    confidence: confidence,
    buySignals: buySignal.signals,
    sellSignals: sellSignal.signals,
    totalScore: buySignal.score - sellSignal.score
  };
}
```

## スプレッドシート実装

### 判断アルゴリズム用関数

**シート「判断アルゴリズム」のG列（推奨アクション）:**
```javascript
=IF(AND(C2>D2,E2<=30,B2>C2),"強い買い",
  IF(AND(C2<D2,E2>=70,B2<C2),"強い売り",
    IF(OR(C2>D2,E2<=35),"買い検討",
      IF(OR(C2<D2,E2>=65),"売り検討","保持"))))
```

**H列（信頼度）:**
```javascript
=IF(G2="強い買い",90,
  IF(G2="強い売り",85,
    IF(OR(G2="買い検討",G2="売り検討"),70,60)))
```

### 詳細スコア計算

**I列（買いスコア）:**
```javascript
=IF(C2>D2,3,0) + IF(E2<=30,2,0) + IF(B2>C2,1,0) + IF(F2>1.5,1,0)
```

**J列（売りスコア）:**
```javascript
=IF(C2<D2,3,0) + IF(E2>=70,2,0) + IF(B2<C2,1,0) + IF((B2-取得価格)/取得価格<=-0.1,4,0)
```

## 実際の判断例

### 例1: 強い買いシグナル
```
トヨタ自動車（7203）
現在価格: 2,650円
MA5: 2,630円 > MA25: 2,600円 ✓ (+3点)
RSI: 25 ≤ 30 ✓ (+2点)
現在価格 > MA5 ✓ (+1点)
出来高: 平均の1.8倍 ✓ (+1点)
安値からの反発: 2,500円 → 2,650円 = +6% ✓ (+1点)
---
買いスコア: 8点 → 「強い買い」判定（信頼度90%）
```

### 例2: 強い売りシグナル
```
ソニー（6758）
現在価格: 10,800円, 取得価格: 12,000円
MA5: 10,850円 < MA25: 11,200円 ✓ (+3点)
損失率: (10,800-12,000)/12,000 = -10% ✓ (+4点)
RSI: 75 ≥ 70 ✓ (+2点)
現在価格 < MA5 ✓ (+1点)
---
売りスコア: 10点 → 「強い売り」判定（信頼度95%）
```

### 例3: 保持判定
```
任天堂（7974）
現在価格: 5,200円, 取得価格: 5,100円
MA5: 5,180円 > MA25: 5,150円 ✓ (+3点)
RSI: 55（中立）
損失率: +2%（利確・損切りライン未達）
---
買いスコア: 3点, 売りスコア: 0点 → 「保持」判定（信頼度60%）
```

## アラート自動化

### 高信頼度判断の自動通知

```javascript
function dailyAnalysisAndAlert() {
  const sheet = SpreadsheetApp.getActiveSpreadsheet().getSheetByName('ポートフォリオ');
  const lastRow = sheet.getLastRow();
  
  for (let i = 2; i <= lastRow; i++) {
    const stockCode = sheet.getRange(i, 1).getValue();
    const stockName = sheet.getRange(i, 2).getValue();
    const shares = sheet.getRange(i, 3).getValue();
    const purchasePrice = sheet.getRange(i, 4).getValue();
    
    const analysis = comprehensiveAnalysis(stockCode, {
      shares: shares,
      purchasePrice: purchasePrice
    });
    
    // 信頼度80%以上の判断のみ通知
    if (analysis.confidence >= 80) {
      let message = `🎯 投資判断アラート\n`;
      message += `${stockName}(${stockCode})\n`;
      message += `判断: ${analysis.recommendation}\n`;
      message += `信頼度: ${analysis.confidence}%\n`;
      message += `根拠: ${analysis.buySignals.concat(analysis.sellSignals).join(', ')}`;
      
      sendLineNotify(message);
    }
    
    // 結果をスプレッドシートに記録
    sheet.getRange(i, 10).setValue(analysis.recommendation);
    sheet.getRange(i, 11).setValue(analysis.confidence);
  }
}
```

## パフォーマンス指標

### 判断精度の測定
- **適中率**: 正しい判断の割合
- **偽陽性率**: 誤った買いシグナルの割合
- **偽陰性率**: 見逃した売りシグナルの割合

### 改善のためのバックテスト
```javascript
function backtestAlgorithm(stockCode, startDate, endDate) {
  // 過去データでアルゴリズムの精度を検証
  // 実際の価格変動と判断の整合性をチェック
  // パラメータ調整の参考データを作成
}
```

## 注意事項とリスク

### アルゴリズムの限界
1. **市場急変時**: 突発的なニュースや大きな市場変動への対応不可
2. **流動性リスク**: 出来高の少ない銘柄では機能しない可能性
3. **データ遅延**: リアルタイムデータでない場合の判断精度低下

### 推奨される使用方法
1. **補助ツールとして**: 最終判断は人間が行う
2. **複数指標の確認**: 単一の指標に依存しない
3. **定期的な見直し**: 市場環境変化に応じたパラメータ調整

この判断アルゴリズムにより、感情に左右されない客観的な投資判断が可能になります。