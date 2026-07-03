package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var tasks = []Task{}
var nextID = 1

func main() {
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

	// Собираем полноценную задачу из входных данных
	newTask := Task{
		ID:        nextID,
		Title:     input.Title,
		Status:    status,
		DueDate:   input.DueDate,
		CreatedAt: time.Now().Format(time.RFC3339), // Формат времени "2006-01-02T15:04:05Z"
	}

	// Инкрементируем ID для следующей задачи
	nextID++

	// Добавляем новинку в наш глобальный список задач
	tasks = append(tasks, newTask)

	// Возвращаем созданную таску со статусом 201 Created
	c.JSON(http.StatusCreated, newTask)
}

// Handler для получения списка задач
func listTasks(c *gin.Context) {
	// Просто отдаем весь наш массив в формате JSON
	c.JSON(http.StatusOK, tasks)
}
