package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/glebarez/go-sqlite"
)

// setupTestDB initializes an isolated in-memory SQLite database for testing.
func setupTestDB(t *testing.T) {
	var err error
	db, err = sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize test DB: %v", err)
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		status TEXT NOT NULL,
		due_date TEXT,
		created_at TEXT
	);`
	if _, err := db.Exec(createTableSQL); err != nil {
		t.Fatalf("Failed to create test database schema: %v", err)
	}
}

// TestCreateTask_ValidationError enforces the business rule that a task 
// cannot be marked 'done' without a non-empty title.
func TestCreateTask_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	setupTestDB(t)
	defer db.Close()

	r := gin.New()
	r.POST("/tasks", createTask)

	payload := map[string]string{
		"title":  "",
		"status": "done",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 Bad Request, got %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse server response: %v", err)
	}

	if _, exists := response["error"]; !exists {
		t.Error("Expected 'error' key in JSON response payload")
	}
}