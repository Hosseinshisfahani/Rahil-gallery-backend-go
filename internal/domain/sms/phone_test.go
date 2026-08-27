package sms

import "testing"

func TestNormalizeReceptor(t *testing.T) {
	cases := []struct {
		in, want string
		ok       bool
	}{
		{"09940959065", "989940959065", true},
		{"+98 994 095 9065", "989940959065", true},
		{"989940959065", "989940959065", true},
		{"02188776655", "", false},
		{"", "", false},
	}
	for _, tc := range cases {
		got, err := NormalizeReceptor(tc.in)
		if tc.ok {
			if err != nil || got != tc.want {
				t.Fatalf("%q: got (%q,%v) want %q", tc.in, got, err, tc.want)
			}
		} else if err == nil {
			t.Fatalf("%q: expected error, got %q", tc.in, got)
		}
	}
}
