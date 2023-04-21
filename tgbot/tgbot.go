package tgbot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/momaee/telegram-bot/gpt"
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
		if update.Message != nil { // If we got a message
			if err := b.updateMessage(update); err != nil {
				// todo: handle update message error
				log.Println("failed to update message:", err)
			}
		}
	}

	return nil
}

func (b *Bot) updateMessage(update tgbotapi.Update) error {
	// If the message is reply to someone else, ignore it.
	if update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From != nil {
		if update.Message.ReplyToMessage.From.UserName != b.client.Self.UserName {
			return nil
		}
	}

	response, err := b.gpt.Chat(update.Message.Text)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())
		msg.ReplyToMessageID = update.Message.MessageID
		if _, err := b.client.Send(msg); err != nil {
			// todo: handle send message error
			log.Println("failed to send message:", err)
		}
		return nil
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
	msg.ReplyToMessageID = update.Message.MessageID
	if _, err := b.client.Send(msg); err != nil {
		// todo: handle send message error
		log.Println("failed to send message:", err)
	}

	return nil
}
