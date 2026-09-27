package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)
//
// FILE 1
// 
const dateLayout = "02.01.06"

var (
	// break time in monday n other days
	mondaySchedule = []Period{
		P(8, 30, 9, 10),
		P(9, 15, 9, 55),
		P(10, 5, 10, 45),
		P(11, 5, 11, 45),
		P(11, 55, 12, 35),
		P(12, 55, 13, 35),
		P(13, 45, 14, 25),
	}
	weekSchedule = []Period{
		P(8, 30, 9, 10),
		P(9, 20, 10, 0),
		P(10, 10, 10, 50),
		P(11, 10, 11, 50),
		P(12, 0, 12, 40),
		P(13, 0, 13, 40),
		P(13, 50, 14, 30),
	}
)

type Period struct {
	Start time.Duration
	End   time.Duration
}

type LessonStatus struct {
	IsLesson    bool
	Finished    bool
	LessonIndex int
	TimeLeft    time.Duration
	NextBreak   time.Duration
}

type LessonEntry struct {
	Number  int
	Subject string
	Room    int
	Addi    bool
	IsIT    bool
	IsSE    bool
}

type ScheduleDay struct {
	Date    time.Time
	Entries []LessonEntry
}

type WeekSchedule struct {
	Days map[string]ScheduleDay `json:"days"` // ключ: "2006-01-02"
	mu   sync.Mutex
}

// ───────────── WeekSchedule ─────────────

func (w *WeekSchedule) SetDay(day ScheduleDay) {
	w.mu.Lock()
	defer w.mu.Unlock()
	key := day.Date.Format("2006-01-02")
	w.Days[key] = day
}

func (w *WeekSchedule) GetDay(date time.Time) (ScheduleDay, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	day, ok := w.Days[date.Format("2006-01-02")]
	return day, ok
}

func (w *WeekSchedule) Save(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	data, err := json.MarshalIndent(w.Days, "", "  ")
	if err != nil {
		return fmt.Errorf("маршалинг: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func LoadWeekSchedule(path string) (*WeekSchedule, error) {
	w := &WeekSchedule{Days: make(map[string]ScheduleDay)}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return w, nil
	}
	if err != nil {
		return nil, fmt.Errorf("чтение файла: %w", err)
	}
	if err := json.Unmarshal(data, &w.Days); err != nil {
		return nil, fmt.Errorf("разбор JSON: %w", err)
	}
	return w, nil
}

func (w *WeekSchedule) GetDayByDate(dateStr string) (ScheduleDay, bool) {
	t, err := time.Parse(dateLayout, dateStr)
	if err != nil {
		return ScheduleDay{}, false
	}
	return w.GetDay(t)
}
func (w *WeekSchedule) DeleteDay(date time.Time) error {
	w.mu.Lock()
	key := date.Format("2006-01-02")
	if _, ok := w.Days[key]; !ok {
		w.mu.Unlock()
		return fmt.Errorf("дата %s не найдена", key)
	}
	delete(w.Days, key)
	w.mu.Unlock()
	return w.Save(path)
}

// ───────────── ScheduleDay ─────────────

func (d ScheduleDay) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Расписание на %s:\n", d.Date.Format(dateLayout))
	if len(d.Entries) == 0 {
		sb.WriteString("(пусто)")
		return sb.String()
	}
	for _, e := range d.Entries {
		fmt.Fprintf(&sb, "%d. %s (каб. %d)\n", e.Number, e.Subject, e.Room)
	}
	return strings.TrimRight(sb.String(), "\n")
}

func dayFinished(day ScheduleDay, now time.Time) bool {
	periods := weekSchedule
	if now.Weekday() == time.Monday {
		periods = mondaySchedule
	}
	if len(day.Entries) < len(periods) {
		periods = periods[:len(day.Entries)]
	}
	if len(periods) == 0 {
		return true
	}
	last := periods[len(periods)-1]
	return now.Sub(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())) >= last.End
}

func (d ScheduleDay) FindEntry(number int) (LessonEntry, bool) {
	for _, e := range d.Entries {
		if e.Number == number {
			return e, true
		}
	}
	return LessonEntry{}, false
}

func ParseSchedule(r io.Reader) (ScheduleDay, error) {
	var day ScheduleDay
	dateParsed := false
	scanner := bufio.NewScanner(r)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if !dateParsed {
			date, err := time.Parse(dateLayout, line)
			if err != nil {
				return ScheduleDay{}, fmt.Errorf("строка %d: некорректная дата %q: %w", lineNum, line, err)
			}
			day.Date = date
			dateParsed = true
			continue
		}
		
		parts := strings.Split(line, "|")
		if len(parts) != 3 {
			return ScheduleDay{}, fmt.Errorf("строка %d: ожидалось 3 поля, получено %d (%q)", lineNum, len(parts), line)
		}
		
		for i, prt := range parts {
			parts[i] = strings.TrimSpace(prt)
		}
		
		number, err := strconv.Atoi(parts[0])
		if err != nil {
			return ScheduleDay{}, fmt.Errorf("строка %d: некорректный номер урока: %w", lineNum, err)
		}
		
		additLes := utf8.RuneCountInString(parts[0]) > 1
		
		it := strings.Contains(parts[1], "и.т.")
		se := strings.Contains(parts[1], "с.э.")
		subj := strings.TrimSuffix(parts[1], "(и.т.)")
		subj = strings.TrimSuffix(subj, "(с.э.)")
		
		room, err := strconv.Atoi(parts[2])
		if err != nil {
			return ScheduleDay{}, fmt.Errorf("строка %d: некорректный номер кабинета: %w", lineNum, err)
		}
		day.Entries = append(day.Entries, LessonEntry{
			Number:  number,
			Subject: subj,
			Room:    room,
			Addi:    additLes,
			IsIT:    it,
			IsSE:    se,
		})
	}
	
	if err := scanner.Err(); err != nil {
		return ScheduleDay{}, fmt.Errorf("ошибка чтения: %w", err)
	}
	if !dateParsed {
		return ScheduleDay{}, fmt.Errorf("дата не найдена")
	}
	return day, nil
}

// ───────────── Period ─────────────

func P(startH, startM, endH, endM int) Period {
	return Period{
		Start: time.Duration(startH)*time.Hour + time.Duration(startM)*time.Minute,
		End:   time.Duration(endH)*time.Hour + time.Duration(endM)*time.Minute,
	}
}

// ───────────── LessonStatus ─────────────

func (s LessonStatus) String() string {
	if s.Finished {
		return "УРОКОВ НЕЕТ 😝😝🤟🤘"
	}
	fmtLeft := fmtDur(s.TimeLeft)
	fmtBreak := fmtDur(s.NextBreak)
	if s.IsLesson {
		return fmt.Sprintf(
			"идёт урок №%d, до конца %s, после него перемена %s",
			s.LessonIndex+1, fmtLeft, fmtBreak,
		)
	}
	return fmt.Sprintf(
		"идёт перемена, до урока №%d осталось %s (длительность перемены: %s)",
		s.LessonIndex+1, fmtLeft, fmtBreak,
	)
}


func Status(now time.Time) LessonStatus {
	lessons := weekSchedule
	if now.Weekday() == time.Monday {
		lessons = mondaySchedule
	}
	n := now.Sub(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()))
	u, ok := week.GetDay(now)
	if !ok {
		return LessonStatus{}
	}
	maxNumber := 0
	for _, e := range u.Entries {
		if e.Number > maxNumber {
			maxNumber = e.Number
		}
	}
	if maxNumber > 0 && maxNumber < len(lessons) {
		lessons = lessons[:maxNumber]
	}

	for i, l := range lessons {
		switch {
		case (n >= l.Start && n < l.End):
			var next time.Duration
			if i+1 < len(lessons) {
				next = lessons[i+1].Start - l.End
			}
			return LessonStatus{
				IsLesson:    true,
				LessonIndex: i,
				TimeLeft:    l.End - n,
				NextBreak:   next,
			}
		case n < l.Start:
			left := l.Start - n
			return LessonStatus{
				IsLesson:    false,
				LessonIndex: i,
				TimeLeft:    left,
				NextBreak:   left,
			}
		}
	}
	return LessonStatus{Finished: true, LessonIndex: -1}
}

// ───────────── Utils ─────────────

func fmtDur(d time.Duration) string {
	d = d.Round(time.Second)
	m := d / time.Minute
	h := m / 60
	s := (d % time.Minute) / time.Second
	if s == 0 {
		return fmt.Sprintf("%d мин", m)
	}
	if m >= 60 {
		return fmt.Sprintf("%d час %d мин", h, m-60*h)
	}
	return fmt.Sprintf("%d мин %d сек", m, s)
}