package tgbot

import (
	"errors"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/momaee/telegram-bot-golang/gpt"
)

var (
	ErrSendMessage = fmt.Errorf("bot: failed to send message")
)

type Bot struct {
	apiKey string
	client *tgbotapi.BotAPI
	gpt    *gpt.GPT
}

func New(apiKey string, gpt *gpt.GPT) (*Bot, error) {
	client, err := tgbotapi.NewBotAPI(apiKey)
	if err != nil {
		return nil, err
	}

	client.Debug = true

	log.Printf("authorized on account %s", client.Self.UserName)

	return &Bot{
		apiKey: apiKey,
		client: client,
		gpt:    gpt,
	}, nil
}

func (b *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.client.GetUpdatesChan(u)

	for update := range updates {
		// If we got a message, process it.
		if update.Message != nil {
			if err := b.updateMessage(update); err != nil {
				log.Println("failed to update message:", err)
				if errors.Is(err, ErrSendMessage) {
					continue
				}
				// todo: think about this.
				return err
			}
		}
	}

	return nil
}

func (b *Bot) updateMessage(update tgbotapi.Update) error {
	asists := []string{}
	if update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From != nil {
		// If the message is reply to someone else, ignore it.
		if update.Message.ReplyToMessage.From.UserName != b.client.Self.UserName {
			return nil
		}

		asists = append(asists, update.Message.ReplyToMessage.Text)
	}

	response := ""

	if res, err := b.gpt.Chat(update.Message.Text, asists...); err != nil {
		log.Println("failed to get response from GPT:", err)
		response = "Failed to get response from GPT"
	} else {
		response = res
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
	msg.ReplyToMessageID = update.Message.MessageID

	if _, err := b.client.Send(msg); err != nil {
		log.Println("failed to send message:", err)
		return fmt.Errorf("%w: %v", ErrSendMessage, err)
	}

	return nil
}
