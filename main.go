package main

import (
	"context"
	"log"
	"os"

	"github.com/momaee/telegram-bot-golang/gpt"
	sm "github.com/momaee/telegram-bot-golang/secrets-manager"
	"github.com/momaee/telegram-bot-golang/server"
	"github.com/momaee/telegram-bot-golang/tgbot"
)

var (
	gptAPIKeyName, telAPIKey string
	telAPIKeyName, gptAPIKey string
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if os.Getenv("SECRETS_PROJECT_ID") == "" {
		log.Println("failed to get SECRETS_PROJECT_ID")
		return
	}

	gptAPIKeyName = "projects/" + os.Getenv("SECRETS_PROJECT_ID") + "/secrets/OPENAI_API_KEY/versions/latest"

	if env := os.Getenv("ENV"); env == "dev" {
		telAPIKeyName = "projects/" + os.Getenv("SECRETS_PROJECT_ID") + "/secrets/TELEGRAM_API_KEY_DEV/versions/latest"
	} else if env == "prod" {
		telAPIKeyName = "projects/" + os.Getenv("SECRETS_PROJECT_ID") + "/secrets/TELEGRAM_API_KEY/versions/latest"
	} else {
		log.Println("Invalid ENV", env)
		return
	}
}

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
