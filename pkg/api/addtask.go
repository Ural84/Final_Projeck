package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"scheduler/pkg/db"
)

// Обрабатывает POST-запрос для добавления задачи
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Читаем JSON из запроса
	var task db.Task
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&task); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()})
		return
	}

	// Проверяем, что указан заголовок задачи
	if len(task.Title) == 0 {
		writeJSON(w, map[string]string{"error": "Не указан заголовок задачи"})
		return
	}

	// Проверяем и исправляем дату
	if err := checkDate(&task); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()})
		return
	}

	// Добавляем задачу в базу данных
	id, err := db.AddTask(&task)
	if err != nil {
		writeJSON(w, map[string]string{"error": "ошибка при добавлении задачи в базу данных: " + err.Error()})
		return
	}

	// Возвращаем ID добавленной задачи
	writeJSON(w, map[string]string{"id": formatInt64(id)})
}

// Проверяет и исправляет дату задачи
func checkDate(task *db.Task) error {
	now := time.Now()
	today := now.Format(DateFormat)

	// Если дата не указана, ставим сегодняшнюю дату
	if len(task.Date) == 0 || task.Date == "" {
		task.Date = today
	}

	// Проверяем, что дата в правильном формате
	_, err := time.Parse(DateFormat, task.Date)
	if err != nil {
		return fmt.Errorf("дата представлена в формате, отличном от 20060102")
	}

	// Если дата равна сегодняшней, оставляем её как есть
	if task.Date == today {
		// Проверяем правило повторения, если оно указано
		if len(task.Repeat) > 0 && task.Repeat != "" {
			_, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return fmt.Errorf("правило повторения указано в неправильном формате: %w", err)
			}
		}
		// Явно ставим сегодняшнюю дату
		task.Date = today
		return nil
	}

	// Если дата в прошлом, исправляем её
	if task.Date < today {
		if len(task.Repeat) == 0 || task.Repeat == "" {
			// Если правила повторения нет, ставим сегодняшнюю дату
			task.Date = today
		} else {
			// Вычисляем следующую дату по правилу повторения
			next, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return fmt.Errorf("правило повторения указано в неправильном формате: %w", err)
			}
			task.Date = next
		}
	}

	return nil
}

// Преобразует число в строку
func formatInt64(n int64) string {
	return strconv.FormatInt(n, 10)
}
