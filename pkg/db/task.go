package db

import (
	"fmt"
	"strconv"
	"time"
)

// Константа для лимита задач по умолчанию
const defaultTasksLimit = 50

// Структура задачи
type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"` // Дата в формате YYYYMMDD
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"` // Правило повторения
}

// Добавляет задачу в базу данных и возвращает её ID
func AddTask(task *Task) (int64, error) {
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Возвращает список задач из базы данных, отсортированных по дате
func Tasks(limit int, search string) ([]*Task, error) {
	// Если лимит не указан или меньше нуля, ставим значение по умолчанию
	if limit <= 0 {
		limit = defaultTasksLimit
	}

	var query string
	var args []interface{}

	// Если поиск не указан, возвращаем все задачи
	if len(search) == 0 {
		query = `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT ?`
		args = []interface{}{limit}
	} else {
		// Пытаемся распарсить строку как дату в формате 02.01.2006
		searchDate, err := time.Parse("02.01.2006", search)
		if err == nil {
			// Если это дата, ищем задачи с этой датой
			dateStr := searchDate.Format("20060102")
			query = `SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date ASC LIMIT ?`
			args = []interface{}{dateStr, limit}
		} else {
			// Иначе ищем подстроку в названии или комментарии
			searchPattern := "%" + search + "%"
			query = `SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date ASC LIMIT ?`
			args = []interface{}{searchPattern, searchPattern, limit}
		}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		var id int64
		err := rows.Scan(&id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, err
		}
		task.ID = strconv.FormatInt(id, 10)
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Если задач нет, возвращаем пустой список
	if tasks == nil {
		tasks = []*Task{}
	}

	return tasks, nil
}

// Возвращает задачу по её ID
func GetTask(id string) (*Task, error) {
	taskID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, err
	}

	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	var task Task
	var dbID int64
	err = db.QueryRow(query, taskID).Scan(&dbID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return nil, err
	}
	task.ID = strconv.FormatInt(dbID, 10)
	return &task, nil
}

// Обновляет задачу в базе данных
func UpdateTask(task *Task) error {
	taskID, err := strconv.ParseInt(task.ID, 10, 64)
	if err != nil {
		return err
	}

	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	result, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, taskID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("задача не найдена")
	}

	return nil
}

// Удаляет задачу из базы данных по ID
func DeleteTask(id string) error {
	taskID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	query := `DELETE FROM scheduler WHERE id = ?`
	result, err := db.Exec(query, taskID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("задача не найдена")
	}

	return nil
}
