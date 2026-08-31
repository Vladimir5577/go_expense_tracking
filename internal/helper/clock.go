package helper

import "time"

// Clock даёт текущий момент (для тестируемости/мокабельности).
// created_at/updated_at всегда UTC, усечённые до секунды.
type Clock struct{}

func NewClock() Clock { return Clock{} }

// Now возвращает текущий момент в UTC, усечённый до секунды.
func (c Clock) Now() time.Time {
	return time.Now().UTC().Truncate(time.Second)
}

// NowString возвращает текущий момент в формате RFC3339 — в этом виде время
// хранится в TEXT-колонках SQLite.
func (c Clock) NowString() string {
	return c.Now().Format(time.RFC3339)
}
