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
	Oneshot         bool
	FailOnInitError bool
	MigStrategy     string
	OutputFilePath  string
	SleepInterval   time.Duration
	Timestamp       bool
}

func (conf *Conf) getConfFromArgv(argv []string) {
	usage := fmt.Sprintf(`%[1]s:
Usage:
  %[1]s [--fail-on-init-error=<bool>] [--mig-strategy=<strategy>] [--oneshot | --sleep-interval=<seconds>] [--timestamp] [--output-file=<file> | -o <file>]
  %[1]s -h | --help
  %[1]s --version

Options:
  -h --help                       Show this help message and exit
  --version                       Display version and exit
  --oneshot                       Label once and exit
  --timestamp                     Add timestamp to the labels
  --fail-on-init-error=<bool>     Fail if there is an error during initialization of any label sources [Default: true]
  --sleep-interval=<seconds>      Time to sleep between labeling [Default: 60s]
  --mig-strategy=<strategy>       Strategy to use for MIG-related labels [Default: none]
  -o <file> --output-file=<file>  Path to output file
                                  [Default: /etc/kubernetes/node-feature-discovery/features.d/gfd]

Arguments:
  <strategy>: none | single | mixed`,
		Bin)

	opts, err := docopt.ParseArgs(usage, argv[1:], Bin+" "+Version)
	if err != nil {
		log.Fatal("Error while parsing command line options: ", err)
	}

	conf.Oneshot, err = opts.Bool("--oneshot")
	if err != nil {
		log.Fatal("Error while parsing command line options: ", err)
	}
	failOnInitErrorTmp, err := opts.String("--fail-on-init-error")
	if err != nil {
		log.Fatal("Error while parsing command line options: ", err)
	}
	if failOnInitErrorTmp != "true" && failOnInitErrorTmp != "false" {
		log.Fatal("Error while parsing command line options: value of --fail-on-init-error must be 'true' or 'false'")
	}
	conf.FailOnInitError = strings.EqualFold(failOnInitErrorTmp, "true")
	conf.MigStrategy, err = opts.String("--mig-strategy")
	if err != nil {
		log.Fatal("Error while parsing command line options: ", err)
	}
	sleepIntervalString, err := opts.String("--sleep-interval")
	if err != nil {
		log.Fatal("Error while parsing command line options: ", err)
	}
	conf.OutputFilePath, err = opts.String("--output-file")
	if err != nil {
		log.Fatal("Error while parsing command line options: ", err)
	}

	conf.SleepInterval, err = time.ParseDuration(sleepIntervalString)
	if err != nil {
		log.Fatal("Invalid value for --sleep-interval option: ", err)
	}

	conf.Timestamp, err = opts.Bool("--timestamp")
	if err != nil {
		log.Fatal("Error while parsing command line options: ", err)
	}

	return
}

func (conf *Conf) getConfFromEnv() {
	val, ok := os.LookupEnv("GFD_ONESHOT")
	if ok && strings.EqualFold(val, "true") {
		conf.Oneshot = true
	}
	val, ok = os.LookupEnv("GFD_TIMESTAMP")
	if ok && strings.EqualFold(val, "true") {
		conf.Timestamp = true
	}
	val, ok = os.LookupEnv("GFD_FAIL_ON_INIT_ERROR")
	if ok && strings.EqualFold(val, "false") {
		conf.FailOnInitError = false
	}
	migStrategyTmp, ok := os.LookupEnv("GFD_MIG_STRATEGY")
	if ok {
		conf.MigStrategy = migStrategyTmp
	}
	sleepIntervalString, ok := os.LookupEnv("GFD_SLEEP_INTERVAL")
	if ok {
		var err error
		conf.SleepInterval, err = time.ParseDuration(sleepIntervalString)
		if err != nil {
			log.Fatal("Invalid value from env for sleep-interval option: ", err)
		}
	}
	outputFilePathTmp, ok := os.LookupEnv("GFD_OUTPUT_FILE")
	if ok {
		conf.OutputFilePath = outputFilePathTmp
	}
}
