package interfaces

import (
	"github.com/Farengier/smart-home/internal/telegram/session"
)

type Command interface {
	Cmd() string
	Description() string
	Usage() string
	FloodControlLevel() int
	IsAuthRequired() bool
	// PreAction делает проверки ввалидности параметров и сессии. Если возвращает true то выполенние команды прерывается
	// Выполняется до проверки на флуд
	PreAction(r Replier, params []string, sess *session.Session) bool
	Action(r Replier, params []string, sess *session.Session) CommandActionResult
}

type CommandActionResult interface {
	ResetSpamFilter() bool
}
