package main

import (
	"flag"
	"github.com/j-keck/lsleases/pkg/config"
	"strconv"
	"time"
)

type durationValue time.Duration

func DurationVar(p *time.Duration, name string, usage string) {
	flag.CommandLine.Var((*durationValue)(p), name, usage)
}

func (d *durationValue) Set(str string) error {

	dur, err := parseDuration(str)
	*d = durationValue(dur)

	return err
}

func (d *durationValue) String() string {
	return (*time.Duration)(d).String()
}

type cleanupMethodValue config.CleanupMethod

func CleanupMethodVar(p *config.CleanupMethod, name string, usage string) {
	flag.CommandLine.Var((*cleanupMethodValue)(p), name, usage)
}

func (self *cleanupMethodValue) Set(str string) error {

	if flagIsTrue, err := strconv.ParseBool(str); err != nil {
		return err
	} else {
		if flagIsTrue {
			*self = cleanupMethodValue(config.TimeBasedCleanup)
		} else {
			*self = cleanupMethodValue(config.PingBasedCleanup)
		}
	}
	return nil
}

func (self *cleanupMethodValue) String() string {
	if *self == cleanupMethodValue(config.TimeBasedCleanup) {
		return "time based"
	}
	return "ping based"
}

func (self *cleanupMethodValue) IsBoolFlag() bool {
	return true
}
