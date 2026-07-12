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
