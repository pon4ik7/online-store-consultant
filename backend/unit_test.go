package main

import (
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// Now these variables are used only in tests - added DB-logic instead
var (
	sessionStore = make(map[string]Session)
	storeMu      sync.Mutex // For locking/unlocking sessionStore

	registeredClientsSessions = make(map[string]Session)
	logStoreMu                sync.RWMutex
)

// Function to get session with an authorized user
func getAuthorizedSession(w http.ResponseWriter, key string) (Session, error) {
	logStoreMu.Lock()
	session, exists := registeredClientsSessions[key]
	logStoreMu.Unlock()
	if !exists {
		log.Printf("No attached session for an authorized client %s exist", key)
		return Session{}, ErrNotExistingUser
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "isRegistered",
		Value:    "true",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
	})

	return session, nil
}

func TestCreateAuthorizedSession(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Creating mock SQL connection failed: %v\n", err)
	}
	defer mockDB.Close()
	db = mockDB

	// Expect the INSERT
	mock.ExpectExec("INSERT INTO sessions").
		WithArgs(sqlmock.AnyArg(), "", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect the CREATE TABLE
	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS "session_messages_.*"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Clean up sessions for testing
	storeMu.Lock()
	sessionStore = make(map[string]Session)
	storeMu.Unlock()

	// Clean up registered users for testing
	logStoreMu.Lock()
	registeredClientsSessions = make(map[string]Session)
	logStoreMu.Unlock()

	recorder := httptest.NewRecorder()
	key := "test_key_123"

	session, err := createAuthorizedSession(recorder, key)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if session.ID == "" {
		t.Errorf("Session ID should not be empty")
	}
	if session.Context != "" {
		t.Errorf("Session Context should be empty")
	}
	if time.Since(session.LastActive) > (1 * time.Second) {
		t.Errorf("Session LastActive should be less than 1s ago")
	}
	if !session.isRegistered {
		t.Errorf("Session should be marked as registered")
	}

	// Check the registeredClientsSessions
	logStoreMu.Lock()
	_, ok := registeredClientsSessions[key]
	logStoreMu.Unlock()
	if !ok {
		t.Errorf("Session should be stored in registeredClientsSessions")
	}

	// Check cookies
	cookies := recorder.Result().Cookies()
	if len(cookies) == 0 {
		t.Errorf("Cookie should not be empty")
	}
	if cookies[0].Name != "session_id" {
		t.Errorf("Cookie Name should be 'session_id'")
	}
	if cookies[0].Value != session.ID {
		t.Errorf("Cookie Value should be equal to session_id, but:  %s != %s", cookies[0].Value, session.ID)
	}

	// Check if all the expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Не все ожидания от mock DB выполнены: %v", err)
	}
}

func TestGetInitialSession_NewSession(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Creating mock SQL connection failed: %v\n", err)
	}
	defer mockDB.Close()
	db = mockDB

	// Expect INSERT
	mock.ExpectExec("INSERT INTO sessions").
		WithArgs(sqlmock.AnyArg(), "", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect CREATE TABLE
	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS "session_messages_.*"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Clean up the sessionStore
	storeMu.Lock()
	sessionStore = make(map[string]Session)
	storeMu.Unlock()

	// New user
	req := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	session := getInitialSession(recorder, req)

	if session.ID == "" {
		t.Errorf("Session ID should not be empty")
	}
	if session.Context != "" {
		t.Errorf("Session Context should be empty")
	}
	if time.Since(session.LastActive) > 1*time.Second {
		t.Errorf("Session LastActive should be less than 1s ago")
	}

	// Check the session is in the sessionStore
	storeMu.Lock()
	_, ok := sessionStore[session.ID]
	storeMu.Unlock()
	if !ok {
		t.Errorf("Session should be stored in sessionStore")
	}

	// Check cookies
	cookies := recorder.Result().Cookies()
	if len(cookies) == 0 {
		t.Errorf("Cookie should not be empty")
	}
	if cookies[0].Name != "session_id" {
		t.Errorf("Cookie name should be 'session_id'")
	}
	if cookies[0].Value != session.ID {
		t.Errorf("Cookie value should match session ID")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Не все ожидания от mock DB выполнены: %v", err)
	}
}

func TestCreateNewInitialSession(t *testing.T) {
	// Mock the database
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Creating mock SQL connection failed: %v", err)
	}
	defer mockDB.Close()
	db = mockDB

	// Expect INSERT in sessions
	mock.ExpectExec("INSERT INTO sessions").
		WithArgs(sqlmock.AnyArg(), "", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect creating the table for messages
	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS "session_messages_.*"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Cleen up the sessionStore
	storeMu.Lock()
	sessionStore = make(map[string]Session)
	storeMu.Unlock()

	recorder := httptest.NewRecorder()

	session := createNewAnonymousSession(recorder)

	// Check the Session
	if session.ID == "" {
		t.Errorf("Session ID should not be empty")
	}
	if session.Context != "" {
		t.Errorf("Session Context should be empty")
	}
	if time.Since(session.LastActive) > (1 * time.Second) {
		t.Errorf("Session LastActive should be recent")
	}
	if session.isRegistered != false {
		t.Errorf("Session should not be registered by default")
	}

	// Check that session is in в sessionStore
	storeMu.Lock()
	_, ok := sessionStore[session.ID]
	storeMu.Unlock()
	if !ok {
		t.Errorf("Session should be stored in sessionStore")
	}

	// Check the cookies
	cookies := recorder.Result().Cookies()
	foundSessionID := false
	foundIsRegistered := false

	for _, c := range cookies {
		if c.Name == "session_id" && c.Value == session.ID {
			foundSessionID = true
		}
		if c.Name == "isRegistered" && c.Value == "false" {
			foundIsRegistered = true
		}
	}
	if !foundSessionID {
		t.Errorf("Cookie 'session_id' with session ID not set correctly")
	}
	if !foundIsRegistered {
		t.Errorf("Cookie 'isRegistered' should be set to 'false'")
	}

	// // Check if all the expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Не все ожидания от mock DB выполнены: %v", err)
	}
}

func TestGetAuthorizedSession_Exist(t *testing.T) {
	key := "test_key_123"
	testSession := Session{
		ID:           "session_id_123",
		LastActive:   time.Now(),
		isRegistered: true,
		Context:      "",
	}
	logStoreMu.Lock()
	registeredClientsSessions = make(map[string]Session)
	registeredClientsSessions[key] = testSession
	logStoreMu.Unlock()

	recorder := httptest.NewRecorder()
	session, err := getAuthorizedSession(recorder, key)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if session.ID != testSession.ID {
		t.Errorf("Session ID should be equal to testSession.ID")
	}
	if !session.isRegistered {
		t.Errorf("Session should be marked as registered")
	}

	cookies := recorder.Result().Cookies()
	if cookies[0].Name != "session_id" || cookies[0].Value != testSession.ID {
		t.Errorf("Expected 'session_id' cookie to be set	")
	}
	if cookies[1].Name != "isRegistered" || cookies[1].Value != "true" {
		t.Errorf("Expected 'isRegistered' cookie to be set to 'true'")
	}
}

func TestGetAuthorizedSession_NotExists(t *testing.T) {
	key := "non_existing_key"

	logStoreMu.Lock()
	registeredClientsSessions = make(map[string]Session)
	logStoreMu.Unlock()

	recorder := httptest.NewRecorder()

	session, err := getAuthorizedSession(recorder, key)
	if err != ErrNotExistingUser {
		t.Fatalf("Expected ErrNotExistingUser, got: %v", err)
	}

	if session != (Session{}) {
		t.Errorf("Expected empty session on error, got: %+v", session)
	}

	// Check for cookies' absence
	cookies := recorder.Result().Cookies()
	if len(cookies) > 0 {
		t.Errorf("No cookies should be set when session doesn't exist")
	}
}

func TestIsRegistered(t *testing.T) {
	req1 := httptest.NewRequest("GET", "/", nil)
	req1.AddCookie(&http.Cookie{Name: "isRegistered", Value: "true"})

	ok, err := isRegistered(req1)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !ok {
		t.Errorf("Expected isRegistered to be true")
	}

	req2 := httptest.NewRequest("GET", "/", nil)
	req2.AddCookie(&http.Cookie{Name: "isRegistered", Value: "false"})

	ok, err = isRegistered(req2)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	req3 := httptest.NewRequest("GET", "/", nil)

	ok, err = isRegistered(req3)
	if err != ErrNotExistingUser {
		t.Errorf("Expected ErrNotExistingUser, got: %v", err)
	}
	if ok {
		t.Errorf("Expected isRegistered to be false when cookie is missing")
	}
}

func TestDeleteAnonymousSession_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer mockDB.Close()
	db = mockDB

	sessionID := "abc123"

	// Ожидаем DROP TABLE
	mock.ExpectExec(`DROP TABLE IF EXISTS "anonymous_messages_abc123"`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Ожидаем DELETE FROM anonymous_sessions
	mock.ExpectExec(`DELETE FROM anonymous_sessions WHERE session_id = \$1`).
		WithArgs(sessionID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = deleteAnonymousSession(sessionID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("not all expectations were met: %v", err)
	}
}

func TestSaveUserMessage_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer mockDB.Close()
	db = mockDB

	sessionID := "abc123"
	userMsg := "Привет"
	response := "Здравствуйте!"

	expectedQuery := fmt.Sprintf(`INSERT INTO "user_messages_%s"`, sessionID)

	mock.ExpectExec(expectedQuery).
		WithArgs(userMsg, response).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = saveUserMessage(sessionID, userMsg, response)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("not all expectations were met: %v", err)
	}
}
