package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

// Переменная для хранения подключения к базе данных
var db *sql.DB

// SQL команды для создания таблицы и индекса
const schema = `
CREATE TABLE scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT "",
    title VARCHAR(255) NOT NULL DEFAULT "",
    comment TEXT NOT NULL DEFAULT "",
    repeat VARCHAR(128) NOT NULL DEFAULT ""
);

CREATE INDEX idx_scheduler_date ON scheduler(date);
`

// Открывает базу данных и создаёт таблицу, если её нет
func Init(dbFile string) error {
	// Проверяем, существует ли файл базы данных
	_, err := os.Stat(dbFile)
	var install bool
	if err != nil {
		install = true
	}

	// Открываем базу данных
	database, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("ошибка открытия БД: %w", err)
	}

	// Проверяем подключение
	if err := database.Ping(); err != nil {
		database.Close()
		return fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	// Если файла не было, создаём таблицу и индекс
	if install {
		if _, err := database.Exec(schema); err != nil {
			database.Close()
			return fmt.Errorf("ошибка создания таблицы и индекса: %w", err)
		}
	}

	// Сохраняем подключение
	db = database

	return nil
}

// Возвращает подключение к базе данных
func GetDB() *sql.DB {
	return db
}
