package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// sendTextToTelegramChat sends a text message to the Telegram chat identified by its chat Id
func sendTextToTelegramChat(chatID int64, text string) (string, error) {
	log.Printf("Sending %s to chat_id: %d", text, chatID)
	var telegramAPI string = "https://api.telegram.org/bot" +
		os.Getenv("CTX_TELEGRAM_BOT_TOKEN") + "/sendMessage"
	response, err := http.PostForm(
		telegramAPI,
		url.Values{
			"chat_id": {strconv.FormatInt(chatID, 10)},
			"text":    {text},
		})

	if err != nil {
		log.Printf("error when posting text to the chat: %s", err.Error())
		return "", err
	}
	defer response.Body.Close()

	var bodyBytes, errRead = ioutil.ReadAll(response.Body)
	if errRead != nil {
		log.Printf("error in parsing telegram answer %s", errRead.Error())
		return "", err
	}
	bodyString := string(bodyBytes)
	log.Printf("Body of Telegram Response: %s", bodyString)

	return bodyString, nil
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
		log.Printf("Telegram callback failed: %s", info.LastErrorMessage)
	}
	updates := bot.ListenForWebhook("/" + bot.Token)
	go http.ListenAndServe("0.0.0.0:3000", nil)

	for update := range updates {
		log.Printf("%s wrote '%s'\n", update.Message.From, update.Message.Text)
		response := fmt.Sprintf("good copy %s. rcvd %s\n",
			update.Message.From, update.Message.Text)
		sendTextToTelegramChat(update.Message.Chat.ID, response)
	}
}
