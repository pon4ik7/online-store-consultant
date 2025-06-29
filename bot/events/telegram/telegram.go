package telegram

import (
	"bot/clients/telegram"
	"bot/events"
	"bot/lib/e"
	"errors"
)

type Processor struct {
	tg     *telegram.Client
	offset int
}

type Meta struct {
	ChatID   int
	Username string
}

var (
	ErrUnknownEventType = errors.New("unknown event type")
	ErrUnknownMetaType  = errors.New("unknown meta type")
)

func New(client *telegram.Client) *Processor {
	return &Processor{
		tg: client,
	}
}

func (p *Processor) Fetch(limit int) ([]events.Event, error) {
	updates, err := p.tg.Updates(p.offset, limit)
	if err != nil {
		return nil, e.Wrap("can't get events", err)
	}

	if len(updates) == 0 {
		return nil, nil
	}

	res := make([]events.Event, 0, len(updates))

	for _, u := range updates {
		res = append(res, event(u))
	}

	p.offset = updates[len(updates)-1].ID + 1

	return res, nil
}

func (p *Processor) processCallback(event events.Event) error {
	meta, err := meta(event)
	if err != nil {
		return e.Wrap("can't process callback", err)
	}

	// Обязательно ответить на callback, чтобы убрать "часики" в клиенте Telegram
	if err := p.tg.AnswerCallbackQuery(event.CallbackQueryID, "Спасибо за вашу оценку!"); err != nil {
		return err
	}

	if err := p.tg.DeleteMessage(meta.ChatID, event.MessageID); err != nil {
		return e.Wrap("can't delete message", err)
	}

	// Убрать inline-клавиатуру, чтобы кнопки пропали и нельзя было нажать повторно
	err = p.tg.EditMessageReplyMarkup(meta.ChatID, event.MessageID, telegram.InlineKeyboardMarkup{InlineKeyboard: [][]telegram.InlineKeyboardButton{}})
	if err != nil {
		return e.Wrap("can't remove inline keyboard", err)
	}

	switch event.Text {
	case "1", "2", "3", "4", "5":
		return p.sendFeedback(meta.ChatID)
	default:
		return ErrUnknownEventType
	}
}

func (p *Processor) Process(event events.Event) error {
	switch event.Type {
	case events.Message:
		return p.processMessage(event)
	case events.CallbackQuery:
		return p.processCallback(event)
	default:
		return e.Wrap("can't process message", ErrUnknownEventType)
	}
}

func (p *Processor) processMessage(event events.Event) error {
	meta, err := meta(event)
	if err != nil {
		return e.Wrap("can't process message", err)
	}

	if err := p.doCmd(event.Text, meta.ChatID); err != nil {
		return e.Wrap("can't process message", err)
	}

	return nil
}

func meta(event events.Event) (Meta, error) {
	res, ok := event.Meta.(Meta)
	if !ok {
		return Meta{}, e.Wrap("can't get meta", ErrUnknownMetaType)
	}

	return res, nil
}

func event(upd telegram.Update) events.Event {
	if upd.CallbackQuery != nil {
		return events.Event{
			Type: events.CallbackQuery,
			Text: upd.CallbackQuery.Data,
			Meta: Meta{
				ChatID:   upd.CallbackQuery.Message.Chat.ID,
				Username: upd.CallbackQuery.From.Username,
			},
			MessageID: upd.CallbackQuery.Message.MessageID,
		}
	}

	updType := fetchType(upd)

	res := events.Event{
		Type: updType,
		Text: fetchText(upd),
	}

	if updType == events.Message {
		res.Meta = Meta{
			ChatID:   upd.Message.Chat.ID,
			Username: upd.Message.From.Username,
		}
	}

	return res
}

func fetchText(upd telegram.Update) string {
	if upd.Message == nil {
		return ""
	}

	return upd.Message.Text
}

func fetchType(upd telegram.Update) events.Type {
	if upd.Message == nil {
		return events.Unknown
	}

	return events.Message
}
