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

	"github.com/stretchr/testify/require"
)

func TestGetConfFromArgv(t *testing.T) {

	defaultDuration := time.Second * 60

	confNoOptions := Conf{}
	confNoOptionsArgv := []string{Bin}
	confNoOptions.getConfFromArgv(confNoOptionsArgv)

	require.False(t, confNoOptions.Oneshot, "Oneshot option with empty argv")
	require.Equal(t, confNoOptions.SleepInterval, defaultDuration,
		"SleepInterval option with empty argv")

	confOneShot := Conf{}
	confOneShotArgv := []string{Bin, "--oneshot"}
	confOneShot.getConfFromArgv(confOneShotArgv)

	require.True(t, confOneShot.Oneshot, "Oneshot option with '--oneshot' argv")
	require.Equal(t, confOneShot.SleepInterval, defaultDuration,
		"SleepInterval option with '--oneshot' argv")

	confSleepInterval := Conf{}
	confSleepIntervalArgv := []string{Bin, "--sleep-interval=1s"}
	confSleepInterval.getConfFromArgv(confSleepIntervalArgv)

	require.False(t, confSleepInterval.Oneshot,
		"Oneshot option with '--sleep-interval=1s' argv")
	require.Equal(t, confSleepInterval.SleepInterval, time.Second,
		"SleepInterval option with '--sleep-interval=1s' argv")

	confOutputFile := Conf{}
	confOutputFileArgv := []string{Bin, "--output-file=test"}
	confOutputFile.getConfFromArgv(confOutputFileArgv)

	require.Equal(t, confOutputFile.OutputFilePath, "test",
		"OutputFilePath option with '--output-file=test' argv")
}

func TestGetConfFromEnv(t *testing.T) {

	defaultDuration, _ := time.ParseDuration("0s")

	confNoEnv := Conf{}
	confNoEnv.getConfFromEnv()

	require.False(t, confNoEnv.Oneshot, "Oneshot option with empty env")
	require.Equal(t, confNoEnv.SleepInterval, defaultDuration,
		"SleepInterval option with empty env")

	confOneShotEnv := Conf{}
	os.Clearenv()
	os.Setenv("GFD_ONESHOT", "TrUe")
	confOneShotEnv.getConfFromEnv()

	require.True(t, confOneShotEnv.Oneshot, "Oneshot option with oneshot env")
	require.Equal(t, confOneShotEnv.SleepInterval, defaultDuration,
		"SleepInterval option with oneshot env")

	confSleepIntervalEnv := Conf{}
	os.Clearenv()
	os.Setenv("GFD_SLEEP_INTERVAL", "1s")
	confSleepIntervalEnv.getConfFromEnv()

	require.False(t, confSleepIntervalEnv.Oneshot,
		"Oneshot option with sleep-interval=1s env")
	require.Equal(t, confSleepIntervalEnv.SleepInterval, time.Second,
		"SleepInterval option with sleep-interval=1s env")

	confOutputFileEnv := Conf{}
	os.Clearenv()
	os.Setenv("GFD_OUTPUT_FILE", "test")
	confOutputFileEnv.getConfFromEnv()

	require.Equal(t, confOutputFileEnv.OutputFilePath, "test",
		"OutputFilePath option with output-file=test env")
}

func TestRunOneshot(t *testing.T) {
	nvmlMock := NvmlMock{}
	conf := Conf{true, "./gfd-test-oneshot", time.Second}

	expected, err := ioutil.ReadFile("tests/expected-output.txt")
	require.NoError(t, err, "Opening expected output file")

	expectedRegexp := regexp.MustCompile(string(expected))

	err = run(nvmlMock, conf)
	require.NoError(t, err, "Error from run function")

	outFile, err := os.Open(conf.OutputFilePath)
	require.NoError(t, err, "Opening output file")

	defer func() {
		err = outFile.Close()
		require.NoError(t, err, "Closing output file")
		err = os.Remove(conf.OutputFilePath)
		require.NoError(t, err, "Removing output file")
	}()

	result, err := ioutil.ReadAll(outFile)
	require.NoError(t, err, "Reading output file")
	require.Regexp(t, expectedRegexp, string(result), "Output mismatch")
}

func waitForFile(fileName string, iter int, sleepInterval time.Duration) (*os.File, error) {
	for i := 0; i < iter-1; i++ {
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

	expected, err := ioutil.ReadFile("tests/expected-output.txt")
	require.NoError(t, err, "Opening expected output file")

	expectedRegexp := regexp.MustCompile(string(expected))

	var runError error
	go func() {
		runError = run(nvmlMock, conf)
	}()

	// Try to get first timestamp
	outFile, err := waitForFile(conf.OutputFilePath, 5, time.Second)
	require.NoError(t, err, "Open output file while searching for first timestamp")

	output, err := ioutil.ReadAll(outFile)
	require.NoError(t, err, "Read output file while searching for first timestamp")

	err = outFile.Close()
	require.NoError(t, err, "Close output file while searching for first timestamp")

	err = os.Remove(conf.OutputFilePath)
	require.NoError(t, err, "Remove output file while searching for first timestamp")

	timestampLabel := string(bytes.Split(output, []byte("\n"))[0])
	require.Contains(t, timestampLabel, "=", "Invalid timestamp label format")

	firstTimestamp := strings.Split(timestampLabel, "=")[1]

	// Wait for second timestamp
	outFile, err = waitForFile(conf.OutputFilePath, 5, time.Second)
	require.NoError(t, err, "Open output file while searching for second timestamp")

	output, err = ioutil.ReadAll(outFile)
	require.NoError(t, err, "Read output file while searching for second timestamp")

	err = outFile.Close()
	require.NoError(t, err, "Close output file while searching for second timestamp")

	timestampLabel = string(bytes.Split(output, []byte("\n"))[0])
	require.Contains(t, timestampLabel, "=", "Invalid timestamp label format")

	currentTimestamp := strings.Split(timestampLabel, "=")[1]

	require.NotEqual(t, firstTimestamp, currentTimestamp, "Timestamp didn't change")
	require.Regexp(t, expectedRegexp, string(output), "Output mismatch")
	require.NoError(t, runError, "Error from run")
}
