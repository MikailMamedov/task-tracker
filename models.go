package main

// Task represents the core task database and API model.
type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	DueDate   string `json:"due_date"`
	CreatedAt string `json:"created_at"`
}

// CreateTaskInput defines the schema for parsing and validating incoming task creation requests.
type CreateTaskInput struct {
	Title   string `json:"title" binding:"required"`
	Status  string `json:"status"`
	DueDate string `json:"due_date"`
}

// UpdateTaskInput uses pointers to distinguish between omitted fields (nil)
// and explicitly provided empty strings ("") during partial updates.
type UpdateTaskInput struct {
	Title   *string `json:"title"`
	Status  *string `json:"status"`
	DueDate *string `json:"due_date"`
}