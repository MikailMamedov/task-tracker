package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/glebarez/go-sqlite"
)


var db *sql.DB

func main() {
	var err error

	// 1. Открываем файл базы данных (если файла нет, Go сам его создаст)
	db, err = sql.Open("sqlite", "./tasks.db")
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer db.Close() // Закроет базу, когда сервер остановится

	// 2. Выполняем SQL-запрос, чтобы создать таблицу, если её еще нет
	createTableSQL := `CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		status TEXT NOT NULL,
		due_date TEXT,
		created_at TEXT
	);`
	if _, err := db.Exec(createTableSQL); err != nil {
		log.Fatalf("Ошибка создания таблицы: %v", err)
	}

	// Инициализируем базовый роутер Gin (он сразу включает логирование запросов)
	router := gin.Default()

	// Наш тестовый эндпоинт для проверки работоспособности
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Эндпоинт для получения всех задач
	router.POST("/tasks", createTask)
	router.GET("/tasks", listTasks)

	router.GET("/tasks/:id", getTask)      // Получить одну таску по ID
	router.DELETE("/tasks/:id", deleteTask)   // Удалить таску по ID
	router.PATCH("/tasks/:id", updateTask) // Частично обновить таску

	// Запускаем сервер на порту 8080
	router.Run(":8080")
}

func createTask(c *gin.Context) {
	var input CreateTaskInput

	// .ShouldBindJSON() читает body запроса и пытается перелить данные в структуру input.
	// Если в JSON нет поля "title" (которое у нас binding:"required"), Gin сразу вернет ошибку.
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: title is required"})
		return
	}

	// 🟢 ДОБАВЛЯЕМ ВАЛИДАЦИЮ ПРИ СОЗДАНИИ:
	if len(input.Title) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is too long (max 200 chars)"})
		return
	}
	if input.Status != "" && input.Status != "pending" && input.Status != "done" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status must be either 'pending' or 'done'"})
		return
	}
	// Бизнес-правило: нельзя закрыть таску, если у неё пустой тайтл
	if input.Status == "done" && input.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A task cannot be marked 'done' if it has no title"})
		return
	}

	// Выставляем статус по умолчанию (pending), если он пришел пустым
	status := input.Status
	if status == "" {
		status = "pending"
	}

	createdAt := time.Now().Format(time.RFC3339)

	// Вместо append пишем SQL-запрос INSERT
	query := `INSERT INTO tasks (title, status, due_date, created_at) VALUES (?, ?, ?, ?)`
	result, err := db.Exec(query, input.Title, status, input.DueDate, createdAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save to DB"})
		return
	}

	// Спрашиваем у базы, какой ID она присвоила этой задаче
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get ID"})
		return
	}

	// Собираем ответ клиенту, используя ID из базы
	newTask := Task{
		ID:        int(lastInsertID),
		Title:     input.Title,
		Status:    status,
		DueDate:   input.DueDate,
		CreatedAt: createdAt,
	}

	c.JSON(http.StatusCreated, newTask)
}


func listTasks(c *gin.Context) {
	// 1. Делаем запрос SELECT в базу данных
	query := `SELECT id, title, status, due_date, created_at FROM tasks ORDER BY created_at DESC`
	rows, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}
	defer rows.Close() // Чистим за собой память

	var tasksList = []Task{} // Создаем пустой срез, куда будем складывать строки

	// 2. Бежим циклом по всем строкам, которые вернула база
	for rows.Next() {
		var t Task
		// rows.Scan по очереди берет значения из колонок SQL и записывает в поля структуры t
		err := rows.Scan(&t.ID, &t.Title, &t.Status, &t.DueDate, &t.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning data"})
			return
		}
		tasksList = append(tasksList, t) // Добавляем заполненную таску в наш список
	}

	// 3. Отдаем клиенту заполненный список
	c.JSON(http.StatusOK, tasksList)
}

// 1. GET /tasks/:id — Получение конкретной задачи
func getTask(c *gin.Context) {
	id := c.Param("id") // Вытаскиваем id из URL
	var t Task

	query := `SELECT id, title, status, due_date, created_at FROM tasks WHERE id = ?`
	// QueryRow используется, когда мы ищем строго одну строку.
	err := db.QueryRow(query, id).Scan(&t.ID, &t.Title, &t.Status, &t.DueDate, &t.CreatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, t)
}

// 2. DELETE /tasks/:id — Удаление задачи
func deleteTask(c *gin.Context) {
	id := c.Param("id")

	// Проверяем, существует ли вообще такая таска
	var existsID int
	err := db.QueryRow(`SELECT id FROM tasks WHERE id = ?`, id).Scan(&existsID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Удаляем
	_, err = db.Exec(`DELETE FROM tasks WHERE id = ?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	// По ТЗ при успешном удалении можно вернуть просто статус 200 OK
	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// 3. PATCH /tasks/:id — Частичное обновление задачи
func updateTask(c *gin.Context) {
	id := c.Param("id")
	var input UpdateTaskInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Сначала вытаскиваем текущее состояние задачи из базы
	var current Task
	querySelect := `SELECT id, title, status, due_date, created_at FROM tasks WHERE id = ?`
	err := db.QueryRow(querySelect, id).Scan(&current.ID, &current.Title, &current.Status, &current.DueDate, &current.CreatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Магия указателей: если поле пришло (не nil), обновляем его. Если nil — оставляем старое значение из базы.
	if input.Title != nil {
		current.Title = *input.Title
	}
	if input.Status != nil {
		current.Status = *input.Status
	}
	if input.DueDate != nil {
		current.DueDate = *input.DueDate
	}

	// Валидируем то, что получилось после слияния старых и новых данных
	if len(current.Title) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title can't exceed 200 characters"})
		return
	}
	if current.Status != "pending" && current.Status != "done" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status must be 'pending' or 'done'"})
		return
	}
	// Проверяем наше критическое правило ТЗ
	if current.Status == "done" && current.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A task cannot be marked 'done' if it has no title"})
		return
	}

	// Сохраняем обновленные данные обратно в базу
	queryUpdate := `UPDATE tasks SET title = ?, status = ?, due_date = ? WHERE id = ?`
	_, err = db.Exec(queryUpdate, current.Title, current.Status, current.DueDate, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	c.JSON(http.StatusOK, current)
}