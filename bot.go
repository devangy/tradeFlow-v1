package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func Bot(jsonCh <-chan any) {

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		panic("Bot token is empty! Set BOT_TOKEN environment variable.")
	}

	// context with a timeout for cancellation of request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	bot, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	botUser, err := bot.GetMe(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Bot user: %+v\n", botUser)

	// updates from bot
	updates, _ := bot.UpdatesViaLongPolling(ctx, nil)

	// // processing updates one by one
	// for update := range updates {
	// 	fmt.Printf("update %v", update)
	// }

	// bot handler to handle req
	bh, _ := th.NewBotHandler(bot, updates)
	defer func() { _ = bh.Stop() }()

	// Register new handler with match on command `/start`
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		_, _ = ctx.Bot().SendMessage(ctx, tu.Message(
			tu.ID(update.Message.Chat.ID),
			fmt.Sprintf("Hello %s!", update.Message.From.FirstName),
		))
		return nil
	}, th.CommandEqual("start"))

	bh.Start()

	defer bh.Stop()

	// for msg := range jsonCh {
	// 	fmt.Println("Received JSON:", msg)
	// }

}
