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

type geminiClient struct {
	config ModelConfig
	client *http.Client
}

func NewGeminiClient(cfg ModelConfig) LLMClient {
	return &geminiClient{
		config: cfg,
		client: &http.Client{},
	}
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	SystemInstruction *geminiContent  `json:"systemInstruction,omitempty"`
	Contents          []geminiContent `json:"contents"`
}

func (c *geminiClient) ChatCompletion(ctx context.Context, req *ChatRequest, w http.ResponseWriter) error {
	baseURL := c.config.BaseURL
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1beta"
	}

	var gReq geminiRequest
	for _, m := range req.Messages {
		if m.Role == "system" {
			if gReq.SystemInstruction == nil {
				gReq.SystemInstruction = &geminiContent{Role: "user", Parts: []geminiPart{}}
			}
			gReq.SystemInstruction.Parts = append(gReq.SystemInstruction.Parts, geminiPart{Text: m.Content})
		} else {
			role := "user"
			if m.Role == "assistant" {
				role = "model"
			}
			gReq.Contents = append(gReq.Contents, geminiContent{
				Role:  role,
				Parts: []geminiPart{{Text: m.Content}},
			})
		}
	}

	bodyBytes, err := json.Marshal(gReq)
	if err != nil {
		return err
	}

	action := "generateContent"
	if req.Stream {
		action = "streamGenerateContent"
	}

	endpoint := fmt.Sprintf("%s/models/%s:%s?key=%s", baseURL, c.config.ActualModelName, action, c.config.APIKey)
	if req.Stream {
		endpoint += "&alt=sse"
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

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
		return streamGeminiToOpenAI(resp.Body, w, req.Model)
	}

	var gResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&gResp); err != nil {
		return err
	}

	var content string
	if cands, ok := gResp["candidates"].([]interface{}); ok && len(cands) > 0 {
		if cand, ok := cands[0].(map[string]interface{}); ok {
			if cnt, ok := cand["content"].(map[string]interface{}); ok {
				if parts, ok := cnt["parts"].([]interface{}); ok && len(parts) > 0 {
					if part, ok := parts[0].(map[string]interface{}); ok {
						if text, ok := part["text"].(string); ok {
							content = text
						}
					}
				}
			}
		}
	}

	oResp := ChatResponse{
		ID:      "chatcmpl-gemini",
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

func streamGeminiToOpenAI(body io.Reader, w http.ResponseWriter, model string) error {
	scanner := bufio.NewScanner(body)
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			dataStr := strings.TrimPrefix(line, "data: ")
			if dataStr == "[DONE]" {
				continue
			}

			var chunk map[string]interface{}
			if err := json.Unmarshal([]byte(dataStr), &chunk); err == nil {
				if cands, ok := chunk["candidates"].([]interface{}); ok && len(cands) > 0 {
					if cand, ok := cands[0].(map[string]interface{}); ok {
						
						var text string
						if cnt, ok := cand["content"].(map[string]interface{}); ok {
							if parts, ok := cnt["parts"].([]interface{}); ok && len(parts) > 0 {
								if part, ok := parts[0].(map[string]interface{}); ok {
									if t, ok := part["text"].(string); ok {
										text = t
									}
								}
							}
						}

						var finishReason *string
						if fr, ok := cand["finishReason"].(string); ok && fr != "" {
							reason := strings.ToLower(fr)
							finishReason = &reason
						}

						sendOpenAIChunk(w, flusher, model, text, finishReason)
					}
				}
			}
		}
	}
	
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
	return scanner.Err()
}
