package helper

import "testing"

func TestResolvePeriod(t *testing.T) {
	tests := []struct {
		name   string
		period string
		anchor string
		from   string
		to     string
	}{
		{"день", PeriodDay, "2026-08-30", "2026-08-30", "2026-08-30"},

		// Неделя начинается с понедельника, а не с воскресенья.
		{"неделя от воскресенья", PeriodWeek, "2026-08-30", "2026-08-24", "2026-08-30"},
		{"неделя от понедельника", PeriodWeek, "2026-08-31", "2026-08-31", "2026-09-06"},
		{"неделя через границу месяца", PeriodWeek, "2026-09-02", "2026-08-31", "2026-09-06"},
		{"неделя через границу года", PeriodWeek, "2026-01-01", "2025-12-29", "2026-01-04"},

		{"месяц", PeriodMonth, "2026-08-30", "2026-08-01", "2026-08-31"},
		{"месяц из первого дня", PeriodMonth, "2026-08-01", "2026-08-01", "2026-08-31"},
		{"месяц из последнего дня", PeriodMonth, "2026-08-31", "2026-08-01", "2026-08-31"},
		{"февраль високосного года", PeriodMonth, "2024-02-15", "2024-02-01", "2024-02-29"},
		{"февраль обычного года", PeriodMonth, "2026-02-15", "2026-02-01", "2026-02-28"},
		{"декабрь", PeriodMonth, "2026-12-31", "2026-12-01", "2026-12-31"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			anchor, err := ParseDate(tc.anchor)
			if err != nil {
				t.Fatalf("ParseDate: %v", err)
			}

			got, err := ResolvePeriod(tc.period, anchor)
			if err != nil {
				t.Fatalf("ResolvePeriod: %v", err)
			}
			if got.From != tc.from || got.To != tc.to {
				t.Errorf("ResolvePeriod(%s, %s) = [%s; %s], ожидалось [%s; %s]",
					tc.period, tc.anchor, got.From, got.To, tc.from, tc.to)
			}
		})
	}
}

func TestResolvePeriodUnknown(t *testing.T) {
	anchor, _ := ParseDate("2026-08-30")
	if _, err := ResolvePeriod("year", anchor); err == nil {
		t.Fatal("ожидалась ошибка для неизвестного периода")
	}
}

func TestParseDateInvalid(t *testing.T) {
	for _, s := range []string{"", "30-08-2026", "2026-8-30", "2026-13-01", "2026-02-30", "не дата"} {
		if _, err := ParseDate(s); err == nil {
			t.Errorf("ParseDate(%q) должен был вернуть ошибку", s)
		}
	}
}
