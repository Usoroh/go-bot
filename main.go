package main

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func dbConn() (db *sql.DB) {
	// dbDriver := "mysql"
	// // dbUser := "bd64185d03cbcet"
	// // dbPass := "08c17f4b"
	// // dbName := "heroku_4438dd451a96a65"
	// db, err := sql.Open(dbDriver, "bc536a91185fda:791fb1bb@tcp(us-cdbr-gcp-east-01.cleardb.net:3306)/gcp_74865c3e4a85c95dfa0c")
	// if err != nil {
	// 	fmt.Println(err)
	// 	panic(err.Error())
	// }
	// return db
	dbDriver := "mysql"
	// dbUser := "root"
	// dbPass := ""
	// dbName := "habits"
	db, err := sql.Open(dbDriver, "b0f1a606882b71:b5ed3ec0@tcp(us-cdbr-gcp-east-01.cleardb.net:3306)/gcp_74865c3e4a85c95dfa0c")
	if err != nil {
		fmt.Println("KOOOOOOOOOL")
		panic(err.Error())
	}
	return db
}

func newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

var buttons = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonURL("Sign-in", "https://myhabitscreator.herokuapp.com/signin"),
		tgbotapi.NewInlineKeyboardButtonURL("Sign-up", "https://myhabitscreator.herokuapp.com/signup"),
	),
)

// var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
// 	tgbotapi.NewInlineKeyboardRow(
// 		tgbotapi.NewInlineKeyboardButtonURL("1.com", "http://1.com"),
// 		tgbotapi.NewInlineKeyboardButtonSwitch("2sw", "open 2"),
// 		tgbotapi.NewInlineKeyboardButtonData("3", "3"),
// 	),
// 	tgbotapi.NewInlineKeyboardRow(
// 		tgbotapi.NewInlineKeyboardButtonData("4", "4"),
// 		tgbotapi.NewInlineKeyboardButtonData("5", "5"),
// 		tgbotapi.NewInlineKeyboardButtonData("6", "6"),
// 	),
// )

type Habit struct {
	Name     string
	Days     int
	DaysDone int
}

func main() {

	db := dbConn()
	statement, err := db.Prepare("CREATE TABLE IF NOT EXISTS users (id INTEGER AUTO_INCREMENT PRIMARY KEY, username TEXT, password TEXT, admin INTEGER, UNIQUE(username))")
	statement.Exec()
	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS habits (id INTEGER AUTO_INCREMENT PRIMARY KEY, habit TEXT, username TEXT, days INTEGER, daysDone INTEGER)")
	statement.Exec()
	if err != nil {
		fmt.Println(err)
	}
	db.Close()

	bot, err := tgbotapi.NewBotAPI("853233315:AAHJ9SmdNd906chqjL703nX-lUl0QaKUEKI")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.CallbackQuery != nil {
			fmt.Println(update)

			bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))
			bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data))
		}

		if update.Message.IsCommand() {
			// msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			switch update.Message.Command() {
			case "start":
				//connect to a database and add user there if not exists
				db := dbConn()
				username := update.Message.From.UserName
				fmt.Println("USERNAME: ", username)
				statement, err := db.Prepare("INSERT INTO users (username, password, admin) VALUES (?, ?, ?)")
				if err != nil {
					fmt.Println(err)
				}
				statement.Exec(username, "keker", false)
				db.Close()
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "List of commands\nAdd a habit: /new <habitname> <days>\nSee habits: /habits")
				bot.Send(msg)
			case "new":
				habit := strings.Split(update.Message.Text, " ")
				if len(habit) > 1 {
					fmt.Println("HABIT: ", habit[1])
					db := dbConn()
					statement, err := db.Prepare("INSERT INTO habits (habit, username, days, daysDone) VALUES (?, ?, ?, ?)")
					if err != nil {
						fmt.Println(err)
					}
					username := update.Message.From.UserName
					days, _ := strconv.Atoi(habit[2])
					fmt.Println("DAYS: ", days)
					fmt.Println("USER", username)
					statement.Exec(strings.ToLower(habit[1]), username, days, 0)
					db.Close()
					if username == "dinadinus" {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You've added a habit! Now try to stick to it :). Also, I love you Dina!")
						bot.Send(msg)
					} else {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You've added a habit! Now try to stick to it :)")
						bot.Send(msg)
					}

				}
			case "habits":
				db := dbConn()
				username := update.Message.From.UserName
				// var daysDone int
				// var days int
				var habits []Habit
				rows, err := db.Query("SELECT habit, days, daysDone FROM habits WHERE username = ?", username)
				if err == nil {
					for rows.Next() {
						h := Habit{}
						if err := rows.Scan(&h.Name, &h.Days, &h.DaysDone); err != nil {
							return
						}
						habits = append(habits, h)
					}
				}
				str := ""
				for i := 0; i < len(habits); i++ {
					str = str + habits[i].Name + ": " + strconv.Itoa(habits[i].DaysDone) + "/" + strconv.Itoa(habits[i].Days) + "\n"
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, str)
				bot.Send(msg)
				db.Close()
			default:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "I don't know that command :(")
				bot.Send(msg)
			}
			// bot.Send(msg)
		} else if strings.Contains(update.Message.Text, "+") || strings.Contains(update.Message.Text, "-") {
			str := strings.Split(update.Message.Text, " ")
			var habit string
			if len(str) == 2 {
				for i := 0; i < len(str); i++ {
					if str[i] != "+" && str[i] != "-" {
						habit = strings.ToLower(str[i])
					}
				}

				db := dbConn()
				stmt := ""
				txt := ""
				if strings.Contains(update.Message.Text, "+") {
					stmt = "UPDATE habits SET daysDone = daysDone + 1 WHERE username = ? AND habit = ?"
					txt = "Great, keep it up"
				} else if strings.Contains(update.Message.Text, "-") {
					stmt = "UPDATE habits SET daysDone = daysDone - 1 WHERE username = ? AND habit = ?"
					txt = "Ouch! What happened??"
				}
				statement, _ := db.Prepare(stmt)
				username := update.Message.From.UserName
				statement.Exec(username, habit)
				db.Close()

				feature := ""
				if username == "dinadinus" {
					feature = "Dina, you are so cool!"
				}

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, txt+feature)
				bot.Send(msg)
			}
		}
	}
}
