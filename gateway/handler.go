package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Gateway struct {
	models map[string]LLMClient
}

func NewGateway(cfg *Config) (*Gateway, error) {
	gw := &Gateway{
		models: make(map[string]LLMClient),
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

	if err := client.ChatCompletion(r.Context(), &req, w); err != nil {
		// If headers are already written, we can't send a 500 cleanly, 
		// but we can log it here.
		fmt.Printf("Error processing completion: %v\n", err)
	}
}
