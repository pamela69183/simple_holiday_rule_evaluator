package simple_holiday_rule_evaluator

import (
	"testing"
	"time"
)

func mustFixed(t *testing.T, month, day int) Rule {
	t.Helper()
	r, err := Fixed(month, day)
	if err != nil {
		t.Fatalf("Fixed(%d, %d) returned error: %v", month, day, err)
	}
	return r
}

func mustNth(t *testing.T, month int, wd time.Weekday, n int) Rule {
	t.Helper()
	r, err := NthWeekday(month, wd, n)
	if err != nil {
		t.Fatalf("NthWeekday(%d, %v, %d) returned error: %v", month, wd, n, err)
	}
	return r
}

func assertDate(t *testing.T, r Rule, year int, want time.Time) {
	t.Helper()
	got, err := r.Date(year)
	if err != nil {
		t.Fatalf("Date(%d) returned error: %v", year, err)
	}
	if !got.Equal(want) {
		t.Fatalf("Date(%d) = %v, want %v", year, got, want)
	}
}

func TestFixedDate(t *testing.T) {
	r := mustFixed(t, 12, 25)
	assertDate(t, r, 2024, time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC))
	assertDate(t, r, 2023, time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC))
}

func TestFixedLeapDay(t *testing.T) {
	r := mustFixed(t, 2, 29)
	assertDate(t, r, 2024, time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC))
	if _, err := r.Date(2023); err == nil {
		t.Fatal("expected error for Feb 29 in non-leap year")
	}
}

func TestFixedValidation(t *testing.T) {
	if _, err := Fixed(0, 1); err == nil {
		t.Fatal("expected error for month 0")
	}
	if _, err := Fixed(13, 1); err == nil {
		t.Fatal("expected error for month 13")
	}
	if _, err := Fixed(4, 31); err == nil {
		t.Fatal("expected error for April 31")
	}
	if _, err := Fixed(2, 30); err == nil {
		t.Fatal("expected error for February 30")
	}
}

func TestNthWeekday(t *testing.T) {
	// Third Monday of January 2024 is the 15th (MLK Day).
	r := mustNth(t, 1, time.Monday, 3)
	assertDate(t, r, 2024, time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC))
	// Fourth Thursday of November 2024 is the 28th (Thanksgiving).
	r = mustNth(t, 11, time.Thursday, 4)
	assertDate(t, r, 2024, time.Date(2024, 11, 28, 0, 0, 0, 0, time.UTC))
}

func TestLastWeekday(t *testing.T) {
	// Last Monday of May 2024 is Memorial Day, May 27.
	r, err := LastWeekday(5, time.Monday)
	if err != nil {
		t.Fatalf("LastWeekday returned error: %v", err)
	}
	assertDate(t, r, 2024, time.Date(2024, 5, 27, 0, 0, 0, 0, time.UTC))
}

func TestNthWeekdayNegative(t *testing.T) {
	// -1 is the last occurrence, same as LastWeekday.
	r := mustNth(t, 5, time.Monday, -1)
	assertDate(t, r, 2024, time.Date(2024, 5, 27, 0, 0, 0, 0, time.UTC))
	// -2 is the second-to-last Monday of May 2024: May 20.
	r = mustNth(t, 5, time.Monday, -2)
	assertDate(t, r, 2024, time.Date(2024, 5, 20, 0, 0, 0, 0, time.UTC))
}

func TestNthWeekdayMissing(t *testing.T) {
	// February 2024 has only four Mondays, so the fifth does not exist.
	r := mustNth(t, 2, time.Monday, 5)
	if _, err := r.Date(2024); err == nil {
		t.Fatal("expected error for fifth Monday of February 2024")
	}
}

func TestNthWeekdayValidation(t *testing.T) {
	if _, err := NthWeekday(1, time.Monday, 0); err == nil {
		t.Fatal("expected error for n=0")
	}
	if _, err := NthWeekday(1, time.Monday, 6); err == nil {
		t.Fatal("expected error for n=6")
	}
	if _, err := NthWeekday(13, time.Monday, 1); err == nil {
		t.Fatal("expected error for month 13")
	}
}

func TestEaster(t *testing.T) {
	cases := []struct {
		year int
		want time.Time
	}{
		{2024, time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC)},
		{2023, time.Date(2023, 4, 9, 0, 0, 0, 0, time.UTC)},
		{2022, time.Date(2022, 4, 17, 0, 0, 0, 0, time.UTC)},
		{2000, time.Date(2000, 4, 23, 0, 0, 0, 0, time.UTC)},
	}
	r := Easter()
	for _, tc := range cases {
		assertDate(t, r, tc.year, tc.want)
	}
}

func TestEasterOffset(t *testing.T) {
	// Good Friday is two days before Easter Sunday.
	r := EasterOffset(-2)
	assertDate(t, r, 2024, time.Date(2024, 3, 29, 0, 0, 0, 0, time.UTC))
	// Easter Monday is one day after.
	r = EasterOffset(1)
	assertDate(t, r, 2024, time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC))
}

func TestString(t *testing.T) {
	r := mustFixed(t, 12, 25)
	if got := r.String(); got != "December 25" {
		t.Fatalf("String() = %q, want %q", got, "December 25")
	}
	r = mustNth(t, 11, time.Thursday, 4)
	if got := r.String(); got != "Thursday 4 of November" {
		t.Fatalf("String() = %q, want %q", got, "Thursday 4 of November")
	}
	r = Easter()
	if got := r.String(); got != "Easter" {
		t.Fatalf("String() = %q, want %q", got, "Easter")
	}
	r = EasterOffset(-2)
	if got := r.String(); got != "Easter-2" {
		t.Fatalf("String() = %q, want %q", got, "Easter-2")
	}
}

func TestZeroRuleKind(t *testing.T) {
	// The zero Rule has Kind FixedRule with Month and Day zero, which is
	// invalid. Date must report an error rather than silently returning a
	// bogus date.
	var r Rule
	if _, err := r.Date(2024); err == nil {
		t.Fatal("expected error for zero Rule")
	}
}
