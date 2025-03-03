package main

import (
	"log"
	"os"
)

func main() {
	log.Println("Инициализация конфигурации бота")

	// Получаем токен из переменных окружения или используем значение по умолчанию
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		token = ""
		log.Println("Используется токен по умолчанию. Рекомендуется установить переменную окружения BOT_TOKEN")
	}

	InitConfig(token, make(map[int64]*UserConfig))

	log.Println("Запуск бота...")
	StartBot()

	select {}
}
