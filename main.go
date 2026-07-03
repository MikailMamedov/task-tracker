package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/glebarez/go-sqlite"
)

var db *sql.DB

func main() {
	var err error

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./tasks.db"
	}

	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer db.Close()

	// Initialize database schema if it doesn't exist
	createTableSQL := `CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		status TEXT NOT NULL,
		due_date TEXT,
		created_at TEXT
	);`
	if _, err := db.Exec(createTableSQL); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.POST("/tasks", createTask)
	router.GET("/tasks", listTasks)
	router.GET("/tasks/:id", getTask)
	router.DELETE("/tasks/:id", deleteTask)
	router.PATCH("/tasks/:id", updateTask)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router.Run(":" + port)
}

func createTask(c *gin.Context) {
	var input CreateTaskInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: title is required"})
		return
	}

	if len(input.Title) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is too long (max 200 chars)"})
		return
	}
	if input.Status != "" && input.Status != "pending" && input.Status != "done" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status must be either 'pending' or 'done'"})
		return
	}
	if input.Status == "done" && input.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A task cannot be marked 'done' if it has no title"})
		return
	}

	status := input.Status
	if status == "" {
		status = "pending"
	}

	createdAt := time.Now().Format(time.RFC3339)

	query := `INSERT INTO tasks (title, status, due_date, created_at) VALUES (?, ?, ?, ?)`
	result, err := db.Exec(query, input.Title, status, input.DueDate, createdAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save to DB"})
		return
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve primary key"})
		return
	}

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
	status := c.Query("status")

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	query := `SELECT id, title, status, due_date, created_at FROM tasks`
	var args []interface{}

	if status != "" {
		query += ` WHERE status = ?`
		args = append(args, status)
	}

	query += ` ORDER BY created_at DESC`
	query += ` LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}
	defer rows.Close()

	var tasksList = []Task{}

	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.Title, &t.Status, &t.DueDate, &t.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning data"})
			return
		}
		tasksList = append(tasksList, t)
	}

	c.JSON(http.StatusOK, tasksList)
}

func getTask(c *gin.Context) {
	id := c.Param("id")
	var t Task

	query := `SELECT id, title, status, due_date, created_at FROM tasks WHERE id = ?`
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

func deleteTask(c *gin.Context) {
	id := c.Param("id")

	var existsID int
	err := db.QueryRow(`SELECT id FROM tasks WHERE id = ?`, id).Scan(&existsID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	_, err = db.Exec(`DELETE FROM tasks WHERE id = ?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

func updateTask(c *gin.Context) {
	id := c.Param("id")
	var input UpdateTaskInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var current Task
	querySelect := `SELECT id, title, status, due_date, created_at FROM tasks WHERE id = ?`
	err := db.QueryRow(querySelect, id).Scan(&current.ID, &current.Title, &current.Status, &current.DueDate, &current.CreatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Merge patch inputs with current state
	if input.Title != nil {
		current.Title = *input.Title
	}
	if input.Status != nil {
		current.Status = *input.Status
	}
	if input.DueDate != nil {
		current.DueDate = *input.DueDate
	}

	if len(current.Title) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title can't exceed 200 characters"})
		return
	}
	if current.Status != "pending" && current.Status != "done" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status must be 'pending' or 'done'"})
		return
	}
	if current.Status == "done" && current.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A task cannot be marked 'done' if it has no title"})
		return
	}

	queryUpdate := `UPDATE tasks SET title = ?, status = ?, due_date = ? WHERE id = ?`
	_, err = db.Exec(queryUpdate, current.Title, current.Status, current.DueDate, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	c.JSON(http.StatusOK, current)
}