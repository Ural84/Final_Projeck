package api

import (
	"net/http"

	"scheduler/pkg/db"
)

// Структура для ответа со списком задач
type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

// Обрабатывает GET-запрос для получения списка задач
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, map[string]string{"error": "метод не поддерживается"})
		return
	}

	// Получаем параметр поиска из запроса
	search := r.URL.Query().Get("search")

	// Получаем задачи из базы данных (максимум 50)
	tasks, err := db.Tasks(50, search)
	if err != nil {
		writeJSON(w, map[string]string{"error": "ошибка при получении задач из базы данных: " + err.Error()})
		return
	}

	// Если задач нет, возвращаем пустой список
	if tasks == nil {
		tasks = []*db.Task{}
	}

	// Возвращаем список задач
	writeJSON(w, TasksResp{
		Tasks: tasks,
	})
}
