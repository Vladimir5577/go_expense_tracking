package helper

import (
	"fmt"
	"time"

	"go_expense_service/internal/apperr"
)

// DateLayout — формат календарной даты, в котором хранится expenses.spent_at
// и в котором даты ходят через API.
const DateLayout = "2006-01-02"

const (
	PeriodDay   = "day"
	PeriodWeek  = "week"
	PeriodMonth = "month"
)

// DateRange — включительный диапазон дат [From; To].
type DateRange struct {
	From string
	To   string
}

// ParseDate разбирает дату в формате YYYY-MM-DD.
// Время не участвует: spent_at — календарная дата, а не момент времени.
func ParseDate(s string) (time.Time, error) {
	d, err := time.ParseInLocation(DateLayout, s, time.UTC)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: invalid date %q, expected YYYY-MM-DD", apperr.ErrValidation, s)
	}
	return d, nil
}

// Today возвращает сегодняшнюю календарную дату в заданной таймзоне.
// Именно таймзона решает, какой день считается «сегодня» — для пользователя в
// Москве в 01:00 сегодня уже наступило, а по UTC ещё нет.
func Today(loc *time.Location) time.Time {
	n := time.Now().In(loc)
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
}

// ResolvePeriod разворачивает период (day|week|month) в диапазон дат вокруг
// опорной даты anchor. Неделя начинается с понедельника.
func ResolvePeriod(period string, anchor time.Time) (DateRange, error) {
	switch period {
	case PeriodDay:
		d := anchor.Format(DateLayout)
		return DateRange{From: d, To: d}, nil

	case PeriodWeek:
		// time.Weekday: воскресенье = 0. Приводим к «дней прошло с понедельника».
		offset := (int(anchor.Weekday()) + 6) % 7
		start := anchor.AddDate(0, 0, -offset)
		return DateRange{
			From: start.Format(DateLayout),
			To:   start.AddDate(0, 0, 6).Format(DateLayout),
		}, nil

	case PeriodMonth:
		start := time.Date(anchor.Year(), anchor.Month(), 1, 0, 0, 0, 0, anchor.Location())
		// +1 месяц −1 день даёт последний день месяца без таблицы длин месяцев
		// и с корректным февралём в високосный год.
		return DateRange{
			From: start.Format(DateLayout),
			To:   start.AddDate(0, 1, -1).Format(DateLayout),
		}, nil

	default:
		return DateRange{}, fmt.Errorf("%w: unknown period %q, expected day|week|month", apperr.ErrValidation, period)
	}
}
