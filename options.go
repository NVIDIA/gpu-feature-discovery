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
	Oneshot        bool
	OutputFilePath string
	SleepInterval  time.Duration
}

func (conf *Conf) getConfFromArgv(argv []string) {
	usage := fmt.Sprintf(`%[1]s:
Usage:
  %[1]s [--oneshot | --sleep-interval=<seconds>] [--output-file=<file> | -o <file>]
  %[1]s -h | --help
  %[1]s --version

Options:
  -h --help                       Show this help message and exit
  --version                       Display version and exit
  --oneshot                       Label once and exit
  --sleep-interval=<seconds>      Time to sleep between labeling [Default: 60s]
  -o <file> --output-file=<file>  Path to output file
                                  [Default: /etc/kubernetes/node-feature-discovery/features.d/gfd]`,
		Bin)

	opts, err := docopt.ParseArgs(usage, argv[1:], Bin+" "+Version)
	if err != nil {
		log.Fatal("Error while parsing command line options: ", err)
	}

	conf.Oneshot, err = opts.Bool("--oneshot")
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

	return
}

func (conf *Conf) getConfFromEnv() {
	val, ok := os.LookupEnv("GFD_ONESHOT")
	if ok && strings.EqualFold(val, "true") {
		conf.Oneshot = true
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
