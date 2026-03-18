package routes

import "testing"

func TestItoa(t *testing.T) {
	cases := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{9, "9"},
		{10, "10"},
		{100, "100"},
		{12345, "12345"},
		{1000000, "1000000"},
	}
	for _, tc := range cases {
		got := itoa(tc.input)
		if got != tc.want {
			t.Errorf("itoa(%d) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
