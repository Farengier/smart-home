package commands

import (
	"errors"
	"fmt"
	"github.com/Farengier/smart-home/internal/orm"
	"github.com/Farengier/smart-home/internal/telegram/domain"
	"github.com/Farengier/smart-home/internal/telegram/interfaces"
	"github.com/Farengier/smart-home/internal/telegram/session"
	"github.com/jltorresm/otpgo"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const wrongCreds = "Wrong credentials"

type loginCmd struct {
	db *gorm.DB
}

func Login(db *gorm.DB) *loginCmd {
	return &loginCmd{db: db}
}
func (lc *loginCmd) Cmd() string {
	return "login"
}
func (lc *loginCmd) Description() string {
	return "Авторизация пользоввателя"
}
func (lc *loginCmd) Usage() string {
	return `Для авторизации пользователя используйте команду
/login \<логин пользователя\> \<одноразовый код\>`
}
func (lc *loginCmd) FloodControlLevel() int {
	return domain.SpamLevelSensitive
}
func (lc *loginCmd) IsAuthRequired() bool {
	return false
}
func (lc *loginCmd) PreAction(r interfaces.Replier, params []string, sess *session.Session) bool {
	if sess.IsAuthenticated() {
		r.ReplyWithMessage(fmt.Sprintf("Already authenticated as %s [%s]", sess.User.Role, sess.User.Login))
		return true
	}

	if len(params) < 2 {
		r.Usage()
		return true
	}

	return false
}
func (lc *loginCmd) Action(r interfaces.Replier, params []string, sess *session.Session) interfaces.CommandActionResult {
	actionRes := &actionResult{resetSpamFilter: false}

	login := params[0]
	code := params[1]
	usr := &orm.User{}
	dbres := lc.db.Model(orm.User{}).Joins("Role").First(usr, orm.User{Login: login})

	if errors.Is(dbres.Error, gorm.ErrRecordNotFound) {
		r.ReplyWithMessage(wrongCreds)
		return actionRes
	}

	totp := otpgo.TOTP{
		Key: usr.OtpKey,
	}
	ok, err := totp.Validate(code)
	if err != nil {
		log.Errorf("[TG Bot Auth Totp] validating error: %s", err)
		r.ReplyWithMessage("Internal error, please contact admin")
		return actionRes
	}
	if !ok {
		r.ReplyWithMessage(wrongCreds)
		return actionRes
	}

	u := session.MakeUser(0, usr.Login, usr.Role.Role)
	sess.User = u
	actionRes.resetSpamFilter = true

	r.ReplyWithMessage(fmt.Sprintf("Successfully authenticated as %s [%s]", u.Role, u.Login))
	return actionRes
}
