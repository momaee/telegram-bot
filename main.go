package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"log"
	"net/http"
	"os"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	telegramAPIKey = "projects/" + os.Getenv("PROJECT_ID") + "/secrets/TELEGRAM_API_KEY/versions/latest"
	openaiAPIKey   = "projects/" + os.Getenv("PROJECT_ID") + "/secrets/OPENAI_API_KEY/versions/latest"
)

func main() {
	// Create the client.
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Printf("failed to create secretmanager client: %v", err)
		return
	}
	defer client.Close()

	// Build the request.
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: telegramAPIKey,
	}

	// Call the API.
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		log.Printf("failed to access secret version: %v", err)
		return
	}

	// Verify the data checksum.
	crc32c := crc32.MakeTable(crc32.Castagnoli)
	checksum := int64(crc32.Checksum(result.Payload.Data, crc32c))
	if checksum != *result.Payload.DataCrc32C {
		log.Printf("Data corruption detected.")
		return
	}

	telegramAPIKey = string(result.Payload.Data)

	// Build the request.
	req = &secretmanagerpb.AccessSecretVersionRequest{
		Name: openaiAPIKey,
	}

	// Call the API.
	result, err = client.AccessSecretVersion(ctx, req)
	if err != nil {
		log.Printf("failed to access secret version: %v", err)
		return
	}

	// Verify the data checksum.
	crc32c = crc32.MakeTable(crc32.Castagnoli)
	checksum = int64(crc32.Checksum(result.Payload.Data, crc32c))
	if checksum != *result.Payload.DataCrc32C {
		log.Printf("Data corruption detected.")
		return
	}

	openaiAPIKey = string(result.Payload.Data)

	bot, err := tgbotapi.NewBotAPI(telegramAPIKey)
	if err != nil {
		log.Println("error creating bot:", err)
		return
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message

			if update.Message.Text == "do my trick to stop the bot" {
				break
			}

			if update.Message.From.UserName == "sirAlif" {
				if update.Message.From.LanguageCode == "en" {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You can eat it's head")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "تو بیا سرشو بخور")
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
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)
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
