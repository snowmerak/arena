package gateway

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type claudeClient struct {
	config ModelConfig
	client *http.Client
}

func NewClaudeClient(cfg ModelConfig) LLMClient {
	return &claudeClient{
		config: cfg,
		client: &http.Client{},
	}
}

type claudeRequest struct {
	Model     string          `json:"model"`
	Messages  []claudeMessage `json:"messages"`
	System    string          `json:"system,omitempty"`
	MaxTokens int             `json:"max_tokens"`
	Stream    bool            `json:"stream"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (c *claudeClient) ChatCompletion(ctx context.Context, req *ChatRequest, w http.ResponseWriter) error {
	baseURL := c.config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/messages"

	var cReq claudeRequest
	cReq.Model = c.config.ActualModelName
	cReq.MaxTokens = 4096
	cReq.Stream = req.Stream

	for _, m := range req.Messages {
		if m.Role == "system" {
			cReq.System += m.Content + "\n"
		} else {
			// Claude expects 'assistant' instead of 'assistant' if it comes from user, but OpenAI uses 'assistant'.
			// Both OpenAI and Claude use 'user' and 'assistant'.
			role := m.Role
			if role == "tool" {
				role = "user" // simplificiation
			}
			cReq.Messages = append(cReq.Messages, claudeMessage{
				Role:    role,
				Content: m.Content,
			})
		}
	}

	bodyBytes, err := json.Marshal(cReq)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
		return nil
	}

	if req.Stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)
		return streamClaudeToOpenAI(resp.Body, w, req.Model)
	}

	// Non-streaming
	var cResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&cResp); err != nil {
		return err
	}

	var content string
	if cArr, ok := cResp["content"].([]interface{}); ok && len(cArr) > 0 {
		if cMap, ok := cArr[0].(map[string]interface{}); ok {
			if text, ok := cMap["text"].(string); ok {
				content = text
			}
		}
	}

	oResp := ChatResponse{
		ID:      "chatcmpl-claude",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []ChatChoice{
			{
				Index: 0,
				Message: &Message{
					Role:    "assistant",
					Content: content,
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(oResp)
}

func streamClaudeToOpenAI(body io.Reader, w http.ResponseWriter, model string) error {
	scanner := bufio.NewScanner(body)
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported by response writer")
	}

	var currentEvent string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "event: ") {
			currentEvent = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "data: ") {
			dataStr := strings.TrimPrefix(line, "data: ")

			if currentEvent == "content_block_delta" {
				var chunk map[string]interface{}
				if err := json.Unmarshal([]byte(dataStr), &chunk); err == nil {
					if delta, ok := chunk["delta"].(map[string]interface{}); ok {
						if text, ok := delta["text"].(string); ok {
							sendOpenAIChunk(w, flusher, model, text, nil)
						}
					}
				}
			} else if currentEvent == "message_stop" {
				reason := "stop"
				sendOpenAIChunk(w, flusher, model, "", &reason)
				fmt.Fprintf(w, "data: [DONE]\n\n")
				flusher.Flush()
			}
		}
	}
	return scanner.Err()
}

func sendOpenAIChunk(w io.Writer, flusher http.Flusher, model, content string, finishReason *string) {
	chunk := ChatResponse{
		ID:      "chatcmpl-stream",
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []ChatChoice{
			{
				Index: 0,
				Delta: &Message{
					Content: content,
				},
				FinishReason: finishReason,
			},
		},
	}
	b, _ := json.Marshal(chunk)
	fmt.Fprintf(w, "data: %s\n\n", string(b))
	if flusher != nil {
		flusher.Flush()
	}
}
