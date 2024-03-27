package telegram

import (
	"fmt"
	"go.uber.org/zap"
	"time"

	"github.com/erwanlbp/trading-bot/pkg/log"
	"gopkg.in/telebot.v3"
)

type Client struct {
	client *telebot.Bot
	Logger *log.Logger
	Chat   *telebot.Chat
}

// Documentation : https://github.com/tucnak/telebot
func NewClient(l *log.Logger, token string, channelId int64) (*Client, error) {
	pref := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("failed to init telegram bot : %w", err)
	}

	chat, err := b.ChatByID(channelId)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat with channel id %d : %w", channelId, err)
	}

	client := Client{
		client: b,
		Logger: l,
		Chat:   chat,
	}

	return &client, nil
}

func (c *Client) StartBot() {
	go func() {
		c.client.Start()
		// TODO: how to close properly ?
	}()
}

func (c *Client) CreateHandler(endpoint interface{}, handlerFunc telebot.HandlerFunc) {
	c.client.Handle(endpoint, handlerFunc)
}

func (c *Client) Send(message string) {
	_, err := c.client.Send(c.Chat, message)
	if err != nil {
		c.Logger.Error(fmt.Sprintf("Error while sending message %s to telegram bot", message), zap.Error(err))
	}
}
