// fmtd.go
package fmtd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	rx *regexp.Regexp
)

/*
const (
	year int = 24 * 256  // really??!
	week int = 24 * 7
)
*/
func init() {
	rx, _ = regexp.Compile(`([0-9\.]+(ms|s|m|h))`)
}

//410h24m35.764062385s //987ms

type Fmtd struct {
	Days      int
	Hours     int
	Minutes   int
	Seconds   int
	Millisecs int
	Err       error
}

func Duration(d time.Duration) (f Fmtd) {
	split := rx.FindAllString(d.String(), -1)
	for _, s := range split {
		if s[len(s)-2:] == "ms" {
			f.Millisecs, f.Err = strconv.Atoi(s[:len(s)-2])
			break
		} else {
			switch s[len(s)-1:] {
			case "h":
				var i int
				i, f.Err = strconv.Atoi(s[:len(s)-1])
				if i >= 24 {
					f.Days = i / 24
					f.Hours = int(((float32(i) / 24) - float32(f.Days)) * 24)
				} else {
					f.Hours = i
				}
			case "m":
				f.Minutes, f.Err = strconv.Atoi(s[:len(s)-1])
			case "s":
				i := strings.Index(s, ".")
				if i == -1 {
					f.Seconds, f.Err = strconv.Atoi(s[:len(s)-1])
				} else {
					f.Seconds, f.Err = strconv.Atoi(s[:i])
					if len(s)-i-2 < 4 {
						f.Millisecs, f.Err = strconv.Atoi(s[i+1 : i+len(s)-i-1])
					} else {
						f.Millisecs, f.Err = strconv.Atoi(s[i+1 : i+5])
					}
					//fmt.Println(s, len(s), i, len(s)-i-2, f.Millisecs)
				}

			}
		}
	}
	return
}

func Largest(d time.Duration) string {
	return rx.FindString(d.String())
}

const (
	SECOND   int64 = 1000000000
	JUST_NOW int64 = SECOND * 10
	MINUTE   int64 = SECOND * 60
	HOUR     int64 = MINUTE * 60
	DAY      int64 = HOUR * 24
	WEEK     int64 = DAY * 7
	YEAR     int64 = DAY * 365
)

func Fuzzy(d time.Duration) string {
	ns := d.Nanoseconds()
	switch {
	case ns < JUST_NOW:
		return "Just now"
	case ns < MINUTE:
		i := ns / SECOND
		return fmt.Sprintf("%d seconds ago", i)
	case ns < HOUR:
		i := ns / MINUTE
		return fmt.Sprintf("%d minute%s ago", i, plural(i))
	case ns < DAY:
		i := ns / HOUR
		return fmt.Sprintf("%d hour%s ago", i, plural(i))
	case ns < WEEK:
		i := ns / DAY
		return fmt.Sprintf("%d day%s ago", i, plural(i))
	case ns < YEAR:
		i := ns / WEEK
		return fmt.Sprintf("%d week%s ago", i, plural(i))
	default:
		i := ns / YEAR
		return fmt.Sprintf("%d year%s ago", i, plural(i))
	}
}

func FuzzySmall(d time.Duration) string {
	ns := d.Nanoseconds()
	switch {
	case ns < JUST_NOW:
		return "Just now"
	case ns < MINUTE:
		i := ns / SECOND
		return fmt.Sprintf("%d secs", i)
	case ns < HOUR:
		i := ns / MINUTE
		return fmt.Sprintf("%d min%s ago", i, plural(i))
	case ns < DAY:
		i := ns / HOUR
		return fmt.Sprintf("%d hr%s ago", i, plural(i))
	case ns < WEEK:
		i := ns / DAY
		return fmt.Sprintf("%d day%s ago", i, plural(i))
	case ns < YEAR:
		i := ns / WEEK
		return fmt.Sprintf("%d wk%s ago", i, plural(i))
	default:
		i := ns / YEAR
		return fmt.Sprintf("%d yr%s ago", i, plural(i))
	}
}

/*
func Fuzzy(d time.Duration) string {
	// TODO: rewrite this without regex (eg if date > nnn == week)
	s := rx.FindString(d.String())
	if s[len(s)-2:] == "ms" {
		return "Just now"
	} else {
		//r := s[:len(s)-1]
		switch s[len(s)-1:] {
		case "h":
			i, _ := strconv.Atoi(s[:len(s)-1])
			if i >= year {
				years := i / year
				return fmt.Sprintf("%d year%s ago", years, plural(years))
			}
			if i >= week {
				weeks := i / week
				return fmt.Sprintf("%d week%s ago", weeks, plural(weeks))
			}
			if i >= 24 {
				days := i / 24
				return fmt.Sprintf("%d day%s ago", days, plural(days))
			} else {
				return fmt.Sprintf("%d hour%s ago", i, plural(i))
			}
		case "m":
			i, _ := strconv.Atoi(s[:len(s)-1])
			return fmt.Sprintf("%s minunte%s ago", s[:len(s)-1], plural(i))
		case "s":
			var secs int
			i := strings.Index(s, ".")
			if i == -1 {
				secs, _ = strconv.Atoi(s[:len(s)-1])
			} else {
				secs, _ = strconv.Atoi(s[:i])
			}
			return fmt.Sprintf("%d second%s ago", secs, plural(secs))
		}
	}
	return "[func Fuzzy(d time.Duration) string {failed}]"
}
*/
func plural(i int64) string {
	if i == 1 {
		return ""
	}
	return "s"
}
