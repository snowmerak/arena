package gateway

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Gateway struct {
	models map[string]LLMClient
	cache  *SemanticCache
	cfg    *Config
}

func NewGateway(cfg *Config) (*Gateway, error) {
	gw := &Gateway{
		models: make(map[string]LLMClient),
		cfg:    cfg,
	}

	for _, m := range cfg.Models {
		switch m.Type {
		case ProviderOpenAI:
			gw.models[m.Name] = NewOpenAIClient(m)
		case ProviderClaude:
			gw.models[m.Name] = NewClaudeClient(m)
		case ProviderGemini:
			gw.models[m.Name] = NewGeminiClient(m)
		default:
			return nil, fmt.Errorf("unknown provider type: %s", m.Type)
		}
	}

	if cfg.Redis.URL != "" {
		cache, err := NewSemanticCache(cfg.Redis.URL, &cfg.Embedding)
		if err != nil {
			return nil, err
		}
		gw.cache = cache
	}

	return gw, nil
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path != "/v1/chat/completions" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	client, ok := g.models[req.Model]
	if !ok {
		http.Error(w, fmt.Sprintf("model %s not found", req.Model), http.StatusNotFound)
		return
	}

	var prompt string
	for _, m := range req.Messages {
		prompt += m.Content + "\n"
	}
	prompt = strings.TrimSpace(prompt)

	var vec []float32
	var cacheHit string
	var exactHash string

	if g.cache != nil && prompt != "" {
		// Layer 1: Deterministic Exact Hash Evaluation
		exactHash = fmt.Sprintf("%x", sha256.Sum256([]byte(prompt)))
		hit, _ := g.cache.SearchHash(r.Context(), exactHash)
		
		if hit != "" {
			cacheHit = hit
		} else if len(prompt) <= 100000 {
			// Layer 2: Semantic Cache Evaluation (Bounded to 100,000 bytes)
			vec, _ = GetEmbedding(r.Context(), &g.cfg.Embedding, prompt)
			if len(vec) > 0 {
				hitV, _ := g.cache.Search(r.Context(), vec)
				if hitV != "" {
					cacheHit = hitV
				}
			}
		}
	}

	if cacheHit != "" {
		w.Header().Set("x-cache", "HIT")
		if req.Stream {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")
			w.WriteHeader(http.StatusOK)

			flusher, ok := w.(http.Flusher)
			
			chunks := strings.Split(cacheHit, " ")
			for i, word := range chunks {
				text := word
				if i < len(chunks)-1 {
					text += " "
				}
				
				respChunk := ChatResponse{
					ID:      "chatcmpl-stream-cached",
					Object:  "chat.completion.chunk",
					Created: time.Now().Unix(),
					Model:   req.Model,
					Choices: []ChatChoice{
						{
							Index: 0,
							Delta: &Message{
								Content: text,
							},
						},
					},
				}
				b, _ := json.Marshal(respChunk)
				fmt.Fprintf(w, "data: %s\n\n", string(b))
				if ok {
					flusher.Flush()
				}
				time.Sleep(20 * time.Millisecond)
			}

			reason := "stop"
			endChunk := ChatResponse{
				ID:      "chatcmpl-stream-cached",
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   req.Model,
				Choices: []ChatChoice{
					{
						Index: 0,
						Delta: &Message{},
						FinishReason: &reason,
					},
				},
			}
			b, _ := json.Marshal(endChunk)
			fmt.Fprintf(w, "data: %s\n\n", string(b))
			fmt.Fprintf(w, "data: [DONE]\n\n")
			if ok {
				flusher.Flush()
			}
			return
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			oResp := ChatResponse{
				ID:      "chatcmpl-cached",
				Object:  "chat.completion",
				Created: time.Now().Unix(),
				Model:   req.Model,
				Choices: []ChatChoice{
					{
						Index: 0,
						Message: &Message{
							Role:    "assistant",
							Content: cacheHit,
						},
					},
				},
			}
			json.NewEncoder(w).Encode(oResp)
			return
		}
	}

	w.Header().Set("x-cache", "MISS")
	recorder := newResponseRecorder(w)

	if err := client.ChatCompletion(r.Context(), &req, recorder); err != nil {
		fmt.Printf("Error processing completion: %v\n", err)
		return
	}

	if g.cache != nil && exactHash != "" && recorder.statusCode == http.StatusOK {
		go func() {
			extractedText := extractTextFromResponse(recorder.body.Bytes(), req.Stream)
			if extractedText != "" {
				_ = g.cache.StoreHybrid(context.Background(), exactHash, vec, extractedText)
			}
		}()
	}
}

func extractTextFromResponse(body []byte, isStream bool) string {
	if isStream {
		var full string
		lines := strings.Split(string(body), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "data: ") {
				dataStr := strings.TrimPrefix(line, "data: ")
				if dataStr == "[DONE]" {
					continue
				}
				var chunk ChatResponse
				if err := json.Unmarshal([]byte(dataStr), &chunk); err == nil {
					if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
						full += chunk.Choices[0].Delta.Content
					}
				}
			}
		}
		return full
	}

	var resp ChatResponse
	if err := json.Unmarshal(body, &resp); err == nil {
		if len(resp.Choices) > 0 && resp.Choices[0].Message != nil {
			return resp.Choices[0].Message.Content
		}
	}
	return ""
}
