package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

// Глобальная переменная для хранения подключения к базе данных
var db *sql.DB

// schema содержит SQL команды для создания таблицы scheduler и индекса
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

// Init открывает базу данных и при необходимости создаёт таблицу с индексом
func Init(dbFile string) error {
	// Проверяем существование файла базы данных
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

	// Сохраняем подключение в глобальную переменную
	db = database

	return nil
}

// GetDB возвращает глобальное подключение к базе данных
func GetDB() *sql.DB {
	return db
}

