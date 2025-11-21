
package models

import "time"

// --- Binance API Data Structures ---

// BinanceKline 代表从币安API获取的单条K线原始数据
type BinanceKline [12]interface{}

// BinanceOI 代表从币安API获取的持仓量数据
type BinanceOI struct {
	Symbol          string `json:"symbol"`
	SumOpenInterest string `json:"sumOpenInterest"`
	Timestamp       int64  `json:"timestamp"`
}

// GlobalLongShortRatio 代表从币安API获取的多空账户比数据
type GlobalLongShortRatio struct {
	Symbol         string `json:"symbol"`
	LongShortRatio string `json:"longShortRatio"`
	LongAccount    string `json:"longAccount"`
	ShortAccount   string `json:"shortAccount"`
	Timestamp      int64  `json:"timestamp"`
}


// --- Internal Data Structures ---

// SignalType 定义了交易信号的类型
type SignalType string

const (
	// New Signals based on the new strategy
	VolumeSignal     SignalType = "成交量异常"
	OpenInterestSignal SignalType = "持仓量异动"
	LSRatioSignal    SignalType = "多空比极端"
	CompositeSignal  SignalType = "复合信号"
)

// KlineData 代表内部使用的、格式化后的单条K线数据
type KlineData struct {
	Symbol    string
	Timestamp int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

// Signal 代表一个分析后得出的、准备发送的信号
type Signal struct {
	Symbol         string                 `json:"symbol"`
	SignalType     SignalType             `json:"signal_type"`
	Timestamp      time.Time              `json:"timestamp"`
	Description    string                 `json:"description"` // 简要描述，例如 "成交量 Z-Score > 2.0"
	Meta           map[string]interface{} `json:"meta"`        // 存储信号相关的元数据，如Z-Score值, 变化率等
	GeminiAnalysis string                 `json:"gemini_analysis,omitempty"` // Gemini的分析结果
}
