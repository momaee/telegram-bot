package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"log"
	"net/http"
	"os"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	telAPIKeyName = "projects/" + os.Getenv("SECRETS_PROJECT_ID") + "/secrets/TELEGRAM_API_KEY/versions/latest"
	gptAPIKeyName = "projects/" + os.Getenv("SECRETS_PROJECT_ID") + "/secrets/OPENAI_API_KEY/versions/latest"
	telAPIKey     string
	gptAPIKey     string
)

func main() {
	if err := getSecrets(context.Background()); err != nil {
		log.Println("error getting secrets:", err)
		return
	}

	go telBot()

	server()
}

func server() {
	http.HandleFunc("/health", indexHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/health" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprint(w, "Healthy")
}

func getSecrets(ctx context.Context) error {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Printf("failed to create secretmanager client: %v", err)
		return err
	}
	defer client.Close()

	telAPIKey, err = getSecret(ctx, client, telAPIKeyName)
	if err != nil {
		log.Printf("failed to get telegram api key: %v", err)
		return err
	}

	gptAPIKey, err = getSecret(ctx, client, gptAPIKeyName)
	if err != nil {
		log.Printf("failed to get openai api key: %v", err)
		return err
	}

	return nil
}

func getSecret(ctx context.Context, client *secretmanager.Client, name string) (string, error) {
	// Build the request.
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	// Call the API.
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", err
	}

	// Verify the data checksum.
	crc32c := crc32.MakeTable(crc32.Castagnoli)
	checksum := int64(crc32.Checksum(result.Payload.Data, crc32c))
	if checksum != *result.Payload.DataCrc32C {
		return "", fmt.Errorf("Data corruption detected.")
	}

	return string(result.Payload.Data), nil
}

// telBot is a Telegram bot that uses GPT-3 to reply to messages.
// This function is endless, unless there is an error.
func telBot() {
	bot, err := tgbotapi.NewBotAPI(telAPIKey)
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
			if err := updateMessage(bot, update); err != nil {
				// todo: handle update message error
				log.Println("error updating message:", err)
			}
		}
	}
}

func updateMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	// If the message is reply to someone else, ignore it.
	if update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From != nil {
		if update.Message.ReplyToMessage.From.UserName != bot.Self.UserName {
			return nil
		}
	}

	response, err := chatGPT(update.Message.Text)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())
		msg.ReplyToMessageID = update.Message.MessageID
		if _, err := bot.Send(msg); err != nil {
			// todo: handle send message error
			log.Println("error sending message:", err)
		}
		return nil
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
	msg.ReplyToMessageID = update.Message.MessageID
	if _, err := bot.Send(msg); err != nil {
		// todo: handle send message error
		log.Println("error sending message:", err)
	}

	return nil
}

func chatGPT(content string) (string, error) {
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
		log.Printf("error creating request: %v", err)
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+gptAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("error sending request: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("error response: %v", resp.Status)
		return "", fmt.Errorf("error response: %v", resp.Status)
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
		log.Printf("error decoding response: %v", err)
		return "", err
	}

	if len(response.Choices) == 0 {
		log.Printf("no response")
		return "", fmt.Errorf("no response")
	}

	return response.Choices[0].Message.Content, nil
}
