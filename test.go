package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
)

func main() {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Your email: ")
	userLog, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading input:", err)
	}
	fmt.Print("Your password: ")
	userPass, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading input:", err)
	}

	userData := map[string]string{
		"login":    userLog,
		"password": userPass,
	}
	userJsonData, _ := json.Marshal(userData)
	startReq, _ := http.NewRequest("POST", "http://localhost:8080/api/register", bytes.NewBuffer(userJsonData))
	startResp, err := client.Do(startReq)
	if err != nil {
		fmt.Println("Ошибка при /api/start:", err)
		return
	}
	defer startResp.Body.Close()

	for {
		fmt.Print("Your question to AI: ")
		userInput, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading input:", err)
			continue
		}
		fmt.Print("Product ID: ")
		userInputID, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading input:", err)
			continue
		}
		userInput = userInput[:len(userInput)-1]
		if userInput == "" {
			continue
		}
		message := map[string]string{
			"message":   userInput,
			"productID": userInputID,
		}
		jsonData, _ := json.Marshal(message)

		msgReq, _ := http.NewRequest("POST", "http://localhost:8080/api/message", bytes.NewBuffer(jsonData))
		msgReq.Header.Set("Content-Type", "application/json")

		msgResp, err := client.Do(msgReq)
		if err != nil {
			fmt.Println("Ошибка при /api/message:", err)
			return
		}
		defer msgResp.Body.Close()

		var result map[string]string
		json.NewDecoder(msgResp.Body).Decode(&result)
		fmt.Println("Online consultant answer: ", result["response"])
	}
}
