
package lark

import (
	"binance-monitor/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Bot encapsulates the logic for interacting with a Lark bot.
type Bot struct {
	webhookURL string
}

// NewBot creates a new Bot instance.
func NewBot(webhookURL string) *Bot {
	return &Bot{webhookURL: webhookURL}
}

// SendSignal sends a formatted trading signal to Lark.
func (b *Bot) SendSignal(signal models.Signal) error {
	cardContent, err := formatSignalToLarkCard(signal)
	if err != nil {
		return fmt.Errorf("failed to format Lark card: %w", err)
	}

	resp, err := http.Post(b.webhookURL, "application/json", bytes.NewBuffer(cardContent))
	if err != nil {
		return fmt.Errorf("failed to send Lark message request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 status after sending Lark message: %d", resp.StatusCode)
	}

	fmt.Printf("Successfully sent signal to Lark: %s for %s\n", signal.SignalType, signal.Symbol)
	return nil
}

// --- Lark Card Message Formatting ---

type LarkCard struct {
	MsgType string  `json:"msg_type"`
	Card    CardDef `json:"card"`
}

type CardDef struct {
	Header   HeaderDef    `json:"header"`
	Elements []interface{} `json:"elements"`
}

type HeaderDef struct {
	Title    TextDef `json:"title"`
	Template string  `json:"template"`
}

type TextDef struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

type DivDef struct {
	Tag    string     `json:"tag"`
	Text   *TextDef   `json:"text,omitempty"`
	Fields []FieldDef `json:"fields,omitempty"`
}

type FieldDef struct {
	IsShort bool    `json:"is_short"`
	Text    TextDef `json:"text"`
}

type HrDef struct {
	Tag string `json:"tag"`
}

type NoteDef struct {
	Tag      string    `json:"tag"`
	Elements []TextDef `json:"elements"`
}

func formatSignalToLarkCard(signal models.Signal) ([]byte, error) {
	cardColor := "blue" // Default color
	switch signal.SignalType {
	case models.VolumeSignal, models.OpenInterestSignal:
		cardColor = "orange"
	case models.LSRatioSignal:
		cardColor = "purple"
	}

	elements := []interface{}{
		DivDef{
			Tag:  "div",
			Text: &TextDef{Tag: "lark_md", Content: signal.Description},
		},
		HrDef{Tag: "hr"},
	}

	if signal.GeminiAnalysis != "" {
		// Sanitize Gemini analysis for Lark Markdown
		formattedAnalysis := strings.ReplaceAll(signal.GeminiAnalysis, "【", "**【")
		formattedAnalysis = strings.ReplaceAll(formattedAnalysis, "】", "】**")

		elements = append(elements, DivDef{
			Tag: "div",
			Text: &TextDef{
				Tag:     "lark_md",
				Content: "**🤖 AI 智能分析**\n" + formattedAnalysis,
			},
		}, HrDef{Tag: "hr"})
	}

	elements = append(elements, NoteDef{
		Tag: "note",
		Elements: []TextDef{
			{Tag: "plain_text", Content: fmt.Sprintf("时间: %s", signal.Timestamp.In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05 CST"))},
		},
	})

	card := LarkCard{
		MsgType: "interactive",
		Card: CardDef{
			Header: HeaderDef{
				Title: TextDef{
					Tag:     "plain_text",
					Content: fmt.Sprintf("📈 %s 交易信号: %s", signal.Symbol, signal.SignalType),
				},
				Template: cardColor,
			},
			Elements: elements,
		},
	}

	return json.Marshal(card)
}
