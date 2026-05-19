package cmd

import (
	"testing"
	"time"
)

func TestParseWhen_Empty(t *testing.T) {
	fb := time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC)
	got, err := parseWhen("", fb, false)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(fb) {
		t.Errorf("empty did not return fallback")
	}
}

func TestParseWhen_DateOnly(t *testing.T) {
	t1, err := parseWhen("2026-05-11", time.Time{}, false)
	if err != nil {
		t.Fatal(err)
	}
	if t1.Hour() != 0 {
		t.Errorf("start-of-day expected, got %s", t1)
	}

	t2, err := parseWhen("2026-05-11", time.Time{}, true)
	if err != nil {
		t.Fatal(err)
	}
	if t2.Day() != 11 || t2.Hour() != 23 {
		t.Errorf("end-of-day expected, got %s", t2)
	}
}

func TestParseWhen_RFC3339(t *testing.T) {
	got, err := parseWhen("2026-05-11T15:04:05Z", time.Time{}, false)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 5, 11, 15, 4, 5, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("got %s want %s", got, want)
	}
}

func TestParseWhen_Invalid(t *testing.T) {
	if _, err := parseWhen("garbage", time.Time{}, false); err == nil {
		t.Error("expected error for invalid input")
	}
}
