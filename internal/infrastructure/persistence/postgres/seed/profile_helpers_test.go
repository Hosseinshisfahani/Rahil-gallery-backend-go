package seed

import (
	"encoding/json"
	"testing"
	"time"
)

func TestFixtureImportProfile_includesAgeAndGender(t *testing.T) {
	t.Parallel()

	email := "sara@example.com"
	raw, err := fixtureImportProfile(
		"Sara", "Mohammadi", "+989121234567", &email,
		"vip", []string{"gold_and_gemstones"},
		"21-40", "female",
		time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		450,
	)
	if err != nil {
		t.Fatal(err)
	}

	var profile seedImportProfile
	if err := json.Unmarshal([]byte(raw), &profile); err != nil {
		t.Fatal(err)
	}
	if profile.CustomerAgeRange != "21-40" {
		t.Fatalf("customerAgeRange = %q", profile.CustomerAgeRange)
	}
	if profile.Gender != "female" {
		t.Fatalf("gender = %q", profile.Gender)
	}
}

func TestBuildCRMProfileJSON_includesAgeAndGender(t *testing.T) {
	t.Parallel()

	raw, err := buildCRMProfileJSON(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), 4)
	if err != nil {
		t.Fatal(err)
	}

	var profile bulkCRMProfile
	if err := json.Unmarshal([]byte(raw), &profile); err != nil {
		t.Fatal(err)
	}
	if profile.CustomerAgeRange == "" {
		t.Fatal("expected customerAgeRange to be set")
	}
	if profile.Gender == "" {
		t.Fatal("expected gender to be set for index 4")
	}
}

func TestBuildCRMProfileJSON_allowsMissingGender(t *testing.T) {
	t.Parallel()

	raw, err := buildCRMProfileJSON(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), 7)
	if err != nil {
		t.Fatal(err)
	}

	var profile bulkCRMProfile
	if err := json.Unmarshal([]byte(raw), &profile); err != nil {
		t.Fatal(err)
	}
	if profile.Gender != "" {
		t.Fatalf("gender = %q, want empty for index 7", profile.Gender)
	}
}
