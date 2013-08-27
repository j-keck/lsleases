package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"
)

func exitOnError(err error, msg ...string) {
	if err != nil {
		log.Println(msg, ": ", err)
		os.Exit(1)
	}
}

func panicOnError(err error, msg ...string) {
	if err != nil {
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func parseDuration(str string) (time.Duration, error) {
	r, err := regexp.Compile("^(\\d+)d(.*)?$")
	if err != nil {
		return 0, err
	}

	if r.MatchString(str) {
		found := r.FindStringSubmatch(str)
		daysStr := found[1]
		restStr := found[2]

		days, err := strconv.Atoi(daysStr)
		if err != nil {
			return 0, err
		}

		daysDuration, err := time.ParseDuration(fmt.Sprintf("%dh", days*24))
		if err != nil {
			return 0, err
		}

		var restDuration time.Duration
		if len(restStr) > 0 {
			restDuration, err = time.ParseDuration(restStr)
			if err != nil {
				return 0, err
			}
		}

		return time.Duration(int64(daysDuration) + int64(restDuration)), nil
	}
	return time.ParseDuration(str)
}
