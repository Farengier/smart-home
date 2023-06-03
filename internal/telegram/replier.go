package telegram

import (
	"github.com/Farengier/smart-home/internal/telegram/interfaces"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"io"
	"time"
)

type replier struct {
	b      *bot
	chatID int64
	cmd    interfaces.Command
}

func (r *replier) InternalError() {

}

func (r *replier) Usage() {
	r.ReplyWithMessage(r.cmd.Usage())
}

func (r *replier) ReplyWithMessage(msg string) {
	reply := tgbotapi.NewMessage(r.chatID, msg)
	reply.ParseMode = "MarkdownV2"
	r.b.send(reply)
}

func (r *replier) SensitivePicture(pic io.Reader) {
	msg := tgbotapi.NewPhoto(r.chatID, tgbotapi.FileReader{
		Name:   "QR.png",
		Reader: pic,
	})

	sentMsg, err := r.b.botAPI.Send(msg)
	if err != nil {
		log.Errorf("[TG Bot] failed sending: %s", err)
	}

	time.Sleep(time.Minute)
	_, err = r.b.botAPI.Send(tgbotapi.NewDeleteMessage(r.chatID, sentMsg.MessageID))
	if err != nil {
		log.Errorf("[TG Bot] failed deleting message: %s", err)
	}
}
