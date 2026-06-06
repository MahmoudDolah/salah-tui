package api

import "testing"

func TestStripTimezone(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"05:30 (EDT)", "05:30"},
		{"12:00 (UTC)", "12:00"},
		{"05:30", "05:30"},
		{"", ""},
		{"19:30 (IST)", "19:30"},
	}
	for _, tc := range tests {
		got := stripTimezone(tc.input)
		if got != tc.want {
			t.Errorf("stripTimezone(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
