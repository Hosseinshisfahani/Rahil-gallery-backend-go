package customer

import "testing"

func TestStringSliceOrEmpty(t *testing.T) {
	t.Run("nil slice becomes non-nil empty", func(t *testing.T) {
		got := StringSliceOrEmpty(nil)
		if got == nil {
			t.Fatal("expected non-nil empty slice")
		}
		if len(got) != 0 {
			t.Fatalf("expected length 0, got %d", len(got))
		}
	})

	t.Run("empty slice stays non-nil empty", func(t *testing.T) {
		got := StringSliceOrEmpty([]string{})
		if got == nil {
			t.Fatal("expected non-nil empty slice")
		}
		if len(got) != 0 {
			t.Fatalf("expected length 0, got %d", len(got))
		}
	})

	t.Run("values are copied", func(t *testing.T) {
		src := []string{"gold_and_stones"}
		got := StringSliceOrEmpty(src)
		if len(got) != 1 || got[0] != "gold_and_stones" {
			t.Fatalf("unexpected slice: %#v", got)
		}
		src[0] = "changed"
		if got[0] == "changed" {
			t.Fatal("expected defensive copy")
		}
	})
}

func TestNormalizeDigits(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"persian phone", "۰۹۱۳۲۰۴۷۱۰۶", "09132047106"},
		{"arabic phone", "٠٩١٣٢٠٤٧١٠٦", "09132047106"},
		{"already english", "09132047106", "09132047106"},
		{"mixed with plus", "+۹۸۹۱۳۲۰۴۷۱۰۶", "+989132047106"},
		{"non-digits preserved", "tel: ۰۹۱۳", "tel: 0913"},
		{"empty", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := NormalizeDigits(tc.in); got != tc.want {
				t.Fatalf("NormalizeDigits(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNormalizeInputConvertsPhoneDigits(t *testing.T) {
	in := &Input{Phone: " ۰۹۱۳۲۰۴۷۱۰۶ "}
	NormalizeInput(in)
	if in.Phone != "09132047106" {
		t.Fatalf("expected normalized phone, got %q", in.Phone)
	}
}
