package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"scheduler/pkg/db"
)

// addTaskHandler обрабатывает POST-запросы для добавления задачи
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Декодируем JSON из тела запроса
	var task db.Task
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&task); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()})
		return
	}

	// Проверяем обязательное поле title
	if len(task.Title) == 0 {
		writeJSON(w, map[string]string{"error": "Не указан заголовок задачи"})
		return
	}

	// Проверяем и корректируем дату
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

	// Возвращаем успешный ответ с ID
	writeJSON(w, map[string]string{"id": formatInt64(id)})
}

// checkDate проверяет и корректирует дату задачи
func checkDate(task *db.Task) error {
	now := time.Now()
	today := now.Format(DateFormat)

	// Если task.Date пустая строка, присваиваем текущее время
	if len(task.Date) == 0 || task.Date == "" {
		task.Date = today
	}

	// Проверяем, что в task.Date указана корректная дата
	_, err := time.Parse(DateFormat, task.Date)
	if err != nil {
		return fmt.Errorf("дата представлена в формате, отличном от 20060102")
	}

	// СРАВНИВАЕМ СЕГОДНЯШНЮЮ ДАТУ ПОСЛЕ ПРОВЕРКИ ФОРМАТА
	// Это критично для случая, когда дата равна today - она должна остаться today
	// независимо от правила повторения
	if task.Date == today {
		// Проверяем корректность правила повторения, если оно указано
		if len(task.Repeat) > 0 && task.Repeat != "" {
			_, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return fmt.Errorf("правило повторения указано в неправильном формате: %w", err)
			}
		}
		// ВАЖНО: Явно устанавливаем сегодняшнюю дату и возвращаемся
		// Это гарантирует, что дата останется today даже если NextDate вернул другую дату
		task.Date = today
		return nil
	}

	// Если дата в прошлом (строго меньше today), корректируем дату
	if task.Date < today {
		if len(task.Repeat) == 0 || task.Repeat == "" {
			// Если правила повторения нет, берём сегодняшнее число
			task.Date = today
		} else {
			// Проверяем правило на корректность и получаем следующую дату
			next, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return fmt.Errorf("правило повторения указано в неправильном формате: %w", err)
			}
			// Используем вычисленную следующую дату
			task.Date = next
		}
	}
	// Если task.Date > today, оставляем как есть (не меняем)

	return nil
}

// formatInt64 преобразует int64 в строку
func formatInt64(n int64) string {
	return strconv.FormatInt(n, 10)
}
