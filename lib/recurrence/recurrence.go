package recurrence

import (
	"fmt"
	"time"

	"github.com/teambition/rrule-go"
)

// MaxOccurrences limits the number of occurrences to prevent infinite loops
const MaxOccurrences = 365

// ExpandRecurrence parses an RFC 5545 RRULE string and expands it into
// concrete occurrence times within the given date range.
//
// Parameters:
//   - rruleStr: The iCalendar RRULE string (e.g., "FREQ=WEEKLY;BYDAY=MO,WE,FR")
//   - dtstart: The start time of the original event (anchor for recurrence)
//   - rangeStart: Start of the date range to get occurrences for
//   - rangeEnd: End of the date range to get occurrences for
//
// Returns a slice of times representing each occurrence within the range.
func ExpandRecurrence(rruleStr string, dtstart, rangeStart, rangeEnd time.Time) ([]time.Time, error) {
	if rruleStr == "" {
		return nil, nil
	}

	// Parse the RRULE string
	r, err := rrule.StrToRRule(rruleStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse recurrence rule: %w", err)
	}

	// Set the start time (DTSTART)
	r.DTStart(dtstart)

	// Get occurrences between rangeStart and rangeEnd
	occurrences := r.Between(rangeStart, rangeEnd, true)

	// Limit occurrences to prevent excessive memory usage
	if len(occurrences) > MaxOccurrences {
		occurrences = occurrences[:MaxOccurrences]
	}

	return occurrences, nil
}

// CalculateOccurrenceEnd calculates the end time for an occurrence
// based on the original event's duration.
func CalculateOccurrenceEnd(occurrenceStart, originalStart, originalEnd time.Time) time.Time {
	duration := originalEnd.Sub(originalStart)
	return occurrenceStart.Add(duration)
}

// GenerateOccurrenceID creates a unique ID for a recurring occurrence
// Format: {appointmentID}_{YYYYMMDD}
func GenerateOccurrenceID(appointmentID string, occurrenceTime time.Time) string {
	return fmt.Sprintf("%s_%s", appointmentID, occurrenceTime.Format("20060102"))
}
