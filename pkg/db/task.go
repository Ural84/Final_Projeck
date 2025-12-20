package db

// Task представляет задачу в базе данных
type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`    // Формат: YYYYMMDD (20060102)
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`  // Правило повторения
}

// AddTask добавляет задачу в таблицу scheduler и возвращает идентификатор добавленной записи
func AddTask(task *Task) (int64, error) {
	var id int64
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err == nil {
		id, err = res.LastInsertId()
	}
	return id, err
}

