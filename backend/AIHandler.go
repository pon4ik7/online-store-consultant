package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

// UserMessage - a structure to save the user query and send to the DeepSeek
type UserMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// The AIRequest DeepSeek would handle
type AIRequest struct {
	Model    string        `json:"model"`
	Messages []UserMessage `json:"messages"`
}

// The Response DeepSeek returns
type Response struct {
	Choices []struct {
		ResponseContent struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func HandleUserQuery(messagesCache map[string]map[string]string, query string, isAdmin bool, sessionID string) (string, error) {
	initialPrompt := query
	query = ClarifyProductContext(messagesCache, sessionID) + query
	response, err := GetResponse(query, isAdmin)
	if err != nil {
		log.Println(err)
		return "", err
	} else {
		cacheMessage(messagesCache, initialPrompt, response, sessionID)
	}
	return response, nil
}

func ClarifyProductContext(messagesCache map[string]map[string]string, sessionID string) string {
	instructions := "You are a friendly store consultant helping a customer. Follow these rules: \n" +
		"Always address the customer with formal \"Вы\" (Russian) or \"you\" in a respectful tone (English)\n" +
		"Respond in the same language as the QUESTION:\n " +
		"Use professional but warm language:  \n " +
		"Avoid robotic phrases (\"Based on your query...\")\n" +
		"Treat the CONTEXT: as our prior conversation history\n" +
		"Acknowledge past discussions naturally\n" +
		"Reference prior interactions if appropriate\n" +
		"For complex queries, offer step-by-step guidance (I recommend checking the size first, then I’ll assist with payment)\n" +
		"Use polite fillers\n" +
		"Prioritize clarity and empathy\n" +
		"Answer briefly but concise and meaningful\n" +
		"Do not answer the questions that are not asked\n" +
		"Greet the customer only once do not use \"Здравствуйте\" and Hello each message\n" +
		"If the QUESTION: is unclear, ask for details like a human would\n CONTEXT:"

	// The data about the product that the user is asking about - it must be obtained using HTTP-requests.
	// TODO: Use requests to get information about the current product
	var productName = ""
	var productCategory = ""
	var productDescription = ""

	var similarProductName string
	var similarProductPrice float64
	var similarProductRating float64
	var similarProductDescription string
	var similarProductURL string
	var similarProductImageURL string

	// This query looks through the saved popular products from the website: first it searches for the same category,
	// then it looks for a match with the first word from the name
	query := `
		SELECT name, price, rating, description, product_url, image_url
		FROM popular_products
		WHERE category = $1 AND split_part(name, ' ', 1) ILIKE split_part($2, ' ', 1)
		LIMIT 1;`
	err := db.QueryRow(query, productCategory, "%"+productName+"%").Scan(&similarProductName, &similarProductPrice,
		&similarProductRating, &similarProductDescription, &similarProductURL, &similarProductImageURL)

	// If there are no similar products, nothing will be added to the message for DeepSeek, just logs are displayed
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No similar product found for the query: %s in category: %s", productName, productCategory)
		} else {
			log.Printf("Error finding similar product: %v", err)
		}
	} else {
		// Checks if we found the same product
		if similarProductName == productName && similarProductDescription == productDescription {
			instructions += "\nThis product is popular on our website."
		} else {
			instructions += "\nI found a similar popular product for you: " + similarProductName + "\n" +
				"Category: " + productCategory + "\n" +
				"Price: $" + fmt.Sprintf("%.2f", similarProductPrice) + "\n" +
				"Rating: " + fmt.Sprintf("%.2f", similarProductRating) + "\n" +
				"Description: " + similarProductDescription + "\n" +
				"Product link: " + similarProductURL + "\n" +
				"Image: " + similarProductImageURL
		}
	}
	for query, response := range messagesCache[sessionID] {
		concat := query + ":" + response
		instructions += "\n" + concat
	}
	return instructions + "QUESTION: "
}

func GetResponse(query string, isAdmin bool) (string, error) {
	var role string
	if isAdmin {
		role = "system"
	} else {
		role = "user"
	}

	request := AIRequest{
		Model: "deepseek-chat", //deepseek-reasoner for complex topics
		Messages: []UserMessage{
			{
				Role:    role,
				Content: query,
			},
		},
	}

	jsonEncoded, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to encode request: %v", err)
	}

	httpReq, err := http.NewRequest(
		"POST",
		"https://api.deepseek.com/v1/chat/completions",
		bytes.NewBuffer(jsonEncoded),
	)

	if err != nil {
		return "", fmt.Errorf("failed to create html request: %v", err)
	}

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey, exists := os.LookupEnv("API_KEY")
	if !exists {
		log.Fatal("API_KEY environment variable not set")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	response, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("API request failed: %v", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("DeepSeek API returned status: %d", response.StatusCode)
	}

	var aiReply Response
	if err := json.NewDecoder(response.Body).Decode(&aiReply); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if len(aiReply.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return aiReply.Choices[0].ResponseContent.Content, nil
}

func cacheMessage(messagesCache map[string]map[string]string, query string, response string, sessionID string) {
	if _, ok := messagesCache[sessionID]; !ok {
		messagesCache[sessionID] = make(map[string]string)
	}
	messagesCache[sessionID][query] = response
}

func SaveDialogueContext(sessionIDStr string, keyWords string, db *sql.DB) {
	// This request works at the end of the session - it preserves its context
	sessionID, err := uuid.Parse(sessionIDStr)
	_, err = db.Exec(`
		INSERT INTO sessions (session_id, context)
		VALUES ($1, $2)
		ON CONFLICT (session_id) DO UPDATE 
		SET context = EXCLUDED.context;
	`, sessionID, keyWords)
	if err != nil {
		log.Fatalf("Failed to save dialogue context for session %s: %v", sessionID, err)
		return
	}
	log.Printf("Context for session %s has been saved successfully", sessionID)
}
