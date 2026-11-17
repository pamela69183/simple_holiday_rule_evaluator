# simple_holiday_rule_evaluator

Computes the date of fixed and floating holidays for a given year, including Western Easter and Nth-weekday rules, with no third-party dependencies.

```go
package main

import (
	"fmt"
	"time"

	shre "simple_holiday_rule_evaluator"
)

func main() {
	christmas, _ := shre.Fixed(12, 25)
	thanksgiving, _ := shre.NthWeekday(11, time.Thursday, 4)
	goodFriday := shre.EasterOffset(-2)

	d, _ := christmas.Date(2024)
	fmt.Println(d.Format("2006-01-02")) // 2024-12-25

	d, _ = thanksgiving.Date(2024)
	fmt.Println(d.Format("2006-01-02")) // 2024-11-28

	d, _ = goodFriday.Date(2024)
	fmt.Println(d.Format("2006-01-02")) // 2024-03-29
}
```

## Why this exists

Holiday calendars are a recurring source of small, annoying bugs: a holiday
is either a fixed date (December 25), a floating rule (the fourth Thursday in
November), or a date relative to Easter (Good Friday). Rather than scatter
that logic through an application, this package keeps the rule as a value and
computes the date on demand. The trade-off is that rules are validated at
construction time and again at evaluation time: a `Rule` value is cheap to
copy, but a year that cannot satisfy the rule (February 29 in a non-leap year,
or a fifth Monday that does not exist) returns an error instead of a
best-effort date. That keeps the failure explicit.

## API

The package exports:

- `Rule` — a value describing a holiday. The zero value is not useful; build
  rules with the constructors below.
- `RuleKind` — `FixedRule`, `FloatingRule`, `EasterRule`.
- `Fixed(month, day int) (Rule, error)` — a fixed month and day.
- `NthWeekday(month int, weekday time.Weekday, n int) (Rule, error)` — the
  Nth weekday of a month. Positive `n` counts from the start; negative `n`
  counts from the end, so `-1` is the last such weekday.
- `LastWeekday(month int, weekday time.Weekday) (Rule, error)` — shorthand for
  `NthWeekday(month, weekday, -1)`.
- `Easter() Rule` — Western Easter Sunday.
- `EasterOffset(days int) Rule` — a date offset from Easter Sunday.
- `(Rule) Date(year int) (time.Time, error)` — the date in UTC at midnight.
- `(Rule) String() string` — a short description such as `December 25` or
  `Thursday 4 of November`.

## The awkward edge

Easter is computed with the anonymous Gregorian algorithm, which is exact for
Gregorian years but does not model the Julian Easter used by some churches.
If you need Julian Easter, this package will give you the wrong date for those
years, and there is no flag to switch algorithms. That is a deliberate scope
decision: the package targets civil calendars, where Western Easter is the
common case.

A second edge: `NthWeekday` accepts `n` in `[-5, 5]` but not every month has a
fifth occurrence of a weekday, and February never has a fifth in a non-leap
year. `Date` returns an error in those cases rather than clamping to the
fourth occurrence.
