package recurrence

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandRecurrence(t *testing.T) {
	tests := []struct {
		name          string
		rruleStr      string
		dtstart       time.Time
		rangeStart    time.Time
		rangeEnd      time.Time
		wantCount     int
		wantErr       bool
		validateDates func(t *testing.T, dates []time.Time)
	}{
		{
			name:       "empty rrule returns nil",
			rruleStr:   "",
			dtstart:    time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC),
			rangeStart: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			rangeEnd:   time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC),
			wantCount:  0,
			wantErr:    false,
		},
		{
			name:       "daily recurrence",
			rruleStr:   "FREQ=DAILY",
			dtstart:    time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC),
			rangeStart: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			rangeEnd:   time.Date(2026, 1, 7, 23, 59, 59, 0, time.UTC),
			wantCount:  7,
			wantErr:    false,
			validateDates: func(t *testing.T, dates []time.Time) {
				for i, d := range dates {
					expectedDay := 1 + i
					assert.Equal(t, expectedDay, d.Day(), "Day %d should be %d", i, expectedDay)
				}
			},
		},
		{
			name:       "weekly on Monday, Wednesday, Friday",
			rruleStr:   "FREQ=WEEKLY;BYDAY=MO,WE,FR",
			dtstart:    time.Date(2026, 1, 5, 10, 0, 0, 0, time.UTC), // Monday
			rangeStart: time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
			rangeEnd:   time.Date(2026, 1, 11, 23, 59, 59, 0, time.UTC),
			wantCount:  3, // Mon 5, Wed 7, Fri 9
			wantErr:    false,
			validateDates: func(t *testing.T, dates []time.Time) {
				expectedDays := []int{5, 7, 9}
				for i, d := range dates {
					assert.Equal(t, expectedDays[i], d.Day(), "Occurrence %d day mismatch", i)
				}
			},
		},
		{
			name:       "monthly recurrence",
			rruleStr:   "FREQ=MONTHLY",
			dtstart:    time.Date(2026, 1, 15, 14, 0, 0, 0, time.UTC),
			rangeStart: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			rangeEnd:   time.Date(2026, 4, 30, 23, 59, 59, 0, time.UTC),
			wantCount:  4, // Jan 15, Feb 15, Mar 15, Apr 15
			wantErr:    false,
		},
		{
			name:       "with COUNT limit",
			rruleStr:   "FREQ=DAILY;COUNT=3",
			dtstart:    time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC),
			rangeStart: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			rangeEnd:   time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC),
			wantCount:  3,
			wantErr:    false,
		},
		{
			name:       "with UNTIL limit",
			rruleStr:   "FREQ=DAILY;UNTIL=20260105T235959Z",
			dtstart:    time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC),
			rangeStart: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			rangeEnd:   time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC),
			wantCount:  5, // Jan 1-5
			wantErr:    false,
		},
		{
			name:       "invalid rrule",
			rruleStr:   "INVALID_RULE",
			dtstart:    time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC),
			rangeStart: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			rangeEnd:   time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dates, err := ExpandRecurrence(tt.rruleStr, tt.dtstart, tt.rangeStart, tt.rangeEnd)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, dates, tt.wantCount)

			if tt.validateDates != nil {
				tt.validateDates(t, dates)
			}
		})
	}
}

func TestCalculateOccurrenceEnd(t *testing.T) {
	originalStart := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	originalEnd := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC) // 1.5 hour duration
	occurrenceStart := time.Date(2026, 1, 8, 9, 0, 0, 0, time.UTC)

	result := CalculateOccurrenceEnd(occurrenceStart, originalStart, originalEnd)

	expected := time.Date(2026, 1, 8, 10, 30, 0, 0, time.UTC)
	assert.Equal(t, expected, result)
}

func TestGenerateOccurrenceID(t *testing.T) {
	appointmentID := "apt-abc123"
	occurrenceTime := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)

	result := GenerateOccurrenceID(appointmentID, occurrenceTime)

	assert.Equal(t, "apt-abc123_20260315", result)
}
