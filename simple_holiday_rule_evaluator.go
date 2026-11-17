// Package simple_holiday_rule_evaluator computes fixed and floating holidays
// for a given year, using only the standard library.
package simple_holiday_rule_evaluator

import (
	"errors"
	"time"
)

// RuleKind distinguishes the two supported ways to describe a holiday.
type RuleKind int

const (
	// FixedRule is a holiday that always falls on the same month and day.
	FixedRule RuleKind = iota
	// FloatingRule is a holiday described by a weekday occurrence in a month.
	FloatingRule
	// EasterRule is a holiday relative to Western (Gregorian) Easter Sunday.
	EasterRule
)

// Rule describes a holiday.
//
// A Rule is a value type; the zero value is not useful because Kind is
// FixedRule with Month and Day both zero. Construct rules with the helpers
// Fixed, NthWeekday, LastWeekday, Easter, or EasterOffset.
type Rule struct {
	Kind RuleKind
	// Month is 1-12 for FixedRule and FloatingRule.
	Month int
	// Day is 1-31 for FixedRule.
	Day int
	// Weekday is the target weekday for FloatingRule.
	Weekday time.Weekday
	// N is the occurrence number for FloatingRule. A positive N means the
	// Nth occurrence from the start of the month; a negative N means the
	// Nth-from-last occurrence (so -1 is the last occurrence).
	N int
	// Offset is the number of days from Easter Sunday for EasterRule.
	// Offset 0 is Easter Sunday itself; negative values are before Easter.
	Offset int
}

// Fixed returns a rule for a holiday on a fixed month and day.
// Month must be in 1..12 and Day must be valid for that month.
func Fixed(month, day int) (Rule, error) {
	if month < 1 || month > 12 {
		return Rule{}, errors.New("month out of range")
	}
	if day < 1 || day > daysInMonth(2000, time.Month(month)) {
		return Rule{}, errors.New("day out of range")
	}
	return Rule{Kind: FixedRule, Month: month, Day: day}, nil
}

// NthWeekday returns a rule for the Nth weekday of the given month.
// A positive n selects from the start of the month (n=1 is the first such
// weekday). A negative n selects from the end (n=-1 is the last such weekday).
func NthWeekday(month int, weekday time.Weekday, n int) (Rule, error) {
	if month < 1 || month > 12 {
		return Rule{}, errors.New("month out of range")
	}
	if weekday < time.Sunday || weekday > time.Saturday {
		return Rule{}, errors.New("weekday out of range")
	}
	if n == 0 {
		return Rule{}, errors.New("n must not be zero")
	}
	if n > 5 || n < -5 {
		return Rule{}, errors.New("n out of range")
	}
	return Rule{Kind: FloatingRule, Month: month, Weekday: weekday, N: n}, nil
}

// LastWeekday returns a rule for the last weekday of the given month.
// It is equivalent to NthWeekday(month, weekday, -1).
func LastWeekday(month int, weekday time.Weekday) (Rule, error) {
	return NthWeekday(month, weekday, -1)
}

// Easter returns a rule for Easter Sunday itself.
func Easter() Rule {
	return Rule{Kind: EasterRule, Offset: 0}
}

// EasterOffset returns a rule for a holiday that is offset days from Easter
// Sunday. Negative offsets are before Easter, positive offsets are after.
func EasterOffset(days int) Rule {
	return Rule{Kind: EasterRule, Offset: days}
}

// Date computes the date of the rule in the given year.
// The returned time.Time is in UTC at midnight, so it is safe to compare and
// format without worrying about the local timezone.
func (r Rule) Date(year int) (time.Time, error) {
	switch r.Kind {
	case FixedRule:
		if r.Month < 1 || r.Month > 12 {
			return time.Time{}, errors.New("month out of range")
		}
		if r.Day < 1 || r.Day > daysInMonth(year, time.Month(r.Month)) {
			return time.Time{}, errors.New("day out of range for month")
		}
		return time.Date(year, time.Month(r.Month), r.Day, 0, 0, 0, 0, time.UTC), nil
	case FloatingRule:
		if r.Month < 1 || r.Month > 12 {
			return time.Time{}, errors.New("month out of range")
		}
		if r.Weekday < time.Sunday || r.Weekday > time.Saturday {
			return time.Time{}, errors.New("weekday out of range")
		}
		if r.N == 0 || r.N > 5 || r.N < -5 {
			return time.Time{}, errors.New("n out of range")
		}
		d, err := nthWeekdayOfMonth(year, time.Month(r.Month), r.Weekday, r.N)
		if err != nil {
			return time.Time{}, err
		}
		return d, nil
	case EasterRule:
		d := easterSunday(year).AddDate(0, 0, r.Offset)
		return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC), nil
	default:
		return time.Time{}, errors.New("unknown rule kind")
	}
}

// String returns a short human-readable description of the rule.
func (r Rule) String() string {
	switch r.Kind {
	case FixedRule:
		return time.Month(r.Month).String() + " " + itoa(r.Day)
	case FloatingRule:
		return weekdayName(r.Weekday) + " " + itoa(r.N) + " of " + time.Month(r.Month).String()
	case EasterRule:
		if r.Offset == 0 {
			return "Easter"
		}
		if r.Offset > 0 {
			return "Easter+" + itoa(r.Offset)
		}
		return "Easter" + itoa(r.Offset)
	default:
		return "unknown"
	}
}

// daysInMonth returns the number of days in the given month of the given year.
// It uses the standard library's time package rather than reimplementing leap
// year logic, which keeps the edge cases (century years, etc.) in one place.
func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// nthWeekdayOfMonth returns the date of the Nth weekday of the month.
// Negative n counts from the end of the month. It returns an error when the
// requested occurrence does not exist (for example a 5th Monday in a month
// that only has four).
func nthWeekdayOfMonth(year int, month time.Month, weekday time.Weekday, n int) (time.Time, error) {
	if n > 0 {
		first := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
		offset := (int(weekday) - int(first.Weekday()) + 7) % 7
		day := 1 + offset + (n-1)*7
		if day > daysInMonth(year, month) {
			return time.Time{}, errors.New("no such occurrence")
		}
		return time.Date(year, month, day, 0, 0, 0, 0, time.UTC), nil
	}
	// Negative n: count backwards from the last day of the month.
	last := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC)
	offset := (int(last.Weekday()) - int(weekday) + 7) % 7
	day := last.Day() - offset + (n+1)*7
	if day < 1 {
		return time.Time{}, errors.New("no such occurrence")
	}
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC), nil
}

// easterSunday computes Western (Gregorian) Easter Sunday for the given year
// using the anonymous Gregorian algorithm. It is valid for all years in the
// Gregorian calendar; the algorithm is well known and avoids any dependency.
func easterSunday(year int) time.Time {
	a := year % 19
	b := year / 100
	c := year % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	month := (h + l - 7*m + 114) / 31
	day := ((h + l - 7*m + 114) % 31) + 1
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

// itoa is a tiny integer-to-string helper used by String so that the package
// does not need to import strconv for a single call site.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// weekdayName returns the English name of a weekday. The standard library
// already has this via time.Weekday.String, but wrapping it keeps String
// readable and avoids surprising output for out-of-range values.
func weekdayName(w time.Weekday) string {
	if w < time.Sunday || w > time.Saturday {
		return "InvalidWeekday"
	}
	return w.String()
}
