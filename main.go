package main

import (
	"context"
	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"nyne_bot/config"
	"nyne_bot/services"
	"os"
	"os/signal"
)

func main() {
	config.ReadConfig()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []tgbot.Option{
		tgbot.WithDefaultHandler(handler),
	}

	b, err := tgbot.New(config.GetConfig().Token, opts...)
	if err != nil {
		panic(err)
	}

	b.Start(ctx)
}

func handler(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	services.HandleUpdate(ctx, b, update)
}
