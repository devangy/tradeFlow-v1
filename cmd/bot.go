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

func Bot(tgEventC chan JData, walletStatsC chan WalletStats) {

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

		go func() {
			bgctx := context.Background()

			go func() {
				for wallet := range walletStatsC {

					wallet := fmt.Sprintf(
						"🧠 <b>Smart Wallet Detected</b>\n\n"+
							"👤 <b>Trader:</b> %s\n\n"+
							"🏦 <b>Wallet:</b> "+
							"<a href=\"https://polymarket.com/profile/%s\">%s</a>\n\n"+
							"📊 <b>Performance</b>\n\n"+
							"• Trades: <b>%d</b>\n"+
							"• Wins: <b>%d</b>\n"+
							"• Losses: <b>%d</b>\n"+
							"• Win Rate: <b>%.2f%%</b>\n\n"+
							"💰 <b>Profit</b>\n\n"+
							"• Total PnL: <b>$%.2f</b>\n"+
							"• Profit Factor: <b>%.2f</b>\n\n"+
							"🤖 <b>Bot Risk</b>\n"+
							"• Bot Flags: <b>%.0f</b>\n\n"+
							"⭐ <b>Final Score:</b> <b>%.3f</b>\n\n"+
							"<i>AppendTime: %s</i>",
						wallet.Trader,
						wallet.Address, // URL part
						wallet.Address, // visible text part  ✅ THIS WAS MISSINGdd
						wallet.TotalTrades,
						wallet.Wins,
						wallet.Losses,
						wallet.WinRate,
						wallet.TotalProfit,
						wallet.ProfitFactor,
						wallet.BotFlags,
						wallet.Score,
						wallet.AppendTime,
					)

					_, err := ctx.Bot().SendMessage(
						bgctx,
						tu.Message(tu.ID(chatID), wallet).
							WithParseMode("HTML"),
					)
					if err != nil {
						log.Error("PolyWallet | Failed to send message:", err)
					}

				}
			}()

			for event := range tgEventC {

				var msg string

				switch event.Name {

				case "poly":
					// optional guard if you insist on Name
					log.Debug("msg", msg)
					//

					msg = fmt.Sprintf(
						"🟣 Polymarket\n\n"+
							"Title: %s\n\n"+
							"Volume:$%.2f\n\n",
						event.Title,
						event.Volume,
					)

				case "kalshi":
					log.Debug("msg", msg)

					msg = fmt.Sprintf(
						"🔵 Kalshi\n\n"+
							"Title: %s\n\n"+
							"Category: %s\n\n"+
							"Event Ticker: %s\n\n"+
							"Series: %s\n\n",
						event.Title,
						event.Category,
						event.EventTicker,
						event.SeriesTicker,
					)

				default:
					continue
				}

				if msg == "" {
					log.Debug("Empty message")
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

				time.Sleep(1 * time.Second)
			}
		}()

		return nil

	}, th.AnyCallbackQueryWithMessage(), th.CallbackDataEqual("all_events"))

	bh.Start()

}
