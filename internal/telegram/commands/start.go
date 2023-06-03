package commands

import (
	"github.com/Farengier/smart-home/internal/telegram/domain"
	"github.com/Farengier/smart-home/internal/telegram/interfaces"
	"github.com/Farengier/smart-home/internal/telegram/session"
)

type startCmd struct {
}

func Start() *startCmd {
	return &startCmd{}
}
func (sc *startCmd) Cmd() string {
	return "start"
}
func (sc *startCmd) Description() string {
	return "Стартует новую сессию взаимодейстия"
}
func (sc *startCmd) Usage() string {
	return `Для старта новой сессии взаимодействия просто используйте
[/start](/start)`
}
func (sc *startCmd) FloodControlLevel() int {
	return domain.SpamLevelNone
}
func (sc *startCmd) IsAuthRequired() bool {
	return false
}
func (sc *startCmd) PreAction(r interfaces.Replier, params []string, sess *session.Session) bool {
	return false
}
func (sc *startCmd) Action(r interfaces.Replier, params []string, sess *session.Session) interfaces.CommandActionResult {
	msg := `**Добро пожаловать в мой умный дом\!**  

Чтобы продолжить используйте следующие команды:  
 \* /login \<user\_name\> \<code\>
 \* /register \<user\_name\>`
	sess.Clear()
	r.ReplyWithMessage(msg)
	return (*actionResult)(nil)
}

/*
msg := tgbotapi.NewMessage(sess.ChatID, "Welcome to zy smart home!")
	msg.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			[]tgbotapi.InlineKeyboardButton{
				tgbotapi.InlineKeyboardButton{
					Text:         "Login",
					CallbackData: strp("login"),
				},
				tgbotapi.InlineKeyboardButton{
					Text:         "Register",
					CallbackData: strp("Register"),
				},
			},
		},
	}
*/
