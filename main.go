package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Инициализируем базовый роутер Gin (он сразу включает логирование запросов)
	router := gin.Default()

	// Наш тестовый эндпоинт для проверки работоспособности
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Запускаем сервер на порту 8080
	router.Run(":8080")
}