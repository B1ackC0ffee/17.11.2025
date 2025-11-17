package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Инициализируем всё что нужно
	fmt.Println("Запускаем сервис проверки ссылок...")

	// Настройка хранилища
	storage := NewStorage()
	defer storage.Close()

	// Инициализируем менеджер проверок
	checker := NewLinkChecker(storage)
	checker.Start()
	defer checker.Stop()

	// Восстанавливаем задачи после перезапуска
	go checker.RecoverTasks()

	// Настраиваем Gin
	router := setupRouter(storage, checker)

	// Запускаем сервер
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Запуск в отдельной горутине
	go func() {
		log.Println("Сервер запущен на порту 8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	// Ждем сигнал завершения
	waitForShutdown(server, checker)
}

// Настройка маршрутов
func setupRouter(storage *Storage, checker *LinkChecker) *gin.Engine {
	router := gin.Default()

	// Группа API
	api := router.Group("/api/v1")
	{
		api.POST("/check", func(c *gin.Context) {
			handleCheckLinks(c, storage, checker)
		})
		api.POST("/report", func(c *gin.Context) {
			handleReport(c, storage)
		})
	}

	// Проверка здоровья
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "timestamp": time.Now()})
	})

	return router
}

// Ожидание graceful shutdown
func waitForShutdown(server *http.Server, checker *LinkChecker) {
	// Канал для системных сигналов
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)

	// Ждем сигнал
	<-quit
	log.Println(" Получен сигнал завершения...")

	// Даем время на завершение операций
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Останавливаем сервер
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Ошибка при остановке сервера: %v", err)
	}

	log.Println(" Сервер остановлен корректно")
}
