package commands

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Farengier/smart-home/internal/img"
	"github.com/Farengier/smart-home/internal/telegram/domain"
	"github.com/Farengier/smart-home/internal/telegram/interfaces"
	"github.com/Farengier/smart-home/internal/telegram/session"
	"github.com/jltorresm/otpgo"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"image/png"

	"github.com/Farengier/smart-home/internal/orm"
)

const totpIssuer = "zy-smart-home"

type registerCmd struct {
	db *gorm.DB
}

func Register(db *gorm.DB) *registerCmd {
	return &registerCmd{db: db}
}
func (rc *registerCmd) Cmd() string {
	return "register"
}
func (rc *registerCmd) Description() string {
	return "Регистраия нового пользователя"
}
func (rc *registerCmd) Usage() string {
	return `Для регистрации нового пользователя введите команду
/register \<логин пользователя\>

Держите телефон наготове для сканирования QR\-кода
Через минуту сообщение с кодом будет удалено`
}
func (rc *registerCmd) FloodControlLevel() int {
	return domain.SpamLevelSensitive
}
func (rc *registerCmd) IsAuthRequired() bool {
	return false
}
func (rc *registerCmd) PreAction(r interfaces.Replier, params []string, sess *session.Session) bool {
	if sess.IsAuthenticated() {
		r.ReplyWithMessage(fmt.Sprintf("Already authenticated as %s [%s]", sess.User.Role, sess.User.Login))
		return true
	}

	if len(params) < 1 {
		r.Usage()
		return true
	}

	return false
}
func (rc *registerCmd) Action(r interfaces.Replier, params []string, sess *session.Session) interfaces.CommandActionResult {
	actionRes := &actionResult{resetSpamFilter: false}
	login := params[0]
	usr := &orm.User{}
	dbres := rc.db.Joins("Role").First(usr, orm.User{Login: login})
	if !errors.Is(dbres.Error, gorm.ErrRecordNotFound) {
		r.ReplyWithMessage(fmt.Sprintf("User already registered"))
		return actionRes
	}

	totp := otpgo.TOTP{}
	_, err := totp.Generate()
	if err != nil {
		log.Errorf("[TG Bot Totp] key generate failed: %s", err)
		r.InternalError()
		return actionRes
	}

	usr.OtpKey = totp.Key
	usr.Login = login
	usr.Role = orm.UserRole{Role: "user"}
	rc.db.Create(usr)

	ku := totp.KeyUri(login, totpIssuer)
	qrcode, err := ku.QRCode()
	if err != nil {
		log.Errorf("[TG Bot Totp] key generate failed: %s", err)
		r.InternalError()
		return actionRes
	}

	im, err := img.ParseB64(qrcode)
	if err != nil {
		log.Errorf("[TG Bot register] img parse failed: %s", err)
		r.InternalError()
		return actionRes
	}

	bb := bytes.NewBuffer([]byte{})
	err = png.Encode(bb, im)
	if err != nil {
		log.Errorf("[TG Bot register] encoding img failed: %s", err)
		r.InternalError()
		return actionRes
	}

	actionRes.resetSpamFilter = true
	r.SensitivePicture(bb)
	return actionRes
}
