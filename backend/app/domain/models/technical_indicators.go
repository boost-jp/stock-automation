// Code generated by SQLBoiler 4.19.5 (https://github.com/aarondl/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/types"
)

//go:generate go run  ../../../cmd/generator/repoinit --fields=ID,Code,Date,Rsi14,Macd,MacdSignal,MacdHistogram,Sma5,Sma25,Sma75,CreatedAt,UpdatedAt, TechnicalIndicator

// You can edit this as you like.

// TechnicalIndicator is an object representing the database table.
// Set the "validate" tags as needed.
// https://pkg.go.dev/gopkg.in/go-playground/validator.v10
type TechnicalIndicator struct {
	ID            string
	Code          string            // 銘柄コード
	Date          time.Time         // 計算日
	Rsi14         types.NullDecimal // RSI(14日)
	Macd          types.NullDecimal // MACD
	MacdSignal    types.NullDecimal // MACDシグナル
	MacdHistogram types.NullDecimal // MACDヒストグラム
	Sma5          types.NullDecimal // 5日移動平均
	Sma25         types.NullDecimal // 25日移動平均
	Sma75         types.NullDecimal // 75日移動平均
	CreatedAt     null.Time         // 作成日時
	UpdatedAt     null.Time         // 更新日時
}

func NewTechnicalIndicator(
	ID string,
	Code string,
	Date time.Time,
	Rsi14 types.NullDecimal,
	Macd types.NullDecimal,
	MacdSignal types.NullDecimal,
	MacdHistogram types.NullDecimal,
	Sma5 types.NullDecimal,
	Sma25 types.NullDecimal,
	Sma75 types.NullDecimal,
	CreatedAt null.Time,
	UpdatedAt null.Time,
) *TechnicalIndicator {
	do := &TechnicalIndicator{
		ID:            ID,
		Code:          Code,
		Date:          Date,
		Rsi14:         Rsi14,
		Macd:          Macd,
		MacdSignal:    MacdSignal,
		MacdHistogram: MacdHistogram,
		Sma5:          Sma5,
		Sma25:         Sma25,
		Sma75:         Sma75,
		CreatedAt:     CreatedAt,
		UpdatedAt:     UpdatedAt,
	}
	return do
}
