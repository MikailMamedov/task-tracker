package main

// Task описывает модель задачи в нашей системе
type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	DueDate   string `json:"due_date"`
	CreatedAt string `json:"created_at"`
}

// CreateTaskInput используется для парсинга и валидации входящего JSON при создании
type CreateTaskInput struct {
	Title   string `json:"title" binding:"required"`
	Status  string `json:"status"`
	DueDate string `json:"due_date"`
}

// UpdateTaskInput использует указатели (*string), чтобы мы могли отличить:
// 1. Поле вообще не передали в JSON (будет nil)
// 2. Поле передали пустым (будет указатель на пустую строку "")
type UpdateTaskInput struct {
	Title   *string `json:"title"`
	Status  *string `json:"status"`
	DueDate *string `json:"due_date"`
}