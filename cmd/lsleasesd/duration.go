package main

import (
	"regexp"
	"strconv"
	"time"
)

//
// parseDuration is like time.ParseDuration, but it also accepts days with the 'd' suffix
//
// https://github.com/golang/go/issues/11473
// > Some days are only 23 hours long, some are 25. Most are 24, but not today. I think it's fine to stop at hours,
// > which are often 1/24 of a day. Otherwise where do you stop? Week? Month? Year? Decade? Century? Millennium? Era?
// > We must stop somewhere, and for computer usage hour seems like a fine drawing point since real ambiguity sets in
// > at the next level. (The one second of a minute around a leap second is not a real issue here.)
//
// Not every day has 24h, but this is perfectly fine for my use-case
//
func parseDuration(s string) (time.Duration, error) {
	r, _ := regexp.Compile("^(\\d+)d(.*)?$")

	if r.MatchString(s) {
		groups := r.FindStringSubmatch(s)

		days, err := strconv.Atoi(groups[1])
		if err != nil {
			return 0, err
		}

		var dur time.Duration
		if len(groups[2]) > 0 {
			dur, err = time.ParseDuration(groups[2])
			if err != nil {
				return 0, err
			}
		}

		var nanosPerDay int64 = 86400000000000
		return time.Duration(int64(days)*nanosPerDay + int64(dur)), nil
	} else {
		return time.ParseDuration(s)
	}
}
