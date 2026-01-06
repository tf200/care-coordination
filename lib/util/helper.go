package util

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
