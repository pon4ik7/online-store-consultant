package telegram

type UpdatesResponse struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result"`
}

type Update struct {
	ID            int              `json:"update_id"`
	Message       *IncomingMessage `json:"message"`
	CallbackQuery *CallbackQuery   `json:"callback_query,omitempty"`
}

type CallbackQuery struct {
	ID      string           `json:"id"`
	From    *User            `json:"from"`
	Message *IncomingMessage `json:"message"`
	Data    string           `json:"data"`
}

type User struct {
	Username string `json:"username"`
}

type IncomingMessage struct {
	Text      string `json:"text"`
	From      From   `json:"from"`
	Chat      Chat   `json:"chat"`
	MessageID int    `json:"message_id"`
}

type From struct {
	Username string `json:"username"`
}

type Chat struct {
	ID int `json:"id"`
}
