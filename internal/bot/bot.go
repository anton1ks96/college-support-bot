package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/anton1ks96/college-support-bot/internal/config"
	"github.com/anton1ks96/college-support-bot/internal/domain"
	"github.com/anton1ks96/college-support-bot/pkg/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api        *tgbotapi.BotAPI
	groupID    int64
	userStates map[int64]*domain.UserState
}

func New(cfg *config.Config) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, err
	}

	logger.Info(fmt.Sprintf("Authorized on account %s", api.Self.UserName))

	bot := &Bot{
		api:        api,
		groupID:    cfg.Group,
		userStates: make(map[int64]*domain.UserState),
	}

	go bot.cleanupExpiredStates()

	return bot, nil
}

func (b *Bot) cleanupExpiredStates() {
	ticker := time.NewTicker(60 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		for userID, state := range b.userStates {
			if now.Sub(state.CreatedAt) > 60*time.Minute {
				delete(b.userStates, userID)
			}
		}
	}
}

func (b *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			b.handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			b.handleCallbackQuery(update.CallbackQuery)
		}
	}

	return nil
}

func (b *Bot) handleMessage(message *tgbotapi.Message) {
	userID := message.From.ID

	if message.IsCommand() {
		switch message.Command() {
		case "start":
			b.handleStart(message)
		case "done":
			state, exists := b.userStates[userID]
			if exists && state.State != domain.StateNone {
				b.sendToGroup(message, state)
				delete(b.userStates, userID)

				msg := tgbotapi.NewMessage(message.Chat.ID, "Спасибо! Ваше сообщение отправлено.")
				b.api.Send(msg)
			}
		}
		return
	}

	state, exists := b.userStates[userID]
	if !exists || state.State == domain.StateNone {
		return
	}

	if message.Photo != nil {
		b.handlePhoto(message)
	} else if message.Text != "" {
		b.handleText(message)
	}
}

func (b *Bot) handleStart(message *tgbotapi.Message) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Сообщить о проблеме", "report_problem"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Сделать предложение", "make_suggestion"),
		),
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, "Выберите действие:")
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *Bot) handleCallbackQuery(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID

	callback := tgbotapi.NewCallback(query.ID, "")
	b.api.Request(callback)

	switch query.Data {
	case "report_problem":
		b.userStates[userID] = &domain.UserState{
			State:     domain.StateAwaitingProblemReport,
			Photos:    make([]interface{}, 0),
			CreatedAt: time.Now(),
		}
		msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Опишите проблему и приложите до 4 фотографий (если нужно). Когда закончите, отправьте /done")
		b.api.Send(msg)

	case "make_suggestion":
		b.userStates[userID] = &domain.UserState{
			State:     domain.StateAwaitingSuggestion,
			Photos:    make([]interface{}, 0),
			CreatedAt: time.Now(),
		}
		msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Опишите ваше предложение и приложите до 4 фотографий (если нужно). Когда закончите, отправьте /done")
		b.api.Send(msg)
	}
}

func (b *Bot) handleText(message *tgbotapi.Message) {
	userID := message.From.ID
	state := b.userStates[userID]

	if state.Text == "" {
		state.Text = message.Text
		msg := tgbotapi.NewMessage(message.Chat.ID, "Текст принят. Можете отправить фото (до 4) или отправьте /done для завершения.")
		b.api.Send(msg)
	} else {
		state.Text += "\n" + message.Text
	}
}

func (b *Bot) handlePhoto(message *tgbotapi.Message) {
	userID := message.From.ID
	state := b.userStates[userID]

	if state.Text == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Сначала отправьте текст, а затем фотографии.")
		b.api.Send(msg)
		return
	}

	if state.PhotoCount >= 4 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Вы уже прикрепили максимум фотографий (4). Отправьте /done для завершения.")
		b.api.Send(msg)
		return
	}

	photos := message.Photo
	if len(photos) > 0 {
		largestPhoto := photos[len(photos)-1]
		state.Photos = append(state.Photos, largestPhoto.FileID)
		state.PhotoCount++

		msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Фото принято (%d/4). Отправьте ещё фото или /done для завершения.", state.PhotoCount))
		b.api.Send(msg)
	}
}

func (b *Bot) sendToGroup(message *tgbotapi.Message, state *domain.UserState) {
	username := message.From.UserName
	firstName := message.From.FirstName
	userID := message.From.ID

	var typeText string
	if state.State == domain.StateAwaitingProblemReport {
		typeText = "Проблема"
	} else {
		typeText = "Предложение"
	}

	var userInfo string
	if username != "" {
		userInfo = fmt.Sprintf("@%s", username)
	} else {
		userInfo = firstName
	}

	text := fmt.Sprintf("%s от %s:\n\n%s", typeText, userInfo, state.Text)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Перейти к пользователю", fmt.Sprintf("tg://user?id=%d", userID)),
		),
	)

	if len(state.Photos) == 0 {
		msg := tgbotapi.NewMessage(b.groupID, text)
		msg.ReplyMarkup = keyboard
		b.api.Send(msg)
	} else if len(state.Photos) == 1 {
		msg := tgbotapi.NewPhoto(b.groupID, tgbotapi.FileID(state.Photos[0].(string)))
		msg.Caption = text
		msg.ReplyMarkup = keyboard
		b.api.Send(msg)
	} else {
		mediaGroup := make([]interface{}, len(state.Photos))
		for i, photo := range state.Photos {
			media := tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(photo.(string)))
			if i == 0 {
				media.Caption = text
			}
			mediaGroup[i] = media
		}

		mediaGroupMsg := tgbotapi.NewMediaGroup(b.groupID, mediaGroup)
		messages, _ := b.api.SendMediaGroup(mediaGroupMsg)

		if len(messages) > 0 {
			msg := tgbotapi.NewMessage(b.groupID, fmt.Sprintf("%s", strings.Split(text, "\n\n")[0]))
			msg.ReplyToMessageID = messages[0].MessageID
			msg.ReplyMarkup = keyboard
			b.api.Send(msg)
		}
	}
}
