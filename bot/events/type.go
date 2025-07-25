package events

type Fetcher interface {
	Fetch(limit int) ([]Event, error)
}

type Processor interface {
	Process(e Event) error
}

type Type int

const (
	Unknown Type = iota
	Message
	CallbackQuery
)

type Event struct {
	Type            Type
	Text            string
	Meta            interface{}
	MessageID       int
	CallbackQueryID string
}
