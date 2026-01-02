package api

import (
	"encoding/json"
	"net/http"

	"scheduler/pkg/db"
)

// Обрабатывает PUT-запрос для обновления задачи
func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Читаем JSON из запроса
	var task db.Task
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&task); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Проверяем, что указан ID задачи
	if len(task.ID) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Не указан идентификатор задачи"})
		return
	}

	// Проверяем, что указан заголовок задачи
	if len(task.Title) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Не указан заголовок задачи"})
		return
	}

	// Проверяем и исправляем дату
	if err := checkDate(&task); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Обновляем задачу в базе данных
	if err := db.UpdateTask(&task); err != nil {
		if err.Error() == "задача не найдена" {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		} else {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}

	// Возвращаем пустой ответ при успехе
	writeJSON(w, http.StatusOK, map[string]interface{}{})
}

