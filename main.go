package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_API_KEY"))
	if err != nil {
		fmt.Println("error new bot api:", err)
		return
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message

			if update.Message.From.UserName == "sirAlif" {
				if update.Message.From.LanguageCode == "en" {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Come eat it's head")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "بیا سرشو بخور")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
				}
				continue
			}

			// If the message is reply to someone else, ignore it.
			if update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From != nil {
				if update.Message.ReplyToMessage.From.UserName != bot.Self.UserName {
					continue
				}
			}

			response, err := call(update.Message.Text)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)
				continue
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
		}
	}
}

func loop() {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Enter your question:")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		input = strings.Replace(input, "\n", "", -1)

		if input == "exit" {
			break
		}

		output, err := call(input)
		if err != nil {
			fmt.Println("error:", err)
			return
		}

		fmt.Println("output:", output)
	}
}

func call(content string) (string, error) {
	// Set your OpenAI API key as an environment variable.
	apiKey := os.Getenv("OPENAI_API_KEY")

	type Messages struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	type Payload struct {
		Model    string     `json:"model"`
		Messages []Messages `json:"messages"`
	}

	data := Payload{
		Model: "gpt-3.5-turbo",
		Messages: []Messages{
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
		fmt.Println("error:", err)
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("error:", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("error:", resp.Status)
		return "", err
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

	var response Response
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		fmt.Println("error:", err)
		return "", err
	}

	return response.Choices[0].Message.Content, nil
}
