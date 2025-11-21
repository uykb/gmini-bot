
// Package gemini provides a client for communicating with an OpenAI-compatible API.
package gemini

import (
	"binance-monitor/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	apiTimeout = 45 * time.Second
)

// --- OpenAI-Compatible API Structures ---

// OpenAIRequest defines the structure for a chat completion request.
type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	// Temperature float32   `json:"temperature,omitempty"`
}

// Message represents a single message in the chat history.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse defines the structure for a chat completion response.
type OpenAIResponse struct {
	Choices []Choice `json:"choices"`
}

// Choice contains the generated message.	ype Choice struct {
	Message Message `json:"message"`
}

// GetAIAnalysis sends the detected signal and market context to an OpenAI-compatible API.
func GetAIAnalysis(endpoint, modelName, apiKey string, signal models.Signal, contextData string) (string, error) {
	if endpoint == "" || apiKey == "" || modelName == "" {
		return "", fmt.Errorf("AI endpoint, model name, or API key is not set")
	}

	// 1. Construct the prompt using the messages format
	systemPrompt := `You are a professional crypto market analyst. Your task is to provide a concise and insightful analysis in Chinese based on the data provided. Your entire response must follow this three-section format strictly: "【核心信号】", "【市场背景】", and "【潜在影响】". Be concise and straight to the point.`
	userPrompt := fmt.Sprintf("A trading signal was detected for %s.\n\n**Detected Signal:**\n- Signal Type: %s\n- Description: %s\n\n**Market Context Data:**\n%s\n\nNow, please provide your analysis based on the instructions.", signal.Symbol, signal.SignalType, signal.Description, contextData)

	// 2. Create the request payload
	reqPayload := OpenAIRequest{
		Model: modelName,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	jsonData, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal openai request: %w", err)
	}

	// 3. Send the HTTP request
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create openai request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to openai-compatible endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai-compatible API returned non-200 status: %d, body: %s", resp.StatusCode, string(tbody))
	}

	// 4. Parse the response
	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return "", fmt.Errorf("failed to decode openai response: %w", err)
	}

	if len(openAIResp.Choices) > 0 {
		return openAIResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("received empty response from OpenAI-compatible API")
}
