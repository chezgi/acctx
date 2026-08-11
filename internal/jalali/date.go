package jalali

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var datePattern = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)

type Date struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

func Parse(value string) (Date, error) {
	if !datePattern.MatchString(value) {
		return Date{}, fmt.Errorf("invalid Jalali date %q", value)
	}
	parts := strings.Split(value, "-")
	year, _ := strconv.Atoi(parts[0])
	month, _ := strconv.Atoi(parts[1])
	day, _ := strconv.Atoi(parts[2])
	date := Date{Year: year, Month: month, Day: day}
	if err := date.Validate(); err != nil {
		return Date{}, err
	}
	return date, nil
}

func (date Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", date.Year, date.Month, date.Day)
}

func (date Date) Validate() error {
	if date.Year < -60 || date.Year >= 3178 || date.Month < 1 || date.Month > 12 {
		return fmt.Errorf("invalid Jalali date %s", date.String())
	}
	if date.Day < 1 || date.Day > DaysInMonth(date.Year, date.Month) {
		return fmt.Errorf("invalid Jalali date %s", date.String())
	}
	return nil
}

func IsLeap(year int) bool {
	breaks := []int{-61, 9, 38, 199, 426, 686, 756, 818, 1111, 1181, 1210, 1635, 2060, 2097, 2192, 2262, 2324, 2394, 2456, 3178}
	previous := breaks[0]
	jump := 0
	for _, current := range breaks[1:] {
		jump = current - previous
		if year < current {
			break
		}
		previous = current
	}
	n := year - previous
	if jump-n < 6 {
		n = n - jump + ((jump+4)/33)*33
	}
	leap := ((n+1)%33 - 1) % 4
	if leap == -1 {
		leap = 4
	}
	return leap == 0
}

func DaysInMonth(year, month int) int {
	if month >= 1 && month <= 6 {
		return 31
	}
	if month >= 7 && month <= 11 {
		return 30
	}
	if month == 12 && IsLeap(year) {
		return 30
	}
	if month == 12 {
		return 29
	}
	return 0
}

func (date Date) AddDays(days int) (Date, error) {
	if days < 0 {
		return Date{}, fmt.Errorf("negative day offsets are not supported")
	}
	if err := date.Validate(); err != nil {
		return Date{}, err
	}
	result := date
	for step := 0; step < days; step++ {
		result.Day++
		if result.Day > DaysInMonth(result.Year, result.Month) {
			result.Day = 1
			result.Month++
			if result.Month > 12 {
				result.Month = 1
				result.Year++
			}
		}
	}
	return result, nil
}

func (date Date) AddMonths(months int) (Date, error) {
	if months < 0 {
		return Date{}, fmt.Errorf("negative month offsets are not supported")
	}
	if err := date.Validate(); err != nil {
		return Date{}, err
	}
	total := date.Year*12 + date.Month - 1 + months
	result := Date{Year: total / 12, Month: total%12 + 1, Day: date.Day}
	lastDay := DaysInMonth(result.Year, result.Month)
	if result.Day > lastDay {
		result.Day = lastDay
	}
	return result, nil
}

func (date Date) EndOfMonthAfter(months int) (Date, error) {
	result, err := date.AddMonths(months)
	if err != nil {
		return Date{}, err
	}
	result.Day = DaysInMonth(result.Year, result.Month)
	return result, nil
}
