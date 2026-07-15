package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OpenAICompatibleClient calls a Chat Completions-compatible API for JSON incident analyses.
type OpenAICompatibleClient struct {
	endpoint string
	apiKey   string
	model    string
	client   *http.Client
}

// NewOpenAICompatibleClient creates a completion client for an OpenAI-compatible base URL.
func NewOpenAICompatibleClient(baseURL, apiKey, model string) (*OpenAICompatibleClient, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("AI API key is required")
	}
	if strings.TrimSpace(model) == "" {
		return nil, fmt.Errorf("AI model is required")
	}
	base, err := url.Parse(baseURL)
	if err != nil || (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" {
		return nil, fmt.Errorf("AI base URL must be an absolute HTTP URL")
	}
	base.Path = strings.TrimSuffix(base.Path, "/") + "/chat/completions"
	base.RawQuery = ""
	base.Fragment = ""
	return &OpenAICompatibleClient{endpoint: base.String(), apiKey: apiKey, model: model, client: &http.Client{Timeout: 30 * time.Second}}, nil
}

// Complete submits prompts and returns the first completion message as JSON text.
func (c *OpenAICompatibleClient) Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	body, err := json.Marshal(chatRequest{
		Model:          c.model,
		Messages:       []chatMessage{{Role: "system", Content: systemPrompt}, {Role: "user", Content: userPrompt}},
		ResponseFormat: responseFormat{Type: "json_object"},
	})
	if err != nil {
		return "", fmt.Errorf("marshal AI request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create AI request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	response, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("call AI provider: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 8*1024))
		return "", fmt.Errorf("AI provider returned %s: %s", response.Status, strings.TrimSpace(string(message)))
	}

	var completion chatResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&completion); err != nil {
		return "", fmt.Errorf("decode AI response: %w", err)
	}
	if len(completion.Choices) == 0 || strings.TrimSpace(completion.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("AI response contained no completion content")
	}
	return completion.Choices[0].Message.Content, nil
}

type chatRequest struct {
	Model          string         `json:"model"`
	Messages       []chatMessage  `json:"messages"`
	ResponseFormat responseFormat `json:"response_format"`
}
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type responseFormat struct {
	Type string `json:"type"`
}
type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}
