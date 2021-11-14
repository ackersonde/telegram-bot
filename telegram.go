package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ackersonde/telegram-bot/commands"
	"github.com/ackersonde/telegram-bot/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func pollForMessages(bot *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel) {
	setCommands := tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{Command: "help", Description: "show this list"},
		tgbotapi.BotCommand{Command: "version", Description: "which version am I?"},
		tgbotapi.BotCommand{Command: "sw", Description: "7d forecast schwabhausen"},
		tgbotapi.BotCommand{Command: "rmls", Description: "list contents of remarkable"})

	if _, err := bot.Request(setCommands); err != nil {
		log.Printf("Unable to set commands: %s\n", err.Error())
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		msg := tgbotapi.NewMessage(chatID, "")

		if update.Message.IsCommand() {
			args := update.Message.CommandArguments()
			command := update.Message.Command()

			if strings.HasSuffix(update.Message.Text, "@hop_on_pop_bot") {
				args = command
				command = "rmls"
			}

			switch command {
			case "help":
				msg.ParseMode = "MarkdownV2"

				cmds, _ := bot.GetMyCommands()
				for _, cmd := range cmds {
					msg.Text = msg.Text + "`" + cmd.Command + "` " + tgbotapi.EscapeText(msg.ParseMode, cmd.Description) + "\n"
				}

			case "version":
				msg.ParseMode = "MarkdownV2"

				githubRunID := os.Getenv("GITHUB_RUN_ID")
				fingerprint := utils.GetDeployFingerprint("/root/.ssh/id_ed25519-cert.pub")
				// cut from Principals:
				fingerprint = tgbotapi.EscapeText(msg.ParseMode,
					fingerprint[0:strings.LastIndex(fingerprint, "Principals:")])

				msg.Text = "[" + githubRunID + "](https://github.com/ackersonde/telegram-bot/actions/runs/" +
					githubRunID + ") using " + fingerprint

			case "sw":
				msg.ParseMode = "MarkdownV2"
				msg.Text = "[7d forecast Schwabhausen](https://darksky.net/forecast/48.3028,11.3591/ca24/en#week)"

			case "rmls":
				var err error
				msg.Text, err = commands.ShowTreeAtPath(args)
				if err != nil {
					msg.Text = err.Error()
				} else {
					args = "/" + args
					msg.Text = "reMarkable files at '" + args + "':\n\n" + msg.Text
				}

			default:
				msg.Text = "I don't know the command '" + update.Message.Text + "'"
			}
		} else if update.Message.Document != nil { // || update.Message.Photo != nil {
			msg.Text = commands.StoreTelegramFile(bot, update.Message.Document)
			resp, err := bot.Send(msg)

			if err == nil {
				switch update.Message.Document.MimeType {
				case "application/pdf", "application/epub", "application/epub+zip":
					msg.Text = commands.UploadTelegramPDFEPUB2RemarkableCloud(bot, update.Message.Document)
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
		} else if strings.ToLower(update.Message.Text) == "help" {
			msg.Text = "This bot responds only to commands - try /help"
		}

		if msg.Text != "" {
			_, err := bot.Send(msg)
			if err != nil {
				log.Printf("Unable to send msg to Telegram: %s\n", err.Error())
			}
		}
	}
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("CTX_TELEGRAM_BOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)
	webHookURL := "https://" + os.Getenv("TELEGRAM_BOT_WEB_URL") + "/" + bot.Token

	wh, err := tgbotapi.NewWebhook(webHookURL)
	if err != nil {
		log.Fatal(err)
	}

	_, err = bot.Request(wh)
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

	log.Printf("Now listening on %s", webHookURL)
	pollForMessages(bot, updates)
}
