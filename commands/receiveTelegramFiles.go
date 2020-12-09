package commands

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// StoreTelegramFile is now commented
func StoreTelegramFile(bot *tgbotapi.BotAPI, message *tgbotapi.Message) string {
	response := "Sorry, I couldn't download your file."

	downloadURL, err := bot.GetFileDirectURL(message.Document.FileID)
	if err != nil {
		response = fmt.Sprintf("couldn't get file URL: %s", err.Error())
	} else {
		response = fmt.Sprintf("Attempting to download: %s", downloadURL)
	}
	err = downloadFile(downloadURL, message.Document.FileName)
	if err != nil {
		response = fmt.Sprintf("failed to download file: %s", err.Error())
	} else {
		response = fmt.Sprintf("Downloaded %s", message.Document.FileName)
	}
	return response
}

// SendDirectlyToRemarkable is now commented
func SendDirectlyToRemarkable(bot *tgbotapi.BotAPI, fileName string) string {
	response := "Unable to convert/send file to Remarkable :( "
	cmd := exec.Command("/bin/sh", "pdf2Remarkable.sh", "-r",
		os.TempDir()+"/"+strings.ReplaceAll(fileName, " ", "_"))
	cmd.Env = append(os.Environ(), "REMARKABLE_HOST=192.168.178.80")

	results, err := cmd.CombinedOutput()
	if err != nil {
		response = response + err.Error()
	} else {
		response = string(results)
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
