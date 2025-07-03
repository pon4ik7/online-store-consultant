package telegram

import (
	"bot/clients/telegram"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strings"
)

const (
	HelpCmd  = "/help"
	StartCmd = "/start"
)

var clients = make(map[int]*http.Client)

func (p *Processor) doCmd(text string, chatID int) error {
	text = strings.TrimSpace(text)

	if text == HelpCmd {
		return p.sendHelp(chatID)
	} else if text == StartCmd {
		jar, _ := cookiejar.New(nil)
		newClient := &http.Client{Jar: jar}
		startReq, _ := http.NewRequest("POST", "http://localhost:8080/api/start", nil)
		startResp, err := newClient.Do(startReq)

		if err != nil {
			log.Printf("Error with /api/start: %v", err)
			return p.sendResponse(chatID, "Возникли технические неполадки, пожалуйста, попробуйте позже")
		}

		defer startResp.Body.Close()

		clients[chatID] = newClient

		return p.sendHello(chatID)
	} else if text == EndCmd {
		if client, ok := clients[chatID]; !ok {
			return p.sendEndInvalid(chatID)
		} else {
			endReq, _ := http.NewRequest("POST", "http://localhost:8080/api/end", nil)
			endResp, err := client.Do(endReq)

			if err != nil {
				log.Printf("Error with /api/end: %v", err)
				return p.sendResponse(chatID, "Возникли технические неполадки, пожалуйста, попробуйте позже")
			}

			defer endResp.Body.Close()
			var response map[string]string
			json.NewDecoder(endResp.Body).Decode(&response)

			delete(clients, chatID)

			return p.sendResponseWithNumberButtons(chatID, response["response"])
		}

	} else {
		if client, ok := clients[chatID]; !ok {
			return p.sendStartInvalid(chatID)
		} else {
			message := map[string]string{
				"message": text,
			}
			jsonData, _ := json.Marshal(message)

			msgReq, _ := http.NewRequest("POST", "http://localhost:8080/api/message", bytes.NewBuffer(jsonData))
			msgReq.Header.Set("Content-Type", "application/json")

			msgResp, err := client.Do(msgReq)
			if err != nil {
				log.Printf("Ошибка при /api/message: %v", err)
				return p.sendResponse(chatID, "Some technical issues have occurred, please, try again later")
			}
			defer msgResp.Body.Close()

			var response map[string]string
			json.NewDecoder(msgResp.Body).Decode(&response)

			return p.sendResponse(chatID, response["response"])
		}
	}

}

func (p *Processor) sendResponseWithNumberButtons(chatID int, text string) error {
	keyboard := [][]telegram.InlineKeyboardButton{
		{
			{Text: "1", CallbackData: "1"},
			{Text: "2", CallbackData: "2"},
			{Text: "3", CallbackData: "3"},
			{Text: "4", CallbackData: "4"},
			{Text: "5", CallbackData: "5"},
		},
	}

	return p.tg.SendMessageWithInlineKeyboard(chatID, text, keyboard)
}

func (p *Processor) sendHelp(chatID int) error {
	return p.tg.SendMessage(chatID, msgHelp)
}

func (p *Processor) sendStartInvalid(chatID int) error {
	return p.tg.SendMessage(chatID, startInvalid)
}

func (p *Processor) sendEndInvalid(chatID int) error {
	return p.tg.SendMessage(chatID, endInvalid)
}

func (p *Processor) sendResponse(chatID int, response string) error {
	return p.tg.SendMessage(chatID, response)
}

func (p *Processor) sendHello(chatID int) error {
	keyboard := [][]telegram.InlineKeyboardButton{
		{
			{Text: "iPhone 13", CallbackData: "p1"},
			{Text: "MacBook Pro 16", CallbackData: "p2"},
			{Text: "Sony WH‑1000XM6", CallbackData: "p3"},
			{Text: "Apple Watch Ultra2", CallbackData: "p4"},
		},
	}

	return p.tg.SendMessageWithInlineKeyboard(chatID, msgHello, keyboard)
}

func (p *Processor) sendFeedback(chatID int) error {
	return p.tg.SendMessage(chatID, msgFeedback)
}
