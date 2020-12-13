package commands

import (
	"fmt"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/juruen/rmapi/api"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
)

func getRemarkableAPICtx() *api.ApiCtx {
	log.InitLog()
	var ctx *api.ApiCtx
	var err error
	for i := 0; i < 3; i++ {
		auth := api.AuthHttpCtx(i > 0, true)
		log.Trace.Printf("AUTH: %v\n", auth)
		ctx, err = api.CreateApiCtx(auth)

		if err != nil {
			log.Error.Printf("%s\n", err)
		} else {
			break
		}
	}

	if ctx == nil {
		log.Error.Printf("failed to build documents tree, last error: %s\n", err)
	}

	return ctx
}

// UploadTelegramPDF2RemarkableCloud is now commented
func UploadTelegramPDF2RemarkableCloud(bot *tgbotapi.BotAPI,
	telegramDocument *tgbotapi.Document) string {
	response := "Unable to upload doc to Remarkable Cloud"
	uploadDir := "telegram_files"
	var uploadDocDir model.Document

	ctx := getRemarkableAPICtx()
	if ctx != nil {
		uploadDocNode, err := ctx.Filetree.NodeByPath(uploadDir, ctx.Filetree.Root())

		if err != nil && err.Error() == "entry doesn't exist" {
			uploadDocDir, err = ctx.CreateDir(ctx.Filetree.Root().Id(), uploadDir)
		} else {
			uploadDocDir = *uploadDocNode.Document
		}

		if err != nil {
			response = response + " : " + err.Error()
		} else {
			fileName := strings.ReplaceAll(telegramDocument.FileName, " ", "_")
			rmDocument, err := ctx.UploadDocument(uploadDocDir.ID, os.TempDir()+"/"+fileName)
			if err != nil {
				response = fmt.Sprintf("Upload ERR: %s", err.Error())
			} else {
				response = fmt.Sprintf("Successfully uploaded %s to %s", rmDocument.VissibleName, uploadDir)
			}
		}
	}

	return response
}

// ShowTreeAtPath is now commented
func ShowTreeAtPath(path string) (string, error) {
	response := ""
	ctx := getRemarkableAPICtx()
	if ctx != nil {
		node, err := ctx.Filetree.NodeByPath(path, ctx.Filetree.Root())
		if node == nil || err != nil {
			return "Unable to find '" + path + "'", err
		}
		for _, e := range node.Children {
			if e.IsFile() {
				response = response + fmt.Sprintf("- \t%s\n", e.Name())
			} else {
				response = response + fmt.Sprintf("[%s](/rmls %s)\n", e.Name(), path+e.Name())
			}
		}
	}

	return response, nil
}
