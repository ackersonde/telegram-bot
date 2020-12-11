package commands

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// StoreTelegramFile is now commented
func StoreTelegramFile(bot *tgbotapi.BotAPI, document *tgbotapi.Document) string {
	response := "Sorry, I couldn't download your file."

	downloadURL, err := bot.GetFileDirectURL(document.FileID)
	if err != nil {
		response = fmt.Sprintf("couldn't get file URL: %s", err.Error())
	} else {
		response = fmt.Sprintf("Attempting to download: %s", downloadURL)
	}
	err = downloadFile(downloadURL, document.FileName)
	if err != nil {
		response = fmt.Sprintf("failed to download file: %s", err.Error())
	} else {
		response = fmt.Sprintf("Downloaded %s", document.FileName)
	}
	return response
}

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
