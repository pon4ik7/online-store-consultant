package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5"
	"log"
	"net/http"
	"os"
)

var db *sql.DB

func init() {
	if os.Getenv("DISABLE_INIT") == "1" { // Disable init in case DISABLE_INIT == 1,
		return // DISABLE_INIT = 1 is set at the start of testing
	} // (см. unit_test.go)

	var err error
	const connection = "postgres://radat:radatSWP25@localhost:5432/radatDB?sslmode=disable"
	db, err = sql.Open("postgres", connection)
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}
	runMigrations()
	populateTestData()
}

func runMigrations() {
	migrationPath := "file://migrations"
	const connection = "postgres://radat:radatSWP25@localhost:5432/radatDB?sslmode=disable"
	m, err := migrate.New(migrationPath, connection)
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("Error applying migrations: ", err)
	}
	log.Println("Migrations have been successfully applied!")
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte("Добро пожаловать на главную страницу!"))
	})
	http.HandleFunc("/api/products/add", addProduct)
	http.HandleFunc("/api/sessions/create", createSession)
	http.HandleFunc("/api/sessions/messages/add", addMessage)

	log.Println("Сервер запущен на :8081")
	sessionID := uuid.New().String()
	context := "Новый контекст сессии"
	SaveDialogueContext(sessionID, context, db)
	messagesCache := make(map[string]map[string]string)
	messagesCache["sessionID1"] = map[string]string{
		"Test": "test",
	}
	result := ClarifyProductContext(messagesCache, "sessionID1")
	fmt.Println(result)
	log.Fatal(http.ListenAndServe(":8081", nil))
}

// Several basic DB-requests for testing and future development

func addProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_, err := db.Exec(`
        INSERT INTO popular_products 
        (name, description, price, rating, category, product_url, image_url) 
        VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		"Новый продукт", "Описание продукта", 100.00, 4.5, "Категория", "http://example.com", "http://example.com/image.jpg")
	if err != nil {
		http.Error(w, "Failed to add product: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Product added successfully"))
}

func createSession(w http.ResponseWriter, r *http.Request) {
	sessionID := uuid.New()
	_, err := db.Exec(`
        INSERT INTO sessions 
        (session_id, context) 
        VALUES ($1, $2)`,
		sessionID, "Контекст сессии")
	if err != nil {
		http.Error(w, "Failed to create session: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"session_id": sessionID.String(),
		"status":     "created",
	})
}

func addMessage(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}
	_, err := db.Exec(`
        INSERT INTO session_messages 
        (session_id, message, response) 
        VALUES ($1, $2, $3)`,
		sessionID, "Вопрос пользователя", "Ответ системы")
	if err != nil {
		http.Error(w, "Failed to add message: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Message added successfully"))
}

func populateTestData() {
	insertProduct := `INSERT INTO popular_products (name, description, price, rating, category, product_url, image_url) 
	VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := db.Exec(insertProduct, "Smartphone", "A high-end smartphone with great features", 999.99, 4.7, "Electronics", "http://example.com/product/1", "http://example.com/images/product1.jpg")
	if err != nil {
		log.Printf("Error inserting product: %v", err)
	}
	_, err = db.Exec(insertProduct, "Smartphone Y", "A mid-range smartphone with decent specs", 499.99, 4.2, "Electronics", "http://example.com/product/2", "http://example.com/images/product2.jpg")
	if err != nil {
		log.Printf("Error inserting product: %v", err)
	}
	_, err = db.Exec(insertProduct, "Laptop A", "A powerful laptop for gaming and work", 1500.00, 4.8, "Computers", "http://example.com/product/3", "http://example.com/images/product3.jpg")
	if err != nil {
		log.Printf("Error inserting product: %v", err)
	}
	_, err = db.Exec(insertProduct, "Laptop B", "A budget-friendly laptop for daily tasks", 300.00, 3.8, "Computers", "http://example.com/product/4", "http://example.com/images/product4.jpg")
	if err != nil {
		log.Printf("Error inserting product: %v", err)
	}
	_, err = db.Exec(insertProduct, "Headphones X", "Noise-cancelling headphones with great sound", 150.00, 4.5, "Accessories", "http://example.com/product/5", "http://example.com/images/product5.jpg")
	if err != nil {
		log.Printf("Error inserting product: %v", err)
	}
	log.Println("Test data inserted successfully!")
}
