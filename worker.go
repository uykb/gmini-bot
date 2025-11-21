
package main

import (
	"binance-monitor/cache"
	"binance-monitor/gemini"
	"binance-monitor/lark"
	"binance-monitor/models"
	"binance-monitor/strategy"
	"fmt"
	"os"
	"strings"
	"syscall/js"
)

func runCheck() {
	fmt.Println("开始执行检查...")

	// Get configuration from environment variables
	larkWebhookURL := os.Getenv("LARK_WEBHOOK_URL")
	symbolsStr := os.Getenv("SYMBOLS")
	apiKey := os.Getenv("API_KEY")
	aiEndpoint := os.Getenv("OPENAI_COMPATIBLE_ENDPOINT")
	aiModel := os.Getenv("AI_MODEL_NAME")
	kvBinding := "SIGNAL_CACHE" // The binding name from wrangler.toml

	if larkWebhookURL == "" || symbolsStr == "" {
		fmt.Println("错误: 缺少环境变量 LARK_WEBHOOK_URL 或 SYMBOLS")
		return
	}

	bot := lark.NewBot(larkWebhookURL)
	symbols := strings.Split(symbolsStr, ",")

	for _, symbol := range symbols {
		checkSymbol(symbol, bot, aiEndpoint, aiModel, apiKey, kvBinding)
	}

	fmt.Println("检查完成。")
}

func checkSymbol(symbol string, bot *lark.Bot, aiEndpoint, aiModel, apiKey, kvBinding string) {
	const LookbackPeriod = 96
	const DataInterval = "15m"
	const CacheTTL = 3600 // 1 hour in seconds

	fmt.Printf("正在为 %s 获取市场数据...\n", symbol)

	// Get KV Namespace
	kv, err := cache.GetKVNamespace(kvBinding)
	if err != nil {
		fmt.Printf("获取KV命名空间失败: %v。缓存功能将不可用。\n", err)
	}

	marketData, err := strategy.FetchMarketData(symbol, DataInterval, LookbackPeriod)
	if err != nil {
		fmt.Printf("获取 %s 的市场数据失败: %v\n", symbol, err)
		return
	}

	signals := strategy.Analyze(marketData)

	if len(signals) > 0 {
		fmt.Printf("为 %s 发现 %d 个信号:\n", symbol, len(signals))
		contextData := strategy.BuildContextData(marketData)

		for _, signal := range signals {
			// Check cache before sending notification
			cacheKey := fmt.Sprintf("%s:%s", signal.Symbol, signal.SignalType)
			if !kv.IsUndefined() {
				exists, _ := cache.KeyExists(kv, cacheKey)
				if exists {
					fmt.Printf("信号 '%s' 在一小时内已发送过，跳过。\n", cacheKey)
					continue // Skip to the next signal
				}
			}

			fmt.Printf("  - 信号: %s, 描述: %s\n", signal.SignalType, signal.Description)

			if aiEndpoint != "" && aiModel != "" && apiKey != "" {
				analysis, err := gemini.GetAIAnalysis(aiEndpoint, aiModel, apiKey, signal, contextData)
				if err != nil {
					fmt.Printf("AI API 分析失败: %v\n", err)
				} else {
					signal.GeminiAnalysis = analysis
				}
			}

			err := bot.SendSignal(signal)
			if err != nil {
				fmt.Printf("发送飞书消息失败: %v\n", err)
			} else {
				// Cache the signal upon successful sending
				if !kv.IsUndefined() {
					cache.SetKey(kv, cacheKey, CacheTTL)
					fmt.Printf("信号 '%s' 已缓存，有效期 %d 秒。\n", cacheKey, CacheTTL)
				}
			}
		}
	} else {
		fmt.Printf("未发现 %s 的交易信号。\n", symbol)
	}
}

func main() {
	c := make(chan bool)
	js.Global().Set("run", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go runCheck()
		return nil
	}))
	fmt.Println("Go Wasm initialized. Ready to be called from JS.")
	<-c
}
