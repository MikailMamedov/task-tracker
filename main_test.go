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

// setupTestDB инициализирует чистую БД в оперативной памяти для каждого теста
func setupTestDB(t *testing.T) {
	var err error
	// Используем ":memory:", чтобы тесты не трогали реальный файл базы данных
	db, err = sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Не удалось запустить тестовую БД: %v", err)
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		status TEXT NOT NULL,
		due_date TEXT,
		created_at TEXT
	);`
	if _, err := db.Exec(createTableSQL); err != nil {
		t.Fatalf("Не удалось создать таблицу в тестовой БД: %v", err)
	}
}

// TestCreateTask_ValidationError проверяет правило ТЗ:
// Задача не может быть со статусом 'done', если у неё пустой заголовок.
func TestCreateTask_ValidationError(t *testing.T) {
	// Переводим Gin в режим тестирования, чтобы он не спамил лишними логами
	gin.SetMode(gin.TestMode)
	
	setupTestDB(t)
	defer db.Close()

	// Инициализируем изолированный роутер и регистрируем хендлер
	r := gin.New()
	r.POST("/tasks", createTask)

	// Готовим заведомо некорректный payload (status = done, но title пустой)
	payload := map[string]string{
		"title":  "",
		"status": "done",
	}
	body, _ := json.Marshal(payload)

	// Имитируем реальный HTTP-запрос через стандартный httptest
	req, _ := http.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Передаем запрос в роутер
	r.ServeHTTP(w, req)

	// Проверяем утверждения (Assertions)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Ожидали статус 400 Bad Request, но получили %d", w.Code)
	}

	// Дополнительно проверяем, что сервер вернул ошибку в формате JSON
	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Не удалось распарсить ответ сервера: %v", err)
	}

	if _, exists := response["error"]; !exists {
		t.Error("Ожидали поле 'error' в ответе JSON, но его там нет")
	}
}