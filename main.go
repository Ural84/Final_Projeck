package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"scheduler/pkg/db"
	"scheduler/pkg/server"
)

func main() {
	// Определяем путь к файлу базы данных
	// Проверяем переменную окружения TODO_DBFILE, иначе используем значение по умолчанию
	dbFile := os.Getenv("TODO_DBFILE")
	if len(dbFile) == 0 {
		dbFile = "scheduler.db"
	}

	// Инициализируем базу данных
	if err := db.Init(dbFile); err != nil {
		log.Fatalf("Ошибка инициализации БД: %v", err)
	}

	// Убеждаемся, что подключение к БД будет закрыто при завершении
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Ошибка закрытия БД: %v", err)
		} else {
			log.Println("Подключение к БД закрыто")
		}
	}()

	// Запускаем сервер в горутине
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	// Ожидаем сигнал завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Получен сигнал завершения, закрываем подключение к БД...")
}
