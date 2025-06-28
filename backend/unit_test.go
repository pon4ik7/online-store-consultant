package main

import (
	"github.com/DATA-DOG/go-sqlmock"
	"net/http/httptest"
	"testing"
	"time"
)

//func init() {
//	if os.Getenv("DISABLE_INIT") == "1" {
//		return
//	}
//}

// $env:DISABLE_INIT = "1", чтобы отключить init().
// Не 1, чтобы запустить

func TestCreateAuthorizedSession(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Creating mock SQL connection failed: %v\n", err)
	}
	defer mockDB.Close()
	db = mockDB

	// Ожидание INSERT
	mock.ExpectExec("INSERT INTO sessions").
		WithArgs(sqlmock.AnyArg(), "", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Ожидание CREATE TABLE
	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS "session_messages_.*"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Очистка сессий
	storeMu.Lock()
	sessionStore = make(map[string]Session)
	storeMu.Unlock()

	// Очистка зарегистрированных
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

	// Проверка в registeredClientsSessions
	logStoreMu.Lock()
	_, ok := registeredClientsSessions[key]
	logStoreMu.Unlock()
	if !ok {
		t.Errorf("Session should be stored in registeredClientsSessions")
	}

	// Проверка кукисов

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

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Не все ожидания от mock DB выполнены: %v", err)
	}

	// Проверка, что все ожидания mock выполнены
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

	// Ожидание INSERT
	mock.ExpectExec("INSERT INTO sessions").
		WithArgs(sqlmock.AnyArg(), "", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Ожидание CREATE TABLE
	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS "session_messages_.*"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Очистка sessionStore
	storeMu.Lock()
	sessionStore = make(map[string]Session)
	storeMu.Unlock()

	// Запрос без куки (новый пользователь)
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

	// Проверка, что сессия попала в sessionStore
	storeMu.Lock()
	_, ok := sessionStore[session.ID]
	storeMu.Unlock()
	if !ok {
		t.Errorf("Session should be stored in sessionStore")
	}

	// Проверка куки
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
	// Мокаем базу данных
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Creating mock SQL connection failed: %v", err)
	}
	defer mockDB.Close()
	db = mockDB

	// Ожидаем INSERT в sessions
	mock.ExpectExec("INSERT INTO sessions").
		WithArgs(sqlmock.AnyArg(), "", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Ожидаем создание таблицы для сообщений
	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS "session_messages_.*"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Очищаем sessionStore
	storeMu.Lock()
	sessionStore = make(map[string]Session)
	storeMu.Unlock()

	recorder := httptest.NewRecorder()

	session := createNewInitialSession(recorder)

	// Проверки содержимого Session
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

	// Проверка, что сессия попала в sessionStore
	storeMu.Lock()
	_, ok := sessionStore[session.ID]
	storeMu.Unlock()
	if !ok {
		t.Errorf("Session should be stored in sessionStore")
	}

	// Проверка куки
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

	// Проверка, что все mock-ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Не все ожидания от mock DB выполнены: %v", err)
	}
}
