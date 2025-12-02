package util

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

func PgtypeTimeTString(t pgtype.Time) string {
	if !t.Valid {
		return ""
	}
	ms := t.Microseconds
	s := formatTimeFromMicroseconds(ms)
	return s
}

func formatTimeFromMicroseconds(ms int64) string {
	hours := ms / 3600000000
	ms -= hours * 3600000000
	minutes := ms / 60000000
	ms -= minutes * 60000000
	seconds := ms / 1000000
	return formatTwoDigits(int(hours)) + ":" + formatTwoDigits(int(minutes)) + ":" + formatTwoDigits(int(seconds))
}

func formatTwoDigits(n int) string {
	if n < 10 {
		return "0" + fmt.Sprint(n)
	}
	return fmt.Sprint('0'+(n/10)) + fmt.Sprint('0'+(n%10))
}

func StrToPgtypeTime(s string) pgtype.Time {
	var t pgtype.Time
	var hours, minutes, seconds int64
	n, err := fmt.Sscanf(s, "%02d:%02d:%02d", &hours, &minutes, &seconds)
	if err != nil {
		t.Valid = false
		return t
	}
	if n != 3 {
		t.Valid = false
		return t
	}
	ms := hours*3600000000 + minutes*60000000 + seconds*1000000
	t.Microseconds = ms
	t.Valid = true
	return t
}
