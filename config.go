package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"time"
)

type UserConfig struct {
	UserId             int64
	UserName           string
	Message            string
	StartTime          int
	EndTime            int
	Interval           int
	Active             bool
	TimeZone           string
	location           *time.Location
	SentPhrasesIndices []int
}

type BotConfig struct {
	Token   string
	Users   map[int64]*UserConfig
	Phrases []string
}

var botConfig *BotConfig

func InitConfig(token string, users map[int64]*UserConfig) {
	botConfig = &BotConfig{
		Token: token,
		Users: users,
	}

	log.Println("Конфигурация бота инициализирована")
	loadPhrases("phrases.json")
	loadUsers("users.json")
}

// Загружать часовой пояс при создании или обновлении пользователя
func (u *UserConfig) LoadLocation() error {
	loc, err := time.LoadLocation(u.TimeZone)
	if err != nil {
		return err
	}
	u.location = loc
	return nil
}

// Использовать кэшированный часовой пояс
func (u *UserConfig) Now() time.Time {
	if u.location == nil {
		if err := u.LoadLocation(); err != nil {
			u.location = time.UTC
		}
	}
	return time.Now().In(u.location)
}

func loadPhrases(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("Ошибка при чтении файла %s: %v", filename, err)
		botConfig.Phrases = []string{
			"💖 Я люблю тебя! 💖",
			"🌟 Ты особенный человек! 🌟",
			"🌹 Ты украшаешь мой день! 🌹",
		}
		return
	}

	err = json.Unmarshal(data, &botConfig.Phrases)
	if err != nil {
		log.Printf("Ошибка при разборе JSON из файла %s: %v", filename, err)
	}
	log.Printf("Загружено %d фраз из файла %s", len(botConfig.Phrases), filename)
}

func loadUsers(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("Файл пользователей не найден %s: %v", filename)
		return
	}

	var users map[int64]*UserConfig
	err = json.Unmarshal(data, &users)
	if err != nil {
		log.Printf("Ошибка при разборе JSON из файла пользователей %s: %v", filename, err)
	}

	botConfig.Users = users

	// Загружаем часовые пояса для всех пользователей
	for _, user := range users {
		if err := user.LoadLocation(); err != nil {
			log.Printf("Ошибка при загрузке часового пояса для пользователя %d: %v, используется UTC", user.UserId, err)
			user.location = time.UTC
		}
	}

	log.Printf("Загружено %d пользователей из файл %s", len(users), filename)
}

func saveUsers() {
	data, err := json.MarshalIndent(botConfig.Users, "", "  ")
	if err != nil {
		log.Printf("Ошибка при сериализации пользователей: %v", err)
	}

	err = os.WriteFile("users.json", data, 0644)
	if err != nil {
		log.Printf("Ошибка сохранения пользователей : %v", err)
		return
	}

	log.Printf("пользователи успешны сохранены в файл users.json")
}

func createDefaultUser(userID int64, userName string) *UserConfig {
	user := &UserConfig{
		UserId:             userID,
		UserName:           userName,
		Message:            "💖 Я люблю тебя! 💖",
		StartTime:          8,
		EndTime:            22,
		Interval:           120,
		Active:             true,
		TimeZone:           "Europe/Samara",
		SentPhrasesIndices: []int{},
	}

	if err := user.LoadLocation(); err != nil {
		log.Printf("Ошибка при загрузке часового пояса для нового пользователя: %v, используется UTC", err)
		user.TimeZone = "UTC"
		user.location = time.UTC
	}
	return user
}

// Получает следующую уникальную случайную фразу для этого пользователя
func (u *UserConfig) GetNextUniquePhrase() string {
	if len(u.SentPhrasesIndices) >= len(botConfig.Phrases) {
		u.SentPhrasesIndices = []int{}
		log.Printf("Сброс цикла сообщений для пользователя %d(@%s)", u.UserId, u.UserName)
	}
	// Генерируем случайный индекс, который ещё не использовался
	var index int
	for {
		index = rand.Intn(len(botConfig.Phrases))
		// Проверяем, использовался ли этот индекс ранее
		used := false
		for _, sentIndex := range u.SentPhrasesIndices {
			if index == sentIndex {
				used = true
				break
			}
		}
		// Если не использовался, можем использовать эту фразу
		if !used {
			break
		}
	}
	u.SentPhrasesIndices = append(u.SentPhrasesIndices, index)

	return botConfig.Phrases[index]
}
