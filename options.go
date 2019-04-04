// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/docopt/docopt-go"
)

// Conf : Type to represent options
type Conf struct {
	Oneshot       bool
	SleepInterval time.Duration
}

func (conf *Conf) getConfFromArgv(argv []string) {

	usage := fmt.Sprintf(`%s:
Usage:
  %s [--oneshot | --sleep-interval=<seconds>]
  %s -h | --help
  %s --version

Options:
  -h --help                   Show this help message and exit
  --version                   Display version and exit
  --oneshot                   Label once and exit
  --sleep-interval=<seconds>  Time to sleep between labeling [Default: 60s]`,
		bin, bin, bin, bin)

	opts, _ := docopt.ParseArgs(usage, argv[1:], bin + " " + version)

	var err error
	conf.Oneshot, err = opts.Bool("--oneshot")
	sleepIntervalString, err := opts.String("--sleep-interval")
	if err != nil {
		log.Fatal("Error while parsing command line options: ", err)
	}

	conf.SleepInterval, err = time.ParseDuration(sleepIntervalString)
	if err != nil {
		log.Fatal("Invalid value for --sleep-interval option: ", err)
	}

	return
}

func (conf *Conf) getConfFromEnv() {
	// TODO: Change env vars name
	val, ok := os.LookupEnv("NVIDIA_FEATURE_DISCOVERY_ONESHOT")
	if ok && strings.EqualFold(val, "true") {
		conf.Oneshot = true
	}
	sleepIntervalString, ok := os.LookupEnv("NVIDIA_FEATURE_DISCOVERY_SLEEP_INTERVAL")
	if ok {
		var err error
		conf.SleepInterval, err = time.ParseDuration(sleepIntervalString)
		if err != nil {
			log.Fatal("Invalid value from env for sleep-interval option: ", err)
		}
	}
}
