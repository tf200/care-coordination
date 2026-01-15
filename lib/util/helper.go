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
