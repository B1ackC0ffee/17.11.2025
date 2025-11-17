package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type LinkChecker struct {
	storage   *Storage
	taskQueue chan taskItem
	stopChan  chan bool
	wg        sync.WaitGroup
	workers   int
}

type taskItem struct {
	TaskID int
	Links  []string
}

func NewLinkChecker(storage *Storage) *LinkChecker {
	return &LinkChecker{
		storage:   storage,
		taskQueue: make(chan taskItem, 100), // Буфер на 100 задач
		stopChan:  make(chan bool),
		workers:   3, // 3 рабочих для проверки
	}
}

// Запуск воркеров
func (lc *LinkChecker) Start() {
	for i := 0; i < lc.workers; i++ {
		lc.wg.Add(1)
		go lc.worker(i + 1)
	}
	fmt.Printf("Запущено %d воркеров для проверки ссылок\n", lc.workers)
}

// Остановка
func (lc *LinkChecker) Stop() {
	close(lc.stopChan)
	lc.wg.Wait()
	fmt.Println("Все воркеры остановлены")
}

// Добавление задачи в очередь
func (lc *LinkChecker) AddTask(taskID int, links []string) {
	item := taskItem{
		TaskID: taskID,
		Links:  links,
	}

	select {
	case lc.taskQueue <- item:
		fmt.Printf("Задача #%d добавлена в очередь\n", taskID)
	default:
		fmt.Printf("Очередь переполнена, задача #%д ждет\n", taskID)
		lc.taskQueue <- item // Блокируем пока не освободится место
	}
}

// Воркер для проверки ссылок
func (lc *LinkChecker) worker(id int) {
	defer lc.wg.Done()

	fmt.Printf("Воркер #%d запущен\n", id)

	for {
		select {
		case <-lc.stopChan:
			fmt.Printf("Воркер #%d завершает работу\n", id)
			return
		case task := <-lc.taskQueue:
			lc.processTask(task, id)
		}
	}
}

// Обработка одной задачи
func (lc *LinkChecker) processTask(task taskItem, workerID int) {
	fmt.Printf("Воркер #%d обрабатывает задачу #%d (%d ссылок)\n",
		workerID, task.TaskID, len(task.Links))

	for _, link := range task.Links {
		// Обновляем статус на "проверяется"
		lc.storage.UpdateLinkStatus(task.TaskID, link, "checking...")

		// Проверяем доступность
		status := lc.checkLink(link)

		// Сохраняем результат
		lc.storage.UpdateLinkStatus(task.TaskID, link, status)

		// Небольшая пауза чтобы не перегружать
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("Задача #%d завершена воркером #%d\n", task.TaskID, workerID)
}

// Проверка одной ссылки
func (lc *LinkChecker) checkLink(link string) string {
	// Добавляем протокол если нужно
	fullURL := link
	if len(link) > 7 && link[:7] != "http://" && link[:8] != "https://" {
		fullURL = "https://" + link
	}

	// Создаем клиент с таймаутом
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	// Пробуем сделать запрос
	resp, err := client.Head(fullURL)
	if err != nil {
		return "not available"
	}
	defer resp.Body.Close()

	// Проверяем статус код
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return "available"
	}

	return "not available"
}

// Восстановление задач после перезапуска
func (lc *LinkChecker) RecoverTasks() {
	time.Sleep(2 * time.Second) // Ждем запуска

	fmt.Println("Проверяем незавершенные задачи...")
}
