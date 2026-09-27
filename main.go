package main

import (
	"flag"
	"log"

	maxigobot "github.com/maxigo-bot/maxigo-bot"
)

var (
	// ───────────── PATHS ─────────────
	path    = "schedule.json"
	// ───────────── REGISTER ─────────────
	week    *WeekSchedule
	weekErr error
	// ─────────── ARGUMENTS ───────────
	token = ""
	ownID = ""
)

func main() {
	flag.StringVar(&token, "token", token, "MAX Bot api token")
	flag.StringVar(&ownID, "owner-id", ownID, "ID for logging/administer")
	flag.StringVar(&path, "table-path", path, "path to table with data")
	flag.Parse()
	week, weekErr = LoadWeekSchedule(path)
	if token == "" || weekErr != nil {
		log.Fatal("Гдет ошибка")
	}

	bot, err := maxigobot.New(token, maxigobot.WithLongPolling(30))
	if err != nil {
		log.Fatalf("не удалось инициализировать бота: %v", err)
	}

	bot.OnError = func(err error, c maxigobot.Context) {
		log.Printf("ошибка обработчика: %v", err)
	}
	initComs(bot)
	bot.Start()
}
func initComs(bot *maxigobot.Bot) {
	bot.Handle("/start", handleStart)

}

func handleStart(c maxigobot.Context) error {
	name := "друг"
	if sender := c.Sender(); sender != nil && sender.FirstName != "" {
		name = sender.FirstName
	}

	text := "Привет, " + name + "! 👋\n\n" +
		"Я YourSmartClassmate — помогу тебе освоиться на новой ступени обучения:\n" +
		"• покажу расписание\n" +
		"• посчитаю, сколько осталось до конца урока или перемены\n" +
		"• подскажу имя учителя\n" +
		"• передам задания\n\n" +
		"Просто напиши, что тебя интересует!"

	return c.Send(text)
}
