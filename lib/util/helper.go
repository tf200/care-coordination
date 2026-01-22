package util

import "encoding/json"

func StrPtr(s string) *string {
	return &s
}

func HandleNilString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func HandleNilBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func ParseJSONB(data []byte) map[string]any {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil
	}
	return result
}

func SortSlice[T any](slice []T, less func(i, j int) bool) {
	n := len(slice)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if !less(j, j+1) {
				slice[j], slice[j+1] = slice[j+1], slice[j]
			}
		}
	}
}
