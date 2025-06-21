package handlers

import "testing"

func TestNormalizePhoneNumber(t *testing.T) {
	tests := map[string]string{
		"1234567890":     "(123) 456-7890",
		"1-234-567-8901": "1-(234) 567-8901",
		"(123) 456-7890": "(123) 456-7890",
		"11234567890":    "1-(123) 456-7890",
	}

	for input, expected := range tests {
		got := normalizePhoneNumber(input)
		if got != expected {
			t.Errorf("Expected %s, got %s", expected, got)
		}
	}
}

func TestValidatePhoneNumber(t *testing.T) {
	valid := []string{"1234567890", "1-234-567-8901"}
	invalid := []string{"123", "abc", "999999"}

	for _, phone := range valid {
		if !validatePhoneNumber(phone) {
			t.Errorf("Expected %s to be valid", phone)
		}
	}

	for _, phone := range invalid {
		if validatePhoneNumber(phone) {
			t.Errorf("Expected %s to be invalid", phone)
		}
	}
}
