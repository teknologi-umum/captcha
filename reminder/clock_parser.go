package reminder

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

func ParseClock(s string) (hour int, minute int, err error) {
	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Split(bufio.ScanRunes)

	var colonMark bool
	var s_hour string
	var s_minute string
	for scanner.Scan() {
		t := scanner.Text()
		if t == ":" {
			colonMark = true
			continue
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

	if hour > 24 {
		err = fmt.Errorf("invalid hour, exceeds 24")
		return
	}

	if hour < 0 {
		err = fmt.Errorf("invalid hour, negative number")
		return
	}

	if minute > 60 {
		err = fmt.Errorf("invalid minute, exceeds 60")
		return
	}

	if minute < 0 {
		err = fmt.Errorf("invalid minute, negative number")
		return
	}

	return
}
