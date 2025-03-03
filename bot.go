package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StartBot() {
	bot, err := tgbotapi.NewBotAPI(botConfig.Token)
	if err != nil {
		log.Fatalf("Ошибка при создании бота: %v", err)
	}

	log.Println("Бот успешно запущен")

	go messageSender(bot)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			userName := update.Message.From.UserName
			if userName == "" {
				userName = "Unknown"
			}
			log.Printf("Получено сообщение от пользователя %d (@%s): %s", update.Message.Chat.ID, userName, update.Message.Text)
			handleMessage(bot, update.Message)
		}
	}
}

func messageSender(bot *tgbotapi.BotAPI) {

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		for userID, userConfig := range botConfig.Users {
			if !userConfig.Active {
				continue // Пропускаем неактивных пользователей
			}
			// Загружаем часовой пояс пользователя
			location, err := time.LoadLocation(userConfig.TimeZone)
			if err != nil {
				log.Printf("Ошибка при загрузке локации для пользователя %d: %v, используется UTC", userID, err)
				location = time.UTC
			}

			now := time.Now().In(location)
			currentHour := now.Hour()
			currentMinute := now.Minute()
			// Проверяем, находимся ли в разрешенном временном интервале
			if currentHour >= userConfig.StartTime && currentHour < userConfig.EndTime {
				// Проверяем, соответствует ли текущее время интервалу отправки
				if currentMinute%userConfig.Interval == 0 {
					sendMessageToUser(bot, userID, userConfig)
				}
			}
		}
	}
}

func sendMessageToUser(bot *tgbotapi.BotAPI, userID int64, userConfig *UserConfig) {
	randomPhrase := userConfig.GetNextUniquePhrase()

	msg := tgbotapi.NewMessage(userID, randomPhrase)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка при отправке сообщения пользователю %d (@%s): %v)", userID, userConfig.UserName, err)
	} else {
		log.Printf("Сообщение отправлено пользователю %d (@%s): %s", userID, userConfig.UserName, randomPhrase)
	}

	saveUsers()
}

func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	userID := message.Chat.ID
	userName := message.From.UserName
	if userName == "" {
		userName = "Unknown"
	}

	// Разбиваем сообщение на команду и аргументы
	parts := strings.Split(message.Text, " ")
	command := parts[0]

	switch command {
	case "/start":
		// Создаем пользователя с настройками по умолчанию, если его еще нет
		if _, exists := botConfig.Users[userID]; !exists {
			botConfig.Users[userID] = createDefaultUser(userID, userName)
			saveUsers()
		}

		botConfig.Users[userID].Active = true // Активируем пользователя, если он уже был создан ранее
		saveUsers()

		msg := tgbotapi.NewMessage(userID,
			"Привет! Доверься мне, и я сделаю твой день теплее и ярче 💕\n\n"+
				"Начнём? Я буду рядом, просто скажи:\n"+
				"/start 🚀 - запуск бота\n"+
				"/stop 🛑 - остановка получения сообщений\n"+
				"/test 📩 - получить тестовое сообщение\n"+
				"/settings ⚙️ - показать текущие настройки\n"+
				"/time start end ⏰- установить время начала и окончания (часы, 0-23)\n"+
				"/interval min ⏳ - установить интервал в минутах\n"+
				"/help - ❓ показать эту справку\n\n"+
				"Если что-то не работает или есть вопросы, пиши @undermord — он поможет! 💖")

		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка при отправке приветственного сообщения пользователю %d(@%s): %v", userID, userName, err)
		} else {
			log.Printf("Приветственное сообщение отправлено пользователю %d(@%s)", userID, userName)
		}

	case "/stop":
		if _, exists := botConfig.Users[userID]; exists {
			botConfig.Users[userID].Active = false
			saveUsers()
			msg := tgbotapi.NewMessage(userID, "Вы отписались от получения сообщений. Чтобы возобновить, используйте /start")
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Ошибка при отправке сообщения пользователю %d(@%s): %v", userID, userName, err)
			}
		} else {
			msg := tgbotapi.NewMessage(userID, "Вы не были подписаны на сообщения. Используйте /start, чтобы подписаться.")
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Ошибка при отправке сообщения пользователю %d(@%s): %v", userID, userName, err)
			}
		}

	case "/test":
		if user, exists := botConfig.Users[userID]; exists {
			randomPhrase := user.GetNextUniquePhrase()

			msg := tgbotapi.NewMessage(userID, randomPhrase)
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Ошибка при отправке тестового сообщения пользователю %d (@%s): %v", userID, userName, err)
			} else {
				log.Printf("Тестовое сообщение отправлено пользователю %d (@%s): %s", userID, userName, randomPhrase)
			}
			saveUsers()
		} else {
			msg := tgbotapi.NewMessage(userID, "Вы ещё не зарегистрированы. Используйте /start для регистрации.")
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Ошибка при отправке сообщения пользователю %d(@%s): %v", userID, userName, err)
			}
		}

	case "/settings":
		if user, exists := botConfig.Users[userID]; exists {
			msg := tgbotapi.NewMessage(userID, fmt.Sprintf(
				"Ваши текущие настройки:\n"+
					"Активность: %v\n"+
					"Время начала отправки: %d:00\n"+
					"Время окончания отправки: %d:00\n"+
					"Интервал отправки: %d минут\n"+
					"Часовой пояс: %s",
				user.Active, user.StartTime, user.EndTime, user.Interval, user.TimeZone))
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Ошибка при отправке настроек пользователю %d(@%s): %v", userID, userName, err)
			}
		} else {
			msg := tgbotapi.NewMessage(userID, "Вы ещё не зарегистрированы. Используйте /start для регистрации.")
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Ошибка при отправке сообщения пользователю %d(@%s): %v", userID, userName, err)
			}
		}

	case "/time":
		if len(parts) < 3 {
			msg := tgbotapi.NewMessage(userID, "Неверный формат. Используйте: /time start end\nНапример: /time 8 22")
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Ошибка при отправке сообщения пользователю %d(@%s): %v", userID, userName, err)
			}
			return
		}

		startTime, err1 := strconv.Atoi(parts[1])
		endTime, err2 := strconv.Atoi(parts[2])

		if err1 != nil || err2 != nil || startTime < 0 || startTime > 23 || endTime < 0 || endTime > 23 || startTime >= endTime {
			msg := tgbotapi.NewMessage(userID, "Неверное время. Используйте часы от 0 до 23, где время начала меньше времени окончания.")
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Ошибка при отправке сообщения пользователю %d(@%s): %v", userID, userName, err)
			}
			return
		}

		if user, exists := botConfig.Users[userID]; exists {
			user.StartTime = startTime
			user.EndTime = endTime
			saveUsers()
			msg := tgbotapi.NewMessage(userID, fmt.Sprintf("Время отправки сообщений установлено с %d:00 до %d:00", startTime, endTime))
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Ошибка при отправке сообщения пользователю %d(@%s): %v", userID, userName, err)
			}
		} else {
			msg := tgbotapi.NewMessage(userID, "Вы ещё не зарегистрированы. Используйте /start для регистрации.")
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Ошибка при отправке сообщения пользователю %d(@%s): %v", userID, userName, err)
			}
		}

	case "/interval":
		if len(parts) < 2 {
			msg := tgbotapi.NewMessage(userID, "Неверный формат. Используйте: /interval minutes\nНапример: /interval 60")
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Ошибка при отправке сообщения пользователю %d(@%s): %v", userID, userName, err)
			}
			return
		}

		interval, err := strconv.Atoi(parts[1])
		if err != nil || interval < 1 || interval > 1440 {
			msg := tgbotapi.NewMessage(userID, "Неверный интервал. Используйте значение от 1 до 1440 минут (24 часа).")
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Ошибка при отправке сообщения пользователю %d(@%s): %v", userID, userName, err)
			}
			return
		}

		if user, exists := botConfig.Users[userID]; exists {
			user.Interval = interval
			saveUsers()
			msg := tgbotapi.NewMessage(userID, fmt.Sprintf("Интервал отправки сообщений установлен на %d минут", interval))
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Ошибка при отправке сообщения пользователю %d(@%s): %v", userID, userName, err)
			}
		} else {
			msg := tgbotapi.NewMessage(userID, "Вы ещё не зарегистрированы. Используйте /start для регистрации.")
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Ошибка при отправке сообщения пользователю %d(@%s): %v", userID, userName, err)
			}
		}

	case "/help":
		msg := tgbotapi.NewMessage(userID,
			"/start 🚀 - запуск бота\n"+
				"/stop 🛑 - остановка получения сообщений\n"+
				"/test 📩 - получить тестовое сообщение\n"+
				"/settings ⚙️ - показать текущие настройки\n"+
				"/time start end ⏰- установить время начала и окончания (часы, 0-23)\n"+
				"/interval min ⏳ - установить интервал в минутах\n"+
				"/help - ❓ показать эту справку\n\n"+
				"Если что-то не работает или есть вопросы, пиши @undermord — он поможет! 💖")

		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка при отправке справки пользователю %d(@%s): %v", userID, userName, err)
		}

	default:
		msg := tgbotapi.NewMessage(userID, "Я не понимаю эту команду. Напишите /help для получения списка доступных команд.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка при отправке сообщения пользователю %d (@%s): %v", userID, userName, err)
		} else {
			log.Printf("Сообщение отправлено пользователю %d (@%s): %s", userID, userName, msg.Text)
		}
	}
}
