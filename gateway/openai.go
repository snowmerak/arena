package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type openaiClient struct {
	config ModelConfig
	client *http.Client
}

func NewOpenAIClient(cfg ModelConfig) LLMClient {
	return &openaiClient{
		config: cfg,
		client: &http.Client{},
	}
}

func (c *openaiClient) ChatCompletion(ctx context.Context, req *ChatRequest, w http.ResponseWriter) error {
	baseURL := c.config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/chat/completions"

	clonedReq := *req
	clonedReq.Model = c.config.ActualModelName

	bodyBytes, err := json.Marshal(clonedReq)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.config.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)

	if req.Stream {
		flusher, ok := w.(http.Flusher)
		if ok {
			buf := make([]byte, 4096)
			for {
				n, err := resp.Body.Read(buf)
				if n > 0 {
					w.Write(buf[:n])
					flusher.Flush()
				}
				if err != nil {
					if err != io.EOF {
						return err
					}
					break
				}
			}
			return nil
		}
	}

	_, err = io.Copy(w, resp.Body)
	return err
}
