package main

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"log"
)

func SaveDialogueContext(sessionIDStr string, keyWords string, db *sql.DB) {
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
	var productName = "Smartphone X"
	var productCategory = "Electronics"
	var productDescription = "A high-end smartphone with some features"

	var similarProductName string
	var similarProductPrice float64
	var similarProductRating float64
	var similarProductDescription string
	var similarProductURL string
	var similarProductImageURL string
	query := `
		SELECT name, price, rating, description, product_url, image_url
		FROM popular_products
		WHERE category = $1 AND split_part(name, ' ', 1) ILIKE split_part($2, ' ', 1)
		LIMIT 1;`
	err := db.QueryRow(query, productCategory, productName).Scan(&similarProductName, &similarProductPrice, &similarProductRating, &similarProductDescription, &similarProductURL, &similarProductImageURL)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No similar product found for the query: %s in category: %s", productName, productCategory)
		} else {
			log.Printf("Error finding similar product: %v", err)
		}
	} else {
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
