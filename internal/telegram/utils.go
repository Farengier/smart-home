package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
)

func strp(str string) *string {
	return &str
}

func (b *bot) send(c tgbotapi.Chattable) {
	if _, err := b.botAPI.Send(c); err != nil {
		log.Errorf("[TG Bot] failed sending: %s", err)
	}
}
