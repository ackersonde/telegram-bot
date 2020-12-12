package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ackersonde/telegram-bot/commands"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func downloadFile(URL, fileName string) error {
	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}

	fileName = strings.ReplaceAll(fileName, " ", "_")

	//Create a empty file
	file, err := os.Create(os.TempDir() + "/" + fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	//Write the bytes to the file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func pollForMessages(bot *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		imageDir := "https://ackerson.de/images/telegram-bot-images/"
		chatID := update.Message.Chat.ID
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "anchor":
				msg.ParseMode = "html"
				msg.DisableWebPagePreview = false
				msg.Text = "<a href=\"" + imageDir + "rm.png\">&#8205;</a> /rmls (dir): List contents of reMarkable\n"
			case "image":
				msg.ParseMode = "markdownv2"
				msg.DisableWebPagePreview = false
				msg.Text = "[](" + imageDir + "rm.png) /rmls (dir): List img contents of reMarkable\n"
			case "help":
				msg.Text = "Known cmds incl: /help /rmls /sw /crypto /pi /pgp /torq..."

				// sw

				// crypto

				// pi

				// pgp

				// torq

				// trans

				// vpn

				// wg
			case "imgTag":
				msg.ParseMode = "html"
				msg.Text = "This will be interpreted as HTML: <img src=\"" + imageDir + "rm.png\">"

			case "stickerImage":
				downloadFile(imageDir+"rm.png", "rm.png")
				msg := tgbotapi.NewStickerUpload(chatID, os.TempDir()+"/rm.png")
				sticker, err := bot.Send(msg)
				if err != nil {
					log.Printf("%s\n", err.Error())
				} else {
					log.Printf("Reuse sticker ID: %s\n", sticker.Document.FileID)
				}
			case "mediaPhoto":
				image := tgbotapi.NewInputMediaPhoto(imageDir + "rm.png")
				image.Caption = "Testing 123"
				cfg := tgbotapi.NewMediaGroup(
					update.Message.Chat.ID,
					[]interface{}{
						image,
					})

				_, err := bot.Send(cfg)
				if err != nil {
					log.Printf("%s\n", err.Error())
				}

			case "photoUpload":
				downloadFile(imageDir+"rm.png", "rm.png")

				msg := tgbotapi.NewPhotoUpload(chatID, os.TempDir()+"/rm.png")
				msg.ReplyToMessageID = update.Message.MessageID
				msg.Caption = "Test"
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

	bot.Debug = false

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
