package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Printf("usage: %s <prompt words>\n", os.Args[0])
		os.Exit(1)
	}
	keyName := "OPENAI_API_KEY"
	key := os.Getenv(keyName)
	if key == "" {
		fmt.Printf("required env var $%s is not set\n", keyName)
		os.Exit(1)
	}
	if err := chatgpt(key); err != nil {
		fmt.Printf("%s: %v\n", os.Args[0], err)
		os.Exit(1)
	}
}

func chatgpt(apiKey string) error {
	prompt := strings.Join(flag.Args(), " ")
	bod := map[string]any{
		"model":       "gpt-3.5-turbo",
		"messages":    []map[string]any{{"role": "user", "content": prompt}},
		"temperature": 0.7,
	}
	bs, err := json.Marshal(bod)
	if err != nil {
		return fmt.Errorf("marshalling body: %w", err)
	}
	r, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(bs))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	rawResp, err := http.DefaultClient.Do(r)
	if err != nil {
		return fmt.Errorf("doing request: %w", err)
	}
	respBytes, err := io.ReadAll(rawResp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}
	type Response struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int    `json:"created"`
		Model   string `json:"model"`
		Usage   struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
			Index        int    `json:"index"`
		} `json:"choices"`
	}
	var resp Response
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return fmt.Errorf("unmarshalling response: %w", err)
	}
	// {"id":"chatcmpl-6zX7Fzhzg5JtwE6tQ6vl5fKsstkbV","object":"chat.completion","created":1680123325,"model":"gpt-3.5-turbo-0301","usage":{"prompt_tokens":14,"completion_tokens":5,"total_tokens":19},"choices":[{"message":{"role":"assistant","content":"This is a test!"},"finish_reason":"stop","index":0}]}
	if len(resp.Choices) == 0 {
		return fmt.Errorf("no choices returned:\n\n%s\n", respBytes)
	}
	fmt.Printf("%s\n", resp.Choices[0].Message.Content)
	return nil
}
