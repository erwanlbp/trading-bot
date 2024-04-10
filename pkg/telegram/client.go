package telegram

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/erwanlbp/trading-bot/pkg/config/configfile"

	"go.uber.org/zap"
	"gopkg.in/telebot.v3"

	"github.com/erwanlbp/trading-bot/pkg/log"
)

type Client struct {
	client *telebot.Bot
	Logger *log.Logger
	Chat   *telebot.Chat

	queueCh chan string
}

// Documentation : https://github.com/tucnak/telebot
func NewClient(ctx context.Context, l *log.Logger, cf *configfile.ConfigFile) (*Client, error) {
	pref := telebot.Settings{
		Token:  cf.Telegram.Token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("failed to init telegram bot : %w", err)
	}

	chat, err := b.ChatByID(cf.Telegram.ChannelId)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat with channel id %d : %w", cf.Telegram.ChannelId, err)
	}

	client := Client{
		client:  b,
		Logger:  l,
		Chat:    chat,
		queueCh: make(chan string, 1000),
	}

	go client.queueHandler(ctx)

	return &client, nil
}

func (c *Client) Send(message string) {
	if c.queueCh == nil {
		return
	}
	c.queueCh <- message
}

func (c *Client) queueHandler(ctx context.Context) {
	for {
		select {
		case message := <-c.queueCh:
			for {
				_, err := c.client.Send(c.Chat, message, telebot.ModeMarkdown)

				// In case of 429, retry after waiting
				var floodErr telebot.FloodError
				if errors.As(err, &floodErr) {
					c.Logger.Warn(fmt.Sprintf("Received 429 from telegram, retrying in %ds", floodErr.RetryAfter), zap.Error(err))
					time.Sleep(time.Second * time.Duration(floodErr.RetryAfter))
					continue
				}

				// In case of other error, log the error and ignore this message
				if err != nil {
					c.Logger.Error(fmt.Sprintf("Failed to send message '%s' to telegram", message), zap.Error(err))
				}

				break
			}
		case <-ctx.Done():
			close(c.queueCh)
			c.queueCh = nil
			return
		}
	}
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
