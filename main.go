package main

import (
	"flag"
	"log"

	max "github.com/maxigo-bot/maxigo-bot"
)

var (
	// ───────────── PATHS ─────────────
	path = "schedule.json"
	// ───────────── REGISTER ─────────────
	week    *WeekSchedule
	weekErr error
	// ─────────── ARGUMENTS ───────────
	token = ""
	ownID = ""
)

func main() {
	flag.StringVar(&token, "token", token, "MAX Bot api token")
	flag.StringVar(&ownID, "owner-id", ownID, "ID for logging/administrer")
	flag.StringVar(&path, "table-path", path, "path to table with data")
	flag.Parse()
	week, weekErr = LoadWeekSchedule(path)
	if token == "" || weekErr != nil {
		log.Fatal("Где-т ошибка")
	}
	bot, err := max.New(token, max.WithLongPolling(30))
	if err != nil {
		log.Fatalf("не удалось инициализировать бота: %v", err)
	}

	bot.OnError = func(err error, c max.Context) {
		log.Printf("ошибка обработчика: %v", err)
	}
	initComs(bot)
	bot.Start()
}
func initComs(bot *max.Bot) {
	startCom(bot)
}

func startCom(bot *max.Bot) {
	bot.Handle("/start", func(c max.Context) error {
		name := "друг"
		if sender := c.Sender(); sender != nil && sender.FirstName != "" {
			name = sender.FirstName
		}
		text := "Привет, " + name + "! 👋\n\n" +
			"Я твой умный одноклассник — помогу тебе освоиться на новой ступени обучения:\n" +
			"	• покажу расписание\n" +
			"	• посчитаю, сколько осталось до конца урока или перемены\n" +
			"	• подскажу имя учителя\n" +
			"	• передам задания\n\n" +
			"Просто напиши, что тебя интересует!"
		return c.Send(text)
	})
}
