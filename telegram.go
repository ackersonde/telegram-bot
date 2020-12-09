package main2

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ackersonde/telegram-bot/commands"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func pollForMessages(bot *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "help":
				msg.Text = "type /sayhi or /status."
			case "sayhi":
				msg.Text = "Hi :)"
			case "status":
				msg.Text = "I'm ok."
			case "withArgument":
				msg.Text = "You supplied the following argument: " + update.Message.CommandArguments()
			case "html":
				msg.ParseMode = "html"
				msg.Text = "This will be interpreted as HTML, click <a href=\"https://www.example.com\">here</a>"
				/* or for custom images:
				<a href="' + image + '">&#8205;</a> // &#8205; -> never show in message
				also you must set disable_web_page_preview=false */
			default:
				msg.Text = "I don't know that command"
			}
		} else if update.Message.Document != nil { // || update.Message.Photo != nil {
			commands.StoreTelegramFile(bot, update.Message)
			log.Printf("mimetype for %s: %s\n", update.Message.Document.FileName, update.Message.Document.MimeType)
			//msg.Text = commands.SendDirectlyToRemarkable(bot, update.Message.Document.FileName)
			// TODO: only send PDF files
			// TODO: test rMAPI for true cloud native approach: https://github.com/juruen/rmapi
			msg.Text = commands.UploadTelegramPDF2RemarkableCloud(bot, update.Message.Document)
		}

		if msg.Text != "" {
			bot.Send(msg)
		}
	}
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("CTX_TELEGRAM_BOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(
		"https://" + os.Getenv("TELEGRAM_BOT_WEB_URL") + "/" + bot.Token))
	if err != nil {
		log.Fatal(err)
	}
	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}
	if info.LastErrorDate != 0 {
		log.Printf("Telegram callback last failure @ %s: %s",
			time.Unix(int64(info.LastErrorDate), 0),
			info.LastErrorMessage)
	}

	updates := bot.ListenForWebhook("/" + bot.Token)

	// wait for potential large backlog of old msgs and clear
	time.Sleep(time.Millisecond * 500)
	updates.Clear()

	go http.ListenAndServe("0.0.0.0:3000", nil)

	pollForMessages(bot, updates)
}
