package main

import (
	"log"
	"os"
)

func initLogging() {
	// Открываем файл для записи логов
	logFile, err := os.OpenFile("lovebot.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Ошибка при открытии файла логов: %v", err)
	}

	log.SetOutput(logFile)

	log.SetFlags(log.Ldate | log.Ldate | log.Lshortfile)

	log.Println("Логирование инициализировано")
}

func main() {
	initLogging()
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
