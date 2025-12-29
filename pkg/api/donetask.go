package api

import (
	"net/http"
	"time"

	"scheduler/pkg/db"
)

// Обрабатывает POST-запрос для отметки задачи как выполненной
func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, map[string]string{"error": "метод не поддерживается"})
		return
	}

	// Получаем ID из параметров запроса
	id := r.URL.Query().Get("id")
	if len(id) == 0 {
		writeJSON(w, map[string]string{"error": "Не указан идентификатор"})
		return
	}

	// Получаем задачу из базы данных
	task, err := db.GetTask(id)
	if err != nil {
		writeJSON(w, map[string]string{"error": "Задача не найдена"})
		return
	}

	// Если правила повторения нет, удаляем задачу
	if len(task.Repeat) == 0 || task.Repeat == "" {
		if err := db.DeleteTask(id); err != nil {
			writeJSON(w, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, map[string]interface{}{})
		return
	}

	// Если есть правило повторения, вычисляем следующую дату
	now := time.Now()
	nextDate, err := NextDate(now, task.Date, task.Repeat)
	if err != nil {
		writeJSON(w, map[string]string{"error": "Ошибка при вычислении следующей даты: " + err.Error()})
		return
	}

	// Обновляем дату задачи на следующую
	task.Date = nextDate
	if err := db.UpdateTask(task); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, map[string]interface{}{})
}


