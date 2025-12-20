package api

import (
	"encoding/json"
	"net/http"
)

// taskHandler обрабатывает запросы к /api/task в зависимости от HTTP метода
func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
	// обработка других методов будет добавлена на следующих шагах
	default:
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

// writeJSON отправляет данные в формате JSON
func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

