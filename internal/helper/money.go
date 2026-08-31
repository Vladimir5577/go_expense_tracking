package helper

import "math"

// RoundMoney приводит сумму к двум знакам после запятой.
//
// Клиент может прислать что угодно (123.456, результат деления и т.п.), а в БД
// стоит CHECK (amount = ROUND(amount, 2)) — округляем на входе сами, чтобы
// пользователь получал понятную сумму, а не отказ по constraint.
func RoundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}
