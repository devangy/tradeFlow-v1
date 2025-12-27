package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func Bot(tgEventC chan any) {

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("Bot token is empty! Set BOT_TOKEN environment variable.")
	}

	// context with a timeout for cancellation of request
	ctx := context.Background()

	bot, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	botUser, err := bot.GetMe(ctx)
	if err != nil {
		log.Fatal("bot authentication: ", err)
	}

	log.Debug("Bot user: %+v\n", botUser)

	// updates from bot via long polling for testing
	updates, _ := bot.UpdatesViaLongPolling(ctx, nil)

	// // processing updates one by one
	// for update := range updates {
	// 	fmt.Printf("update %v", update)
	// }

	// bot handler to handle req
	bh, _ := th.NewBotHandler(bot, updates)

	defer bh.Stop()

	// Register new handler with match on command `/start`
	bh.HandleMessage(func(ctx *th.Context, message telego.Message) error {
		// Send a message with inline keyboard
		_, _ = ctx.Bot().SendMessage(ctx, tu.Messagef(
			tu.ID(message.Chat.ID),
			`Hello %s !  Welcome to DataLog`, message.From.FirstName,
		).WithReplyMarkup(tu.InlineKeyboard(
			tu.InlineKeyboardRow(tu.InlineKeyboardButton("Start").WithCallbackData("all_events"))),
		))
		return nil
	}, th.CommandEqual("start"))

	// Register new handler with match on a call back query with data equal to `go` and non-nil message
	bh.HandleCallbackQuery(func(ctx *th.Context, query telego.CallbackQuery) error {

		for event := range tgEventC {
			fmt.Print("tgChanEvents", event)
			_, _ = ctx.Bot().SendMessage(ctx, tu.Messagef(
				tu.ID(query.Message.GetChat().ID),
				"Received: %v", event,
			))
		}
		// }()

		// _, _ = bot.SendMessage(ctx, tu.Messagef(tu.ID(query.Message.GetChat().ChatID().ID), "e: %s", single))

		// Answer callback query
		_ = bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID).WithText("Done"))

		return nil

	}, th.AnyCallbackQueryWithMessage(), th.CallbackDataEqual("all_events"))

	bh.Start()

}
