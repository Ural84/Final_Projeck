package api

import (
	"net/http"

	"scheduler/pkg/db"
)

// Обрабатывает GET-запрос для получения задачи по ID
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем ID из параметров запроса
	id := r.URL.Query().Get("id")
	if len(id) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Не указан идентификатор"})
		return
	}

	// Получаем задачу из базы данных
	task, err := db.GetTask(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Задача не найдена"})
		return
	}

	// Возвращаем задачу
	writeJSON(w, http.StatusOK, task)
}


