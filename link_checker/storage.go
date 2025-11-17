package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Структура для хранения
type Task struct {
	ID        int               `json:"id"`
	Links     map[string]string `json:"links"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type Storage struct {
	mu       sync.RWMutex
	tasks    map[int]*Task
	nextID   int
	dataFile string
}

func NewStorage() *Storage {
	storage := &Storage{
		tasks:    make(map[int]*Task),
		nextID:   1,
		dataFile: "tasks_data.json",
	}

	// Загружаем сохраненные данные при старте
	if err := storage.load(); err != nil {
		fmt.Printf("Не удалось загрузить данные: %v\n", err)
	}

	return storage
}

// Сохранение новой задачи
func (s *Storage) SaveTask(links []string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	taskID := s.nextID
	s.nextID++

	// Создаем задачу
	task := &Task{
		ID:        taskID,
		Links:     make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Инициализируем статусы
	for _, link := range links {
		task.Links[link] = "в очереди"
	}

	s.tasks[taskID] = task

	// Сохраняем в файл
	if err := s.save(); err != nil {
		return 0, err
	}

	return taskID, nil
}

// Обновление статуса ссылки
func (s *Storage) UpdateLinkStatus(taskID int, link string, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("задача %d не найдена", taskID)
	}

	task.Links[link] = status
	task.UpdatedAt = time.Now()

	return s.save()
}

// Получение задач
func (s *Storage) GetTasks(ids []int) ([]*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Task
	for _, id := range ids {
		if task, exists := s.tasks[id]; exists {
			result = append(result, task)
		}
	}

	return result, nil
}

// Сохранение в файл
func (s *Storage) save() error {
	data := struct {
		Tasks  map[int]*Task `json:"tasks"`
		NextID int           `json:"next_id"`
	}{
		Tasks:  s.tasks,
		NextID: s.nextID,
	}

	file, err := os.Create(s.dataFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// Загрузка из файла
func (s *Storage) load() error {
	file, err := os.Open(s.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Файла нет - это нормально
		}
		return err
	}
	defer file.Close()

	var data struct {
		Tasks  map[int]*Task `json:"tasks"`
		NextID int           `json:"next_id"`
	}

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return err
	}

	s.tasks = data.Tasks
	s.nextID = data.NextID
	return nil
}

// сохраняем данные при закрытии
func (s *Storage) Close() {
	s.save()
}
