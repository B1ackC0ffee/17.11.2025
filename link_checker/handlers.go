package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Структуры для запросов и ответов
type CheckRequest struct {
	Links []string `json:"links" binding:"required"`
}

type CheckResponse struct {
	Links    map[string]string `json:"links"`
	LinksNum int               `json:"links_num"`
}

type ReportRequest struct {
	LinksList []int `json:"links_list" binding:"required"`
}

// Обработчик проверки ссылок
func handleCheckLinks(c *gin.Context, storage *Storage, checker *LinkChecker) {
	var req CheckRequest

	// Парсим входящий JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Некорректный запрос. Проверьте формат данных",
		})
		return
	}

	// Проверяем что ссылки есть
	if len(req.Links) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Список ссылок не может быть пустым",
		})
		return
	}

	// Сохраняем задачу в базу
	taskID, err := storage.SaveTask(req.Links)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Не удалось создать задачу",
		})
		return
	}

	// Ставим задачу в очередь на проверку
	checker.AddTask(taskID, req.Links)

	// Формируем начальный ответ
	initialStatus := make(map[string]string)
	for _, link := range req.Links {
		initialStatus[link] = "проверяется..."
	}

	c.JSON(http.StatusOK, CheckResponse{
		Links:    initialStatus,
		LinksNum: taskID,
	})
}

// Обработчик генерации отчета
func handleReport(c *gin.Context, storage *Storage) {
	var req ReportRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Некорректный запрос. Укажите номера задач",
		})
		return
	}

	// Получаем данные по задачам
	tasks, err := storage.GetTasks(req.LinksList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка при получении данных",
		})
		return
	}

	if len(tasks) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Задачи с указанными номерами не найдены",
		})
		return
	}

	// Генерируем PDF отчет
	pdfBytes, err := GenerateReportPDF(tasks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Не удалось сгенерировать отчет",
		})
		return
	}

	// Отправляем файл
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
	c.Header("Content-Disposition", "attachment; filename=links_report.pdf")
}
