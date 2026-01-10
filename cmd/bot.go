package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func Bot(tgEventC <-chan any) {

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("Bot token is empty! Set BOT_TOKEN environment variable.")
	}
	//dddddddddddd
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
	//deviig
	log.Debug("Bot user: %+v\n", botUser)

	// updates from bot via long polling for testing
	updates, err := bot.UpdatesViaLongPolling(ctx, nil)
	if err != nil {
		log.Fatal("failed to start long polling:", err)
	}

	// bot handler to handle req
	bh, _ := th.NewBotHandler(bot, updates)

	defer func() { _ = bh.Stop() }()

	// userID_map := make(map[int64]struct{})

	// Register new handler with match on command `/start`
	bh.HandleMessage(func(ctx *th.Context, message telego.Message) error {

		// userID_map[message.Chat.ID] = struct{}{}

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
	//
	//
	//
	//
	// {}

	bh.HandleCallbackQuery(func(ctx *th.Context, query telego.CallbackQuery) error {

		// Answer callback query
		_ = bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID).WithText("Done"))

		chatID := query.Message.GetChat().ID

		// for event := range tgEventC {

		// 	fmt.Println("tgChanEvents", event)

		// 	// 3. Send the event to the user
		// 	_, err := ctx.Bot().SendMessage(ctx, tu.Messagef(
		// 		tu.ID(chatID),
		// 		"Received: %v", event,
		// 	))

		// 	if err != nil {
		// 		fmt.Printf("Failed to send message: %v\n", err)
		// 	}
		// }
		//

		type polymarketdata struct {
			Name     string
			Title    string  `json:"title"`
			Category string  `json:"category"`
			Volume   float64 `json:"volume"`
			Image    string  `json:"image"`
		}

		type kalsh struct {
			Name         string
			Title        string `json:"title"`
			EventTicker  string `json:"event_ticker"`
			SeriesTicker string `json:"series_ticker"`
			Category     string `json:"category"`
		}

		go func() {
			bgctx := context.Background()

			for event := range tgEventC {

				var msg string

				switch ev := event.(type) {

				case polymarketdata:
					// optional guard if you insist on Name

					msg = fmt.Sprintf(
						"🟣 <b>Polymarket</b>\n\n"+
							"📰 <b>Title:</b> %s\n"+
							"🏷️ <b>Category:</b> %s\n"+
							"💰 <b>Volume:</b> $%.2f\n",
						ev.Title,
						ev.Category,
						ev.Volume,
					)

				case kalshi:

					msg = fmt.Sprintf(
						"🔵 <b>Kalshi</b>\n\n"+
							"📰 <b>Title:</b> %s\n"+
							"🏷️ <b>Category:</b> %s\n"+
							"🎯 <b>Event Ticker:</b> <code>%s</code>\n"+
							"📦 <b>Series:</b> <code>%s</code>\n",
						ev.Title,
						ev.Category,
						ev.EventTicker,
						ev.SeriesTicker,
					)

				default:
					// unknown type — ignore safely
					continue
				}

				if msg == "" {
					continue
				}

				_, err := ctx.Bot().SendMessage(
					bgctx,
					tu.Message(tu.ID(chatID), msg).
						WithParseMode("Markdown"),
				)

				if err != nil {
					log.Error("Failed to send message:", err)
				}

				time.Sleep(100 * time.Millisecond)
			}
		}()

		return nil

	}, th.AnyCallbackQueryWithMessage(), th.CallbackDataEqual("all_events"))

	bh.Start()

}
