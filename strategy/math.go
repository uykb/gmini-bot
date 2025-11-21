
package strategy

import (
	"math"
)

// CalculateMean 计算平均值
func CalculateMean(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

// CalculateStandardDeviation 计算标准差
func CalculateStandardDeviation(data []float64) float64 {
	if len(data) < 2 {
		return 0
	}
	mean := CalculateMean(data)
	sumOfSquares := 0.0
	for _, v := range data {
		sumOfSquares += math.Pow(v-mean, 2)
	}
	return math.Sqrt(sumOfSquares / float64(len(data)))
}

// CalculateZScore 计算最后一个数据点的Z-Score
func CalculateZScore(data []float64) float64 {
	if len(data) < 2 {
		return 0
	}
	mean := CalculateMean(data)
	stdDev := CalculateStandardDeviation(data)
	if stdDev == 0 {
		return 0 // 避免除以零
	}
	lastValue := data[len(data)-1]
	return (lastValue - mean) / stdDev
}

// CalculateEMA 计算指数移动平均线 (EMA)
func CalculateEMA(data []float64, period int) float64 {
	if len(data) < period {
		return 0
	}
	multiplier := 2.0 / (float64(period) + 1.0)
	// 从一个SMA开始
	sma := CalculateMean(data[:period])
	ema := sma
	for i := period; i < len(data); i++ {
		ema = (data[i]-ema)*multiplier + ema
	}
	return ema
}

// CalculateRSI 计算相对强弱指数 (RSI)
func CalculateRSI(data []float64, period int) float64 {
	if len(data) < period+1 {
		return 0
	}

	var gains, losses []float64
	for i := 1; i < len(data); i++ {
		change := data[i] - data[i-1]
		if change > 0 {
			gains = append(gains, change)
			losses = append(losses, 0)
		} else {
			gains = append(gains, 0)
			losses = append(losses, -change)
		}
	}

	// 计算第一个周期的平均增益和平均损失
	avgGain := CalculateMean(gains[:period])
	avgLoss := CalculateMean(losses[:period])

	// 平滑后续周期
	for i := period; i < len(gains); i++ {
		avgGain = (avgGain*float64(period-1) + gains[i]) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + losses[i]) / float64(period)
	}

	if avgLoss == 0 {
		return 100 // 极端看涨
	}

	rs := 100 - (100 / (1 + (avgGain / avgLoss)))
	return rs
}
