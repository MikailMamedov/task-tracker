package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/glebarez/go-sqlite"
)

// var tasks = []Task{}
// var nextID = 1

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

	// // Инкрементируем ID для следующей задачи
	// nextID++

	// // Добавляем новинку в наш глобальный список задач
	// tasks = append(tasks, newTask)

	// // Возвращаем созданную таску со статусом 201 Created
	// c.JSON(http.StatusCreated, newTask)

	c.JSON(http.StatusCreated, newTask)
}

// Handler для получения списка задач
// func listTasks(c *gin.Context) {
// 	// Просто отдаем весь наш массив в формате JSON
// 	c.JSON(http.StatusOK, tasks)
// }

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
