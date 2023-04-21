package main

import (
	"context"
	"log"
	"os"

	"github.com/momaee/telegram-bot/gpt"
	sm "github.com/momaee/telegram-bot/secrets-manager"
	"github.com/momaee/telegram-bot/server"
	"github.com/momaee/telegram-bot/tgbot"
)

var (
	telAPIKeyName = "projects/" + os.Getenv("SECRETS_PROJECT_ID") + "/secrets/TELEGRAM_API_KEY/versions/latest"
	gptAPIKeyName = "projects/" + os.Getenv("SECRETS_PROJECT_ID") + "/secrets/OPENAI_API_KEY/versions/latest"
	telAPIKey     string
	gptAPIKey     string
)

func main() {
	if err := getSecrets(context.Background()); err != nil {
		log.Println("failed to get secrets:", err)
		return
	}

	gpt := gpt.New(gptAPIKey)

	tgbot, err := tgbot.New(telAPIKey, gpt)
	if err != nil {
		log.Println("failed to create telegram bot:", err)
		return
	}

	go func() {
		if err := tgbot.Start(); err != nil {
			log.Println("failed to start telegram bot:", err)
		}
	}()

	server := server.New()
	if err := server.Start(); err != nil {
		log.Println("failed to start server:", err)
	}
}

func getSecrets(ctx context.Context) error {
	secMang, err := sm.New(ctx)
	if err != nil {
		log.Printf("failed to create secrets manager: %v", err)
		return err
	}
	defer secMang.Close()

	telAPIKey, err = secMang.GetSecret(ctx, telAPIKeyName)
	if err != nil {
		log.Printf("failed to get telegram api key: %v", err)
		return err
	}

	gptAPIKey, err = secMang.GetSecret(ctx, gptAPIKeyName)
	if err != nil {
		log.Printf("failed to get openai api key: %v", err)
		return err
	}

	return nil
}
