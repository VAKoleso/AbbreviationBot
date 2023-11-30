package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"database/sql"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/lib/pq"
)

type UserState struct {
	WaitingForAbbreviation bool
}

type BotHandler struct {
	bot        *tgbotapi.BotAPI
	db         *sql.DB
	userStates map[int]*UserState
	AddArgs    string
	AddFlag    bool
}

// Создание структуры типов данных для методов

func NewBotHandler(bot *tgbotapi.BotAPI, db *sql.DB) *BotHandler {
	return &BotHandler{
		bot:        bot,
		db:         db,
		userStates: make(map[int]*UserState),
		AddFlag:    false,
	}
}

// Запуск бота

func (bh *BotHandler) Start() {
	updates, _ := bh.bot.GetUpdatesChan(tgbotapi.NewUpdate(0))

	for update := range updates {
		if update.Message == nil {
			continue
		}
		bh.HandleMessage(update)
	}
}

func (bh *BotHandler) HandleMessage(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID

	if update.Message.Text == "" {
		return
	}

	if update.Message.IsCommand() {
		msg := tgbotapi.NewMessage(chatID, "")
		bh.HandleCommand(update, msg)
	} else {
		msg := tgbotapi.NewMessage(chatID, "")
		bh.HandleNonCommandMessage(update, msg)
	}
}

func (bh *BotHandler) HandleCommand(update tgbotapi.Update, msg tgbotapi.MessageConfig) {
	userID := update.Message.From.ID
	switch update.Message.Command() {
	case "add":
		bh.handleAddCommand(update, msg, userID)
	case "start":
		msg.Text = "Приветствую! Я твой личный ассистент по аббревиатурам. Готов помочь тебе разобраться в " +
			"сокращениях и сокращенных выражений. Просто отправь мне аббревиатуру, а я предоставлю ее расшифровку. " +
			"Хочешь добавить новую аббревиатуру? Просто используй команду /add [аббревиатура] [значение]. Давай начнем!"
		SendBot(msg, bh.bot)
	default:
		msg.Text = "Неизвестная команда."
		SendBot(msg, bh.bot)
	}
}

// Обработка команды /add

func (bh *BotHandler) handleAddCommand(update tgbotapi.Update, msg tgbotapi.MessageConfig, userID int) {
	bh.AddArgs = update.Message.CommandArguments()
	if bh.AddArgs == "" {
		msg.Text += "Чтобы добавить аббревиатуру используйте формат:\n/add [аббревиатура] [значение]"
		SendBot(msg, bh.bot)
	} else {
		abbreviation := strings.Fields(bh.AddArgs)
		rows, err := bh.db.Query("SELECT meaning, author FROM abbrevia.abbreviations WHERE abbreviation = $1", strings.ToUpper(abbreviation[0]))

		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		found := false // Флаг для отслеживания нахождения аббревиатуры

		if rows.Next() {
			msg.Text = "Найдены следующие совпадения:"
			SendBot(msg, bh.bot)
		}

		for rows.Next() {
			var meaning, author string
			if err := rows.Scan(&meaning, &author); err != nil {
				// Обработка ошибки
				log.Fatal(err)
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Аббревиатура: "+strings.ToUpper(abbreviation[0])+"\nЗначение: "+meaning)
			if author != "" {
				msg.Text += "\nАвтор: " + author
			}
			SendBot(msg, bh.bot)
			found = true
		}

		if !found {
			args := strings.Fields(update.Message.Text)
			if len(args) < 2 {
				msg.Text = "Используйте команду в следующем формате: [аббревиатура] [значение]"
			}
			if len(args) >= 2 {
				abbreviation := strings.ToUpper(args[1])
				meaning := strings.Join(args[2:], " ")
				author := "@" + update.Message.From.UserName
				// Сохранение аббревиатуры и значения в базу данных
				_, err := bh.db.Exec("INSERT INTO abbrevia.abbreviations (abbreviation, meaning, author) VALUES ($1, $2, $3)", abbreviation, meaning, author)
				if err == nil {
					msg.Text = "Аббревиатура успешно сохранена в базе данных."
					delete(bh.userStates, userID)
				} else {
					msg.Text = "Не удалось сохранить аббревиатуру. Пожалуйста, попробуйте еще раз.\n"
				}
			}
			//Отправка ответа ботом
			SendBot(msg, bh.bot)
		} else {
			msg.Text = "Добавить новую запись?"
			bh.AddFlag = true

			msg.ReplyMarkup = Keyboard()

			SendBot(msg, bh.bot)

			// Установка состояния ожидания добавления аббревиатуры
			bh.userStates[userID] = &UserState{WaitingForAbbreviation: true}
		}
	}
}

// Действия если это сообщение пользователя без команды

func (bh *BotHandler) HandleNonCommandMessage(update tgbotapi.Update, msg tgbotapi.MessageConfig) {
	userID := update.Message.From.ID
	state, ok := bh.userStates[userID]

	if ok && state.WaitingForAbbreviation {
		// Логика обработки сообщений при ожидании аббревиатуры
		bh.handleWaitingForAbbreviation(update, msg, userID)
	} else {
		// Логика обработки сообщений без ожидания аббревиатуры
		bh.handleNonWaitingForAbbreviation(update, msg, userID, state)
	}
}

// Логика обработки сообщений при ожидании аббревиатуры

func (bh *BotHandler) handleWaitingForAbbreviation(update tgbotapi.Update, msg tgbotapi.MessageConfig, userID int) {
	if update.Message.Text == "Нет" {
		// Если пользователь выбрал "Нет", завершаем состояние ожидания и не сохраняем аббревиатуру
		delete(bh.userStates, userID)
		msg.Text = "Ожидание добавления аббревиатуры завершено."

		// Отправляем пустую клавиатуру, чтобы закрыть существующую клавиатуру
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

		SendBot(msg, bh.bot)
		bh.AddFlag = false

	} else if update.Message.Text == "Да" {
		if bh.AddFlag {
			args := strings.Fields(bh.AddArgs)
			abbreviation := strings.ToUpper(args[0])
			meaning := strings.Join(args[1:], " ")
			author := "@" + update.Message.From.UserName
			// Сохранение аббревиатуры и значения в базу данных
			_, err := bh.db.Exec("INSERT INTO abbrevia.abbreviations (abbreviation, meaning, author) VALUES ($1, $2, $3)", abbreviation, meaning, author)
			if err == nil {
				msg.Text = "Аббревиатура успешно сохранена в базе данных. "
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				delete(bh.userStates, userID)
				bh.AddFlag = false
			}
		} else {
			msg.Text = "Добавьте аббревиатуру в следующем формате:\n [аббревиатура] [значение]."
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		}
		SendBot(msg, bh.bot)
	} else {
		args := strings.Fields(update.Message.Text)
		if len(args) < 2 {
			msg.Text = "Используйте команду в следующем формате: /save [аббревиатура] [значение]"
		}
		if len(args) >= 2 {
			abbreviation := strings.ToUpper(args[0])
			meaning := strings.Join(args[1:], " ")
			author := "@" + update.Message.From.UserName
			// Сохранение аббревиатуры и значения в базу данных
			_, err := bh.db.Exec("INSERT INTO abbrevia.abbreviations (abbreviation, meaning, author) VALUES ($1, $2, $3)", abbreviation, meaning, author)
			if err == nil {
				msg.Text = "Аббревиатура успешно сохранена в базе данных."
				delete(bh.userStates, userID)
			} else {
				msg.Text = "Не удалось сохранить аббревиатуру. Пожалуйста, попробуйте еще раз.\n"
			}
		}
		SendBot(msg, bh.bot)
	}
}

// Логика обработки сообщений без ожидания аббревиатуры

func (bh *BotHandler) handleNonWaitingForAbbreviation(update tgbotapi.Update, msg tgbotapi.MessageConfig, userID int, state *UserState) {
	abbreviation := strings.ToUpper(update.Message.Text)

	rows, err := bh.db.Query("SELECT meaning, author FROM abbrevia.abbreviations WHERE abbreviation = $1", abbreviation)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	found := false // Флаг для отслеживания нахождения аббревиатуры

	for rows.Next() {
		var meaning, author string
		if err := rows.Scan(&meaning, &author); err != nil {
			log.Fatal(err)
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Аббревиатура: "+abbreviation+"\nЗначение: "+meaning)
		if author != "" {
			msg.Text += "\nАвтор: " + author
		}
		SendBot(msg, bh.bot)
		found = true
	}
	if !found {
		if state == nil {
			msg.Text = "Аббревиатура не найдена в базе данных. Хотите добавить ее?"

			msg.ReplyMarkup = Keyboard()

			SendBot(msg, bh.bot)

			// Установка состояния ожидания добавления аббревиатуры
			bh.userStates[userID] = &UserState{WaitingForAbbreviation: true}
		}
	}
}

// Создание клавиатуры для предложения добавления аббревиатуры

func Keyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Да"),
			tgbotapi.NewKeyboardButton("Нет"),
		),
	)
	return keyboard
}

// Отправка сообщения в чат

func SendBot(msg tgbotapi.MessageConfig, bot *tgbotapi.BotAPI) {
	_, err := bot.Send(msg)
	if err != nil {
		log.Panic(err)
	}
}

// Чтение файла credentials.txt

func ReadCredentials(filename string) (map[string]string, error) {
	// Открываем файл для чтения
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	credentials := make(map[string]string)

	// Читаем построчно из файла
	for scanner.Scan() {
		line := scanner.Text()
		// Разбиваем строку на метку и значение по " "
		parts := strings.Split(line, " ")
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Сохраняем значение в map
			credentials[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при чтении файла: %w", err)
	}

	return credentials, nil
}

func main() {
	// Файл с кредами для подключения к Telegramm API и PostgreSQL
	filename := "credentials.txt"

	// Чтение файла credentials.txt
	credentials, err := ReadCredentials(filename)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}

	// Подключение к Telegramm API
	bot, err := tgbotapi.NewBotAPI(credentials["Token"])
	if err != nil {
		log.Panic(err)
	}

	// Подключение к Базе

	DataSourceName := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		credentials["User"], credentials["Password"], credentials["DbName"])

	db, err := sql.Open("postgres", DataSourceName)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	bot.Debug = true
	log.Printf("Авторизован как %s", bot.Self.UserName)

	// Создание структуры типов данных для методов
	botHandler := NewBotHandler(bot, db)
	// Запуск бота
	botHandler.Start()
}
