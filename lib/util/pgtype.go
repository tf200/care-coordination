package util

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func PgtypeTimeToString(t pgtype.Time) string {
	if !t.Valid {
		return ""
	}

	// 1. Convert the microseconds count to a time.Duration
	offset := time.Duration(t.Microseconds) * time.Microsecond

	// 2. Add that duration to a base "midnight" time (0000-01-01 00:00:00 UTC)
	// We use UTC to avoid timezone/DST shifts affecting the calculation.
	tm := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC).Add(offset)

	// 3. Format using the strict constant "15:04:05"
	return tm.Format(time.TimeOnly)
}
func StrToPgtypeTime(s string) pgtype.Time {
	// Try parsing with full time format (15:04:05) first
	parsedTime, err := time.Parse(time.TimeOnly, s)
	if err != nil {
		// Fall back to short format (15:04)
		parsedTime, err = time.Parse("15:04", s)
		if err != nil {
			return pgtype.Time{Valid: false}
		}
	}

	midnight := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)

	return pgtype.Time{
		Microseconds: parsedTime.Sub(midnight).Microseconds(),
		Valid:        true,
	}
}

func TimeToPgtypeTime(t time.Time) pgtype.Time {
	midnight := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
	return pgtype.Time{
		Microseconds: t.Sub(midnight).Microseconds(),
		Valid:        true,
	}
}

func StrToPgtypeDate(s string) pgtype.Date {
	// Parse strictly enforces "YYYY-MM-DD"
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return pgtype.Date{Valid: false}
	}

	return pgtype.Date{Time: t, Valid: true}
}

func TimeToPgtypeDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

func PgtypeDateToStr(d pgtype.Date) string {
	if !d.Valid {
		return ""
	}
	return d.Time.Format(time.DateOnly)
}

func PgtypeTimestampToStr(t pgtype.Timestamp) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format(time.RFC3339)
}

func Map[T any, R any](items []T, f func(T) R) []R {
	result := make([]R, len(items))
	for i, item := range items {
		result[i] = f(item)
	}
	return result
}

func IntToPointerInt32(v *int) *int32 {
	if v == nil {
		return nil
	}
	v32 := int32(*v)
	return &v32
}

func PointerInt32ToInt(v *int32) *int {
	if v == nil {
		return nil
	}
	vInt := int(*v)
	return &vInt
}

func PointerInt32ToIntValue(v *int32) int {
	if v == nil {
		return 0
	}
	return int(*v)
}
