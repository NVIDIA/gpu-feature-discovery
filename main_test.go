// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package main

import (
	"os"
	"testing"
	"time"
	"bytes"
	"strings"
)

func TestGetConfFromArgv(t *testing.T) {

	defaultDuration, _ := time.ParseDuration("60s")

	confNoOptions := Conf{}
	confNoOptionsArgv := []string{ProgName}
	confNoOptions.getConfFromArgv(confNoOptionsArgv)
	if confNoOptions.Oneshot != false {
		t.Error("Oneshot option with empty argv: got true, expected false")
	}
	if confNoOptions.SleepInterval != defaultDuration {
		t.Errorf("SleepInterval option with empty argv: got %s, expected %s",
			confNoOptions.SleepInterval, defaultDuration)
	}

	confOneShot := Conf{}
	confOneShotArgv := []string{ProgName, "--oneshot"}
	confOneShot.getConfFromArgv(confOneShotArgv)
	if confOneShot.Oneshot != true {
		t.Error("Oneshot option with '--oneshot' argv: got false, expected true")
	}
	if confOneShot.SleepInterval != defaultDuration {
		t.Errorf("SleepInterval option with '--oneshot' argv: got %s, expected %s",
			confOneShot.SleepInterval, defaultDuration)
	}

	confSleepInterval := Conf{}
	confSleepIntervalArgv := []string{ProgName, "--sleep-interval=10s"}
	confSleepInterval.getConfFromArgv(confSleepIntervalArgv)
	if confSleepInterval.Oneshot != false {
		t.Error("Oneshot option with '--sleep-interval=10s' argv: got true, expected false")
	}
	if duration, _ := time.ParseDuration("10s"); confSleepInterval.SleepInterval != duration {
		t.Errorf("SleepInterval option with '--sleep-interval=10s' argv: got %s, expected %s",
			confSleepInterval.SleepInterval, duration)
	}
}

func TestGetConfFromEnv(t *testing.T) {

	defaultDuration, _ := time.ParseDuration("0s")

	confNoEnv := Conf{}
	confNoEnv.getConfFromEnv()
	if confNoEnv.Oneshot != false {
		t.Error("Oneshot option with empty env: got true, expected false")
	}
	if confNoEnv.SleepInterval != defaultDuration {
		t.Errorf("SleepInterval option with empty env: got %s, expected %s",
			confNoEnv.SleepInterval, defaultDuration)
	}

	confOneShotEnv := Conf{}
	os.Clearenv()
	os.Setenv("NVIDIA_FEATURE_DISCOVERY_ONESHOT", "TrUe")
	confOneShotEnv.getConfFromEnv()
	if confOneShotEnv.Oneshot != true {
		t.Error("Oneshot option with oneshot env: got false, expected true")
	}
	if confOneShotEnv.SleepInterval != defaultDuration {
		t.Errorf("SleepInterval option with oneshot env: got %s, expected %s",
			confOneShotEnv.SleepInterval, defaultDuration)
	}

	confSleepIntervalEnv := Conf{}
	duration, _ := time.ParseDuration("10s")
	os.Clearenv()
	os.Setenv("NVIDIA_FEATURE_DISCOVERY_SLEEP_INTERVAL", "10s")
	confSleepIntervalEnv.getConfFromEnv()
	if confSleepIntervalEnv.Oneshot != false {
		t.Error("Oneshot option with sleep-interval=10s env: got true, expected false")
	}
	if confSleepIntervalEnv.SleepInterval != duration {
		t.Errorf("SleepInterval option with sleep-interval=10s env: got %s, expected %s",
			confSleepIntervalEnv.SleepInterval, defaultDuration)
	}
}

func TestRun(t *testing.T) {
	nvmlMock := NvmlMock{}
	duration, _ := time.ParseDuration("10s")
	conf := Conf{true, duration}

	expected := `nvidia-driver-version=MOCK-DRIVER-VERSION
nvidia-model=MOCK-MODEL
nvidia-memory=128
`

	buf := new(bytes.Buffer)
	run(nvmlMock, conf, buf)

	if strings.Compare(expected, buf.String()) != 0 {
		t.Errorf("Output mismatch: expected '%s', got '%s'", expected, buf.String())
	}
}
