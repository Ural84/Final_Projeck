package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"scheduler/pkg/api"
	"scheduler/tests"
)

// Start запускает веб-сервер на указанном порту
// Порт определяется из переменной окружения TODO_PORT или из tests/settings.go
func Start() error {
	// Регистрируем API обработчики
	api.Init()

	// Определяем порт: сначала проверяем переменную окружения TODO_PORT,
	// затем используем значение из tests/settings.go
	port := tests.Port
	envPort := os.Getenv("TODO_PORT")
	if len(envPort) > 0 {
		if eport, err := strconv.ParseInt(envPort, 10, 32); err == nil {
			port = int(eport)
		}
	}

	// Настраиваем файловый сервер для статических файлов из директории web
	// http.FileServer автоматически обслуживает все файлы из указанной директории:
	// - http://localhost:7540/ → web/index.html
	// - http://localhost:7540/js/scripts.min.js → web/js/scripts.min.js
	// - http://localhost:7540/css/style.css → web/css/style.css
	// - http://localhost:7540/favicon.ico → web/favicon.ico
	webDir := "web"
	fs := http.FileServer(http.Dir(webDir))
	http.Handle("/", fs)

	// Запускаем сервер
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Сервер запущен на порту %d", port)
	log.Printf("Откройте http://localhost:%d в браузере", port)

	if err := http.ListenAndServe(addr, nil); err != nil {
		return fmt.Errorf("ошибка запуска сервера: %w", err)
	}

	return nil
}
