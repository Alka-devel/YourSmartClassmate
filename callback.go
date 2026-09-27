package main

import (
	"strings"
	"time"

	mapi "github.com/maxigo-bot/maxigo-bot"
)

func callback(bot *mapi.Bot) {
	bot.Handle(mapi.OnCallback(""), func(c mapi.Context) error {
		prefix, arg, _ := strings.Cut(c.Data(), ":")
		switch prefix {
		case "group":
			return handleGroupCallback(c, arg)
		}
		return c.Respond("")
	})
}

func handleGroupCallback(c mapi.Context, arg string) error {
	var g Group
	switch arg {
	case "it":
		g = InfTec
	case "se":
		g = SocEco
	default:
		return c.Respond("Неизвестная группа")
	}
	day, ok := week.GetDay(time.Now())
	if !ok {
		return c.Respond("Расписания на сегодня нет")
	}
	status := Status(time.Now())
	_, err := RenderScheduleImage(day, status.LessonIndex+1, fmtDur(status.TimeLeft), DefaultScale, g)
	if err != nil {
		return c.Respond("Ошибка рендера")
	}
	// картинку нужно сначала загрузить через client.UploadMedia/UploadPhoto,
	// получить payload вложения и уже им редактировать/отправлять сообщение
	return c.Respond("Группа выбрана")
}
