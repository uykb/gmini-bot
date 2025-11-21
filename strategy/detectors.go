
package strategy

import (
	"binance-monitor/models"
	"fmt"
	"math"
	"strconv"
	"time"
)

// DetectVolumeSignal 检测成交量异常信号
func DetectVolumeSignal(klines []models.KlineData) *models.Signal {
	const ZScoreThreshold = 2.0

	if len(klines) < 2 {
		return nil
	}

	volumes := make([]float64, len(klines))
	for i, k := range klines {
		volumes[i] = k.Volume
	}

	zScore := CalculateZScore(volumes)

	if math.Abs(zScore) > ZScoreThreshold {
		lastKline := klines[len(klines)-1]
		signal := &models.Signal{
			Symbol:      lastKline.Symbol,
			SignalType:  models.VolumeSignal,
			Timestamp:   time.Unix(0, lastKline.Timestamp*int64(time.Millisecond)),
			Description: fmt.Sprintf("成交量 Z-Score: %.2f (阈值: %.1f)", zScore, ZScoreThreshold),
			Meta: map[string]interface{}{
				"z_score":    zScore,
				"threshold":  ZScoreThreshold,
				"mean_volume": CalculateMean(volumes),
			},
		}
		return signal
	}

	return nil
}

// DetectOpenInterestSignal 检测持仓量异动信号
func DetectOpenInterestSignal(ois []models.BinanceOI) []*models.Signal {
	var signals []*models.Signal
	if len(ois) < 96 { // 需要足够的数据用于24小时比较
		return signals
	}

	lastOI := ois[len(ois)-1]
	lastOIFloat, _ := strconv.ParseFloat(lastOI.SumOpenInterest, 64)

	// 模式1: 24小时变化 > 10%
	oi24hAgo := ois[len(ois)-96]
	oi24hAgoFloat, _ := strconv.ParseFloat(oi24hAgo.SumOpenInterest, 64)
	if oi24hAgoFloat > 0 {
		change24h := (lastOIFloat - oi24hAgoFloat) / oi24hAgoFloat * 100
		if math.Abs(change24h) > 10.0 {
			signals = append(signals, &models.Signal{
				Symbol:      lastOI.Symbol,
				SignalType:  models.OpenInterestSignal,
				Timestamp:   time.Unix(0, lastOI.Timestamp*int64(time.Millisecond)),
				Description: fmt.Sprintf("24小时OI变化: %.2f%% (阈值: 10%%)", change24h),
				Meta: map[string]interface{}{"change_percent_24h": change24h},
			})
		}
	}

	// 模式2: 连续4个周期上涨/下跌
	if len(ois) >= 5 {
		consecutiveRises := 0
		consecutiveFalls := 0
		for i := len(ois) - 5; i < len(ois)-1; i++ {
			current, _ := strconv.ParseFloat(ois[i+1].SumOpenInterest, 64)
			prev, _ := strconv.ParseFloat(ois[i].SumOpenInterest, 64)
			if current > prev {
				consecutiveRises++
			} else if current < prev {
				consecutiveFalls++
			}
		}
		if consecutiveRises == 4 || consecutiveFalls == 4 {
			desc := fmt.Sprintf("OI连续4个周期上涨")
			if consecutiveFalls == 4 {
				desc = fmt.Sprintf("OI连续4个周期下跌")
			}
			signals = append(signals, &models.Signal{
				Symbol:      lastOI.Symbol,
				SignalType:  models.OpenInterestSignal,
				Timestamp:   time.Unix(0, lastOI.Timestamp*int64(time.Millisecond)),
				Description: desc,
				Meta: map[string]interface{}{"consecutive_periods": 4, "direction": map[bool]string{true: "rise", false: "fall"}[consecutiveRises == 4]},
			})
		}
	}

	// 模式3: 单周期剧烈变化 > 3.5%
	if len(ois) >= 2 {
		prevOI := ois[len(ois)-2]
		prevOIFloat, _ := strconv.ParseFloat(prevOI.SumOpenInterest, 64)
		if prevOIFloat > 0 {
			change1p := (lastOIFloat - prevOIFloat) / prevOIFloat * 100
			if math.Abs(change1p) > 3.5 {
				signals = append(signals, &models.Signal{
					Symbol:      lastOI.Symbol,
					SignalType:  models.OpenInterestSignal,
					Timestamp:   time.Unix(0, lastOI.Timestamp*int64(time.Millisecond)),
					Description: fmt.Sprintf("单周期OI剧烈变化: %.2f%% (阈值: 3.5%%)", change1p),
					Meta: map[string]interface{}{"change_percent_1p": change1p},
				})
			}
		}
	}

	return signals
}

// DetectLSRatioSignal 检测多空比极端信号
func DetectLSRatioSignal(lsRatios []models.GlobalLongShortRatio) *models.Signal {
	const ZScoreThreshold = 2.0

	if len(lsRatios) < 2 {
		return nil
	}

	ratios := make([]float64, len(lsRatios))
	for i, r := range lsRatios {
		ratios[i], _ = strconv.ParseFloat(r.LongShortRatio, 64)
	}

	zScore := CalculateZScore(ratios)

	if math.Abs(zScore) > ZScoreThreshold {
		lastRatio := lsRatios[len(lsRatios)-1]
		signal := &models.Signal{
			Symbol:      lastRatio.Symbol,
			SignalType:  models.LSRatioSignal,
			Timestamp:   time.Unix(0, lastRatio.Timestamp*int64(time.Millisecond)),
			Description: fmt.Sprintf("多空账户比 Z-Score: %.2f (阈值: %.1f), 市场情绪可能极端。", zScore, ZScoreThreshold),
			Meta: map[string]interface{}{
				"z_score":   zScore,
				"threshold": ZScoreThreshold,
				"ls_ratio":  ratios[len(ratios)-1],
			},
		}
		return signal
	}

	return nil
}
