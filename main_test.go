// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestGetConfFromArgv(t *testing.T) {

	defaultDuration := time.Second * 60

	confNoOptions := Conf{}
	confNoOptionsArgv := []string{Bin}
	confNoOptions.getConfFromArgv(confNoOptionsArgv)
	if confNoOptions.Oneshot != false {
		t.Error("Oneshot option with empty argv: got true, expected false")
	}
	if confNoOptions.SleepInterval != defaultDuration {
		t.Errorf("SleepInterval option with empty argv: got %s, expected %s",
			confNoOptions.SleepInterval, defaultDuration)
	}

	confOneShot := Conf{}
	confOneShotArgv := []string{Bin, "--oneshot"}
	confOneShot.getConfFromArgv(confOneShotArgv)
	if confOneShot.Oneshot != true {
		t.Error("Oneshot option with '--oneshot' argv: got false, expected true")
	}
	if confOneShot.SleepInterval != defaultDuration {
		t.Errorf("SleepInterval option with '--oneshot' argv: got %s, expected %s",
			confOneShot.SleepInterval, defaultDuration)
	}

	confSleepInterval := Conf{}
	confSleepIntervalArgv := []string{Bin, "--sleep-interval=1s"}
	confSleepInterval.getConfFromArgv(confSleepIntervalArgv)
	if confSleepInterval.Oneshot != false {
		t.Error("Oneshot option with '--sleep-interval=1s' argv: got true, expected false")
	}
	if confSleepInterval.SleepInterval != time.Second {
		t.Errorf("SleepInterval option with '--sleep-interval=1s' argv: got %s, expected %s",
			confSleepInterval.SleepInterval, time.Second)
	}

	confOutputFile := Conf{}
	confOutputFileArgv := []string{Bin, "--output-file=test"}
	confOutputFile.getConfFromArgv(confOutputFileArgv)
	if confOutputFile.OutputFilePath != "test" {
		t.Errorf("OutputFilePath option with '--output-file=test' argv: got %s, expected %s",
			confOutputFile.OutputFilePath, "test")
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
	os.Setenv("GFD_ONESHOT", "TrUe")
	confOneShotEnv.getConfFromEnv()
	if confOneShotEnv.Oneshot != true {
		t.Error("Oneshot option with oneshot env: got false, expected true")
	}
	if confOneShotEnv.SleepInterval != defaultDuration {
		t.Errorf("SleepInterval option with oneshot env: got %s, expected %s",
			confOneShotEnv.SleepInterval, defaultDuration)
	}

	confSleepIntervalEnv := Conf{}
	os.Clearenv()
	os.Setenv("GFD_SLEEP_INTERVAL", "1s")
	confSleepIntervalEnv.getConfFromEnv()
	if confSleepIntervalEnv.Oneshot != false {
		t.Error("Oneshot option with sleep-interval=1s env: got true, expected false")
	}
	if confSleepIntervalEnv.SleepInterval != time.Second {
		t.Errorf("SleepInterval option with sleep-interval=1s env: got %s, expected %s",
			confSleepIntervalEnv.SleepInterval, defaultDuration)
	}

	confOutputFileEnv := Conf{}
	os.Clearenv()
	os.Setenv("GFD_OUTPUT_FILE", "test")
	confOutputFileEnv.getConfFromEnv()
	if confOutputFileEnv.OutputFilePath != "test" {
		t.Errorf("OutputFilePath option with output-file=test env: got %s, expected %s",
			confOutputFileEnv.OutputFilePath, "test")
	}
}

func TestRunOneshot(t *testing.T) {
	nvmlMock := NvmlMock{}
	conf := Conf{true, "./gfd-test-oneshot", time.Second}

	expected, _ := regexp.Compile(`nvidia-timestamp=[0-9]{10}
nvidia-driver-version=MOCK-DRIVER-VERSION
nvidia-model=MOCK-MODEL
nvidia-memory=128
`)

	run(nvmlMock, conf)

	outFile, err := os.Open(conf.OutputFilePath)
	if err != nil {
		t.Fatalf("Error opening output file: %v", err)
	}
	defer func () {
		err = outFile.Close()
		if err != nil {
			t.Logf("Error closing output file '%s': %v", conf.OutputFilePath, err)
		}
		err = os.Remove(conf.OutputFilePath)
		if err != nil {
			t.Logf("Error removing output file '%s': %v", conf.OutputFilePath, err)
		}
	}()

	result, err := ioutil.ReadAll(outFile)
	if err != nil {
		t.Fatalf("Error reading output file: %v", err)
	}

	if !expected.Match(result) {
		t.Errorf("Output mismatch: expected '%s', got '%s'", expected, result)
	}
}

func waitForFile(fileName string, iter int, sleepInterval time.Duration) (*os.File, error) {
	for i := 0; i < iter - 1; i++ {
		file, err := os.Open(fileName)
		if err != nil && os.IsNotExist(err) {
			time.Sleep(sleepInterval)
			continue
		}
		if err != nil {
			return nil, err
		}
		return file, nil
	}
	return os.Open(fileName)
}

func TestRunSleep(t *testing.T) {
	nvmlMock := NvmlMock{}
	conf := Conf{false, "./gfd-test-loop", time.Second}
	expected, _ := regexp.Compile(`nvidia-timestamp=[0-9]{10}
nvidia-driver-version=MOCK-DRIVER-VERSION
nvidia-model=MOCK-MODEL
nvidia-memory=128
`)

	go run(nvmlMock, conf)

	// Try to get first timestamp
	outFile, err := waitForFile(conf.OutputFilePath, 5, time.Second)
	if err != nil {
		t.Fatalf("Failed to open output file while searching for first timestamp: %v", err)
	}

	output, err := ioutil.ReadAll(outFile)
	if err != nil {
		t.Fatalf("Failed to read output file while searching for first timestamp: %v", err)
	}

	err = outFile.Close()
	if err != nil {
		t.Fatalf("Failed to close output file while searching for first timestamp: %v", err)
	}

	err = os.Remove(conf.OutputFilePath)
	if err != nil {
		t.Fatalf("Failed to remove output file while searching for first timestamp: %v", err)
	}

	timestampLabel := string(bytes.Split(output, []byte("\n"))[0])

	if !strings.Contains(timestampLabel, "=") {
		t.Fatal("Invalid timestamp label format")
	}

	firstTimestamp := strings.Split(timestampLabel, "=")[1]

	// Wait for second timestamp
	outFile, err = waitForFile(conf.OutputFilePath, 5, time.Second)
	if err != nil {
		t.Fatalf("Failed to open output file while searching for second timestamp: %v", err)
	}

	output, err = ioutil.ReadAll(outFile)
	if err != nil {
		t.Fatalf("Failed to read output file while searching for second timestamp: %v", err)
	}

	err = outFile.Close()
	if err != nil {
		t.Fatalf("Failed to close output file while searching for second timestamp: %v", err)
	}

	timestampLabel = string(bytes.Split(output, []byte("\n"))[0])

	if !strings.Contains(timestampLabel, "=") {
		t.Fatal("Invalid timestamp label format")
	}

	currentTimestamp := strings.Split(timestampLabel, "=")[1]

	if firstTimestamp == currentTimestamp {
		t.Fatalf("Timestamp didn't change")
	}

	if !expected.Match(output) {
		t.Errorf("Output mismatch: expected '%s', got '%s'", expected, output)
	}
}
