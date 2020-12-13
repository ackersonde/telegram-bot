package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ackersonde/telegram-bot/commands"
	"github.com/ackersonde/telegram-bot/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func pollForMessages(bot *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		msg := tgbotapi.NewMessage(chatID, "")
		msg.ParseMode = "markdownv2"

		var myCommands = []tgbotapi.BotCommand{
			{Command: "help", Description: "show this list"},
			{Command: "version", Description: "which version am I?"},
			{Command: "sw", Description: "7d forecast schwabhausen"},
			{Command: "rmls", Description: "list contents of remarkable"},
		}
		bot.SetMyCommands(myCommands)

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "help":
				cmds, _ := bot.GetMyCommands()
				for _, cmd := range cmds {
					msg.Text = msg.Text + "`" + cmd.Command + "` " + cmd.Description + "\n"
				}
			case "version":
				fingerprint := utils.GetDeployFingerprint("/root/.ssh/id_ed25519-cert.pub")
				githubRunID := os.Getenv("GITHUB_RUN_ID")
				msg.Text = "[" + githubRunID + "](https://github.com/ackersonde/telegram-bot/actions/runs/" +
					githubRunID + ") using " + fingerprint
			case "sw":
				msg.Text = "[7-day forecast Schwabhausen](https://darksky.net/forecast/48.3028,11.3591/ca24/en#week)"
			case "rmls":
				var err error
				msg.Text, err = commands.ShowTreeAtPath(update.Message.CommandArguments())
				if err != nil {
					msg.Text = err.Error()
				}
			default:
				msg.Text = "I don't know the command '" + update.Message.Text + "'"
			}
		} else if update.Message.Document != nil { // || update.Message.Photo != nil {
			msg.Text = commands.StoreTelegramFile(bot, update.Message.Document)
			resp, err := bot.Send(msg)

			if err == nil && update.Message.Document.MimeType == "application/pdf" { // || "application/epub" ?
				msg.Text = commands.UploadTelegramPDF2RemarkableCloud(bot, update.Message.Document)
				edit := tgbotapi.EditMessageTextConfig{
					BaseEdit: tgbotapi.BaseEdit{
						ChatID:    chatID,
						MessageID: resp.MessageID,
					},
					Text: msg.Text,
				}
				bot.Send(edit)
				msg.Text = ""
			}
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
