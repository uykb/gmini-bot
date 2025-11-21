
package strategy

import (
	"binance-monitor/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"io"
)

// MarketData 包含用于分析的所有市场数据
type MarketData struct {
	Symbol   string
	Klines   []models.KlineData
	OIs      []models.BinanceOI
	LSRatios []models.GlobalLongShortRatio
}

// Analyze 是策略分析的主入口函数
func Analyze(data MarketData) []models.Signal {
	var signals []models.Signal

	// 1. 检测成交量异常信号
	if volSignal := DetectVolumeSignal(data.Klines); volSignal != nil {
		signals = append(signals, *volSignal)
	}

	// 2. 检测持仓量异动信号
	oiSignals := DetectOpenInterestSignal(data.OIs)
	if len(oiSignals) > 0 {
		for _, s := range oiSignals {
			signals = append(signals, *s)
		}
	}

	// 3. 检测多空比极端信号
	if lsRatioSignal := DetectLSRatioSignal(data.LSRatios); lsRatioSignal != nil {
		signals = append(signals, *lsRatioSignal)
	}

	return signals
}

// BuildContextData 为AI API构建丰富的市场上下文
func BuildContextData(data MarketData) string {
	var sb strings.Builder

	closePrices := make([]float64, len(data.Klines))
	for i, k := range data.Klines {
		closePrices[i] = k.Close
	}

	rsi14 := CalculateRSI(closePrices, 14)
	ema12 := CalculateEMA(closePrices, 12)
	ema26 := CalculateEMA(closePrices, 26)

	sb.WriteString(fmt.Sprintf("### 关键指标摘要
"))
	sb.WriteString(fmt.Sprintf("- **最新收盘价:** %.4f
", data.Klines[len(data.Klines)-1].Close))
	sb.WriteString(fmt.Sprintf("- **RSI (14):** %.2f
", rsi14))
	sb.WriteString(fmt.Sprintf("- **EMA (12/26):** %.4f / %.4f
", ema12, ema26))
	lastOI, _ := strconv.ParseFloat(data.OIs[len(data.OIs)-1].SumOpenInterest, 64)
	sb.WriteString(fmt.Sprintf("- **最新持仓量 (OI):** %.2f
", lastOI))
	lastLSR, _ := strconv.ParseFloat(data.LSRatios[len(data.LSRatios)-1].LongShortRatio, 64)
	sb.WriteString(fmt.Sprintf("- **最新多空比:** %.4f

", lastLSR))

	sb.WriteString("### 最近K线 (OHLCV)
")
	start := len(data.Klines) - 5
	if start < 0 {
		start = 0
	}
	for i := start; i < len(data.Klines); i++ {
		k := data.Klines[i]
		sb.WriteString(fmt.Sprintf("  - T: %d, O: %.2f, H: %.2f, L: %.2f, C: %.2f, V: %.2f
", k.Timestamp, k.Open, k.High, k.Low, k.Close, k.Volume))
	}

	return sb.String()
}

// --- Data Fetching ---

// FetchMarketData fetches all required market data for a symbol.
func FetchMarketData(symbol, interval string, limit int) (MarketData, error) {
	var data MarketData
	data.Symbol = symbol

	klines, err := getKlines(symbol, interval, limit)
	if err != nil {
		return data, fmt.Errorf("failed to get klines: %w", err)
	}
	data.Klines = klines

	ois, err := getOpenInterest(symbol, interval, limit)
	if err != nil {
		return data, fmt.Errorf("failed to get open interest: %w", err)
	}
	data.OIs = ois

	lsRatios, err := getGlobalLongShortAccountRatio(symbol, interval, limit)
	if err != nil {
		return data, fmt.Errorf("failed to get long/short ratio: %w", err)
	}
	data.LSRatios = lsRatios

	return data, nil
}

func getKlines(symbol, interval string, limit int) ([]models.KlineData, error) {
	// ... implementation remains the same ...
	url := fmt.Sprintf("https://fapi.binance.com/fapi/v1/klines?symbol=%s&interval=%s&limit=%d", symbol, interval, limit)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rawKlines []models.BinanceKline
	if err := json.Unmarshal(body, &rawKlines); err != nil {
		return nil, err
	}

	var klines []models.KlineData
	for _, k := range rawKlines {
		open, _ := strconv.ParseFloat(k[1].(string), 64)
		high, _ := strconv.ParseFloat(k[2].(string), 64)
		low, _ := strconv.ParseFloat(k[3].(string), 64)
		close, _ := strconv.ParseFloat(k[4].(string), 64)
		volume, _ := strconv.ParseFloat(k[5].(string), 64)

		klines = append(klines, models.KlineData{
			Symbol:    symbol,
			Timestamp: int64(k[0].(float64)),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
		})
	}
	return klines, nil
}

func getOpenInterest(symbol, period string, limit int) ([]models.BinanceOI, error) {
	// ... implementation remains the same ...
	url := fmt.Sprintf("https://fapi.binance.com/futures/data/openInterestHist?symbol=%s&period=%s&limit=%d", symbol, period, limit)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ois []models.BinanceOI
	if err := json.Unmarshal(body, &ois); err != nil {
		return nil, fmt.Errorf("json unmarshal error: %w, body: %s", err, string(body))
	}
	return ois, nil
}

func getGlobalLongShortAccountRatio(symbol, period string, limit int) ([]models.GlobalLongShortRatio, error) {
	// ... implementation remains the same ...
	url := fmt.Sprintf("https://fapi.binance.com/futures/data/globalLongShortAccountRatio?symbol=%s&period=%s&limit=%d", symbol, period, limit)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ratios []models.GlobalLongShortRatio
	if err := json.Unmarshal(body, &ratios); err != nil {
		return nil, fmt.Errorf("json unmarshal error: %w, body: %s", err, string(body))
	}
	return ratios, nil
}
