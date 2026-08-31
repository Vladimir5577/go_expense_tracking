package helper

import "testing"

func TestRoundMoney(t *testing.T) {
	tests := []struct {
		in   float64
		want float64
	}{
		{123.456, 123.46},
		{123.454, 123.45},
		{123.45, 123.45},
		{123, 123},
		{0.005, 0.01},
		{1543.2, 1543.2},
		{99.999, 100},
	}

	for _, tc := range tests {
		if got := RoundMoney(tc.in); got != tc.want {
			t.Errorf("RoundMoney(%v) = %v, ожидалось %v", tc.in, got, tc.want)
		}
	}
}
