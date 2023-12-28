package reminder

import (
	"bufio"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrParseClock = errors.New("parse clock")

func ParseClock(s string) (hour int, minute int, err error) {
	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Split(bufio.ScanRunes)

	var colonMark bool
	var s_hour string
	var s_minute string
	for scanner.Scan() {
		t := scanner.Text()
		if t == ":" {
			if !colonMark {
				colonMark = true
				continue
			}

			break
		}

		if !colonMark {
			s_hour += t
		} else {
			s_minute += t
		}
	}

	hour, err = strconv.Atoi(s_hour)
	if err != nil {
		return
	}

	minute, err = strconv.Atoi(s_minute)
	if err != nil {
		return
	}

	if hour >= 24 {
		err = fmt.Errorf("%w: invalid hour, exceeds 24", ErrParseClock)
		return
	}

	if hour < 0 {
		err = fmt.Errorf("%w: invalid hour, negative number", ErrParseClock)
		return
	}

	if minute >= 60 {
		err = fmt.Errorf("%w: invalid minute, exceeds 60", ErrParseClock)
		return
	}

	if minute < 0 {
		err = fmt.Errorf("%w: invalid minute, negative number", ErrParseClock)
		return
	}

	return
}
