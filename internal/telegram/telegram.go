package telegram

import (
	"context"
	"fmt"
	"github.com/Farengier/smart-home/internal/telegram/commands"
	"github.com/Farengier/smart-home/internal/telegram/domain"
	"github.com/Farengier/smart-home/internal/telegram/interfaces"
	"github.com/Farengier/smart-home/internal/telegram/session"
	"gorm.io/gorm"
	"strings"
	"time"

	"github.com/Farengier/smart-home/internal/signal"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
)

const msgInternalErr = "Internal error, please contact admin"

type Config interface {
	Token() string
	SpamFilterDurationSensitive() time.Duration
	SpamFilterDurationLow() time.Duration
}

type DB interface {
	GORM() *gorm.DB
	SyncNow()
}

type bot struct {
	cfg           Config
	botAPI        *tgbotapi.BotAPI
	sessions      *session.Storage
	db            DB
	spamDurations map[int]time.Duration
	handlers      map[string]func(upd tgbotapi.Update)
	commands      map[string]interfaces.Command
}

func (b *bot) initCommands() {
	cmds := []interfaces.Command{
		commands.Start(),
		commands.Login(b.db.GORM()),
		commands.Register(b.db.GORM()),
	}

	b.commands = map[string]interfaces.Command{}
	for _, cmd := range cmds {
		if _, ok := b.commands[cmd.Cmd()]; ok {
			panic("Commands intersection: " + cmd.Cmd())
		}
		b.commands[cmd.Cmd()] = cmd
	}
}

func StartBot(cfg Config, db DB) error {
	ctx := context.Background()
	tgbot, err := tgbotapi.NewBotAPI(cfg.Token())
	if err != nil {
		return fmt.Errorf("telegram bot start failed: %w", err)
	}

	tgbot.Debug = true

	instance := &bot{
		cfg:      cfg,
		botAPI:   tgbot,
		db:       db,
		sessions: session.New(),
		spamDurations: map[int]time.Duration{
			domain.SpamLevelLow:       cfg.SpamFilterDurationLow(),
			domain.SpamLevelSensitive: cfg.SpamFilterDurationSensitive(),
		},
	}
	instance.setCommands()

	// Create a new UpdateConfig struct with an offset of 0. Offsets are used
	// to make sure Telegram knows we've handled previous values and we don't
	// need them repeated.
	updateConfig := tgbotapi.NewUpdate(0)

	// Tell Telegram we should wait up to 30 seconds on each request for an
	// update. This way we can get information just as quickly as making many
	// frequent requests without having to send nearly as many.
	updateConfig.Timeout = 30
	// Start polling Telegram for updates.
	updates := tgbot.GetUpdatesChan(updateConfig)

	tbctx, cncl := context.WithCancel(ctx)
	signal.OnShutdown(func() error {
		log.Info("[TBot] Shutdown telegram bot")
		tgbot.StopReceivingUpdates()
		cncl()
		return nil
	})
	signal.Run(func() { instance.read(tbctx, updates) })
	return nil
}

func (b *bot) setCommands() {
	b.initCommands()
	botCommands := make([]tgbotapi.BotCommand, 0, len(b.commands))
	for cmd, cmdDesc := range b.commands {
		botCommands = append(botCommands, tgbotapi.BotCommand{
			Command:     "/" + cmd,
			Description: cmdDesc.Description(),
		})
	}
	response, err := b.botAPI.Request(tgbotapi.SetMyCommandsConfig{
		Commands:     botCommands,
		Scope:        nil,
		LanguageCode: "",
	})
	if err != nil {
		log.Errorf("[TBot] [init] setting commands failed: %s", err)
	} else {
		log.Infof("[TBot] [init] setting commands: ok, %+v", response)
	}
}

func (b *bot) read(ctx context.Context, updates tgbotapi.UpdatesChannel) {
	// Let's go through each update that we're getting from Telegram.
	for {
		select {
		case upd := <-updates:
			b.update(upd)
		case <-ctx.Done():
			return
		}
	}
}

func (b *bot) update(upd tgbotapi.Update) {
	if upd.CallbackQuery != nil {
		_ = b.sessions.Session(upd.CallbackQuery.Message.Chat.ID)
		return
	}
	if upd.Message != nil {
		sess := b.sessions.Session(upd.Message.Chat.ID)
		b.msgUpdate(upd, sess)
		return
	}
}

func (b *bot) msgUpdate(upd tgbotapi.Update, sess *session.Session) {
	parts := strings.Split(upd.Message.Text, " ")
	r := &replier{chatID: sess.ChatID, b: b}
	if len(parts) == 0 {
		r.ReplyWithMessage("empty message")
		return
	}

	cmdName := parts[0]
	if !strings.HasPrefix(cmdName, "/") {
		r.ReplyWithMessage("not a command")
		return
	}

	cmd, ok := b.commands[parts[0][1:]]
	r = &replier{chatID: sess.ChatID, b: b, cmd: cmd}
	if !ok {
		r.ReplyWithMessage("unknown command")
		return
	}

	if cmd.IsAuthRequired() && !sess.IsAuthenticated() {
		// TODO
		return
	}

	if cmd.PreAction(r, parts[1:], sess) {
		return
	}

	b.spamCheck(r, sess, cmd.FloodControlLevel())

	ares := cmd.Action(r, parts[1:], sess)

	if ares.ResetSpamFilter() {
		sess.Spam.Set(cmd.FloodControlLevel(), time.Time{})
	}
}

func (b *bot) spamCheck(r *replier, sess *session.Session, l int) bool {
	if l == domain.SpamLevelNone {
		return false
	}

	t := sess.Spam.Get(l)
	delta := t.Sub(time.Now())
	if delta > 0 {
		r.ReplyWithMessage(fmt.Sprintf("Try again after %s", delta.Truncate(time.Second)+time.Second))
		return false
	}

	sd := b.spamDurations[l]
	sess.Spam.Set(l, time.Now().Add(sd))
	return true
}
