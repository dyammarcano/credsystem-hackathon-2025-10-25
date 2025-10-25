package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

var bufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}

type (
	OpenRouterRequest struct {
		Model    string    `json:"model"`
		Messages []Message `json:"messages"`
	}

	OpenRouterResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	DataResponse struct {
		ServiceID   uint8  `json:"service_id"`
		ServiceName string `json:"service_name"`
	}

	ContextPrompt struct {
		Prompt   string    `json:"prompt"`
		Model    string    `json:"model"`
		Messages []Message `json:"messages"`
	}

	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
)

func (c *Client) ChatCompletion(ctx context.Context, request OpenRouterRequest) (*DataResponse, error) {
	url := c.baseURL + "/chat/completions"

	// Encode request using a pooled buffer to reduce allocations
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	enc := json.NewEncoder(buf)
	if err := enc.Encode(request); err != nil {
		buf.Reset()
		bufPool.Put(buf)
		return nil, fmt.Errorf("error encoding request: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(buf.Bytes()))
	if err != nil {
		buf.Reset()
		bufPool.Put(buf)
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(ctx, req)
	// We can safely return the buffer after request is created; to be conservative, return it after Do completes
	buf.Reset()
	bufPool.Put(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var openRouterResp OpenRouterResponse
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&openRouterResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	if len(openRouterResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	var dataRes DataResponse
	if err := json.NewDecoder(strings.NewReader(openRouterResp.Choices[0].Message.Content)).Decode(&dataRes); err != nil {
		return nil, fmt.Errorf("error decoding data response: %v. content: %s", err, openRouterResp.Choices[0].Message.Content)
	}

	return &dataRes, nil
}
