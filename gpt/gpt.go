package gpt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type GPT struct {
	apiKey string
}

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Response struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int      `json:"created"`
	Model   string   `json:"model"`
	Usage   Usage    `json:"usage"`
	Choices []Choice `json:"choices"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
	Index        int     `json:"index"`
}

func New(apiKey string) *GPT {
	return &GPT{
		apiKey: apiKey,
	}
}

func (c *GPT) Chat(content string) (string, error) {
	data := Request{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{
				Role:    "user",
				Content: content,
			},
		},
	}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", body)
	if err != nil {
		log.Printf("failed to create request: %v", err)
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("failed to send http request: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("received wrong status code: %v", resp.Status)
		return "", fmt.Errorf("received wrong status code: %v", resp.Status)
	}

	var response Response
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Printf("failed to decode response: %v", err)
		return "", err
	}

	if len(response.Choices) == 0 {
		log.Printf("no response from openai")
		return "", fmt.Errorf("no response from openai")
	}

	return response.Choices[0].Message.Content, nil
}
