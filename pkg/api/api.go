package api

import (
	"net/http"
	"time"
)

// Регистрирует все обработчики API
func Init() {
	http.HandleFunc("/api/nextdate", nextDateHandler)
	http.HandleFunc("/api/task", taskHandler)
	http.HandleFunc("/api/tasks", tasksHandler)
	http.HandleFunc("/api/task/done", doneTaskHandler)
}

// Обрабатывает GET-запрос для вычисления следующей даты
func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры из запроса
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeatStr := r.FormValue("repeat")

	// Если параметр now не указан, используем текущую дату
	var now time.Time
	if len(nowStr) == 0 {
		now = time.Now()
	} else {
		var err error
		now, err = time.Parse(DateFormat, nowStr)
		if err != nil {
			http.Error(w, "некорректный формат параметра now", http.StatusBadRequest)
			return
		}
	}

	// Проверяем обязательные параметры
	if len(dateStr) == 0 {
		http.Error(w, "параметр date обязателен", http.StatusBadRequest)
		return
	}
	if len(repeatStr) == 0 {
		http.Error(w, "параметр repeat обязателен", http.StatusBadRequest)
		return
	}

	// Вычисляем следующую дату
	nextDate, err := NextDate(now, dateStr, repeatStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Возвращаем следующую дату
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(nextDate))
}
