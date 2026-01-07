package api

import (
	"net/http"

	"scheduler/pkg/db"
)

// Обрабатывает DELETE-запрос для удаления задачи
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем ID из параметров запроса
	id := r.URL.Query().Get("id")
	if len(id) == 0 {
		writeJSON(w, map[string]string{"error": "Не указан идентификатор"})
		return
	}

	// Удаляем задачу из базы данных
	if err := db.DeleteTask(id); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()})
		return
	}

	// Возвращаем пустой ответ при успехе
	writeJSON(w, map[string]interface{}{})
}


