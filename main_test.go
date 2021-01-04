// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/NVIDIA/gpu-feature-discovery/pkg/vgpu"
	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
	"github.com/stretchr/testify/require"
)

func NewTestNvmlMock() *NvmlMock {
	one := 1
	model := "MOCKMODEL"
	memory := uint64(128)

	device := nvml.Device{}
	device.Model = &model
	device.Memory = &memory
	device.CudaComputeCapability.Major = &one
	device.CudaComputeCapability.Minor = &one

	return &NvmlMock{
		devices: []NvmlMockDevice{
			NvmlMockDevice{
				instance:   &device,
				attributes: &nvml.DeviceAttributes{},
				migEnabled: false,
				migDevices: []NvmlMockDevice{},
			},
		},
		driverVersion: "400.300",
		cudaMajor:     1,
		cudaMinor:     1,
	}
}

func NewTestVGPUMock(addVGPUMockDevice bool) vgpu.NvidiaMockVGPU {
	mockVGPU := vgpu.NewNvidiaMockVGPU(addVGPUMockDevice)
	return mockVGPU
}

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

	confMigStrategy := Conf{}
	confMigStrategyArgv := []string{Bin, "--mig-strategy=bogus"}
	confMigStrategy.getConfFromArgv(confMigStrategyArgv)

	require.Equal(t, confMigStrategy.MigStrategy, "bogus",
		"MigStrategy option with '--mig-strategy=bogus' argv")

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

	confMigStrategyEnv := Conf{}
	os.Clearenv()
	os.Setenv("GFD_MIG_STRATEGY", "bogus")
	confMigStrategyEnv.getConfFromEnv()

	require.Equal(t, confMigStrategyEnv.MigStrategy, "bogus",
		"MigStrategy option with mig-strategy=bogus env")

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
	nvmlMock := NewTestNvmlMock()
	vgpuMock := NewTestVGPUMock(true)
	conf := Conf{true, "none", "./gfd-test-oneshot", time.Second}

	MachineTypePath = "/tmp/machine-type"
	machineType := []byte("product-name\n")
	err := ioutil.WriteFile("/tmp/machine-type", machineType, 0644)
	require.NoError(t, err, "Write machine type mock file")

	defer func() {
		err = os.Remove(MachineTypePath)
		require.NoError(t, err, "Removing machine type mock file")
	}()

	err = run(nvmlMock, vgpuMock, conf)
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

	err = checkResult(result, "tests/expected-output.txt")
	require.NoError(t, err, "Checking result")
}

func TestRunSleep(t *testing.T) {
	nvmlMock := NewTestNvmlMock()
	vgpuMock := NewTestVGPUMock(true)
	conf := Conf{false, "none", "./gfd-test-loop", 500 * time.Millisecond}

	MachineTypePath = "/tmp/machine-type"
	machineType := []byte("product-name\n")
	err := ioutil.WriteFile("/tmp/machine-type", machineType, 0644)
	require.NoError(t, err, "Write machine type mock file")

	defer func() {
		err = os.Remove(MachineTypePath)
		require.NoError(t, err, "Removing machine type mock file")
	}()

	var runError error
	go func() {
		runError = run(nvmlMock, vgpuMock, conf)
	}()

	// Try to get first timestamp
	outFile, err := waitForFile(conf.OutputFilePath, 5, time.Second)
	require.NoError(t, err, "Open output file while searching for first timestamp")

	output, err := ioutil.ReadAll(outFile)
	require.NoError(t, err, "Read output file while searching for first timestamp")

	err = outFile.Close()
	require.NoError(t, err, "Close output file while searching for first timestamp")

	err = checkResult(output, "tests/expected-output.txt")
	require.NoError(t, err, "Checking result")

	err = os.Remove(conf.OutputFilePath)
	require.NoError(t, err, "Remove output file while searching for first timestamp")

	labels, err := buildLabelMapFromOutput(output)
	require.NoError(t, err, "Building map of labels from output file")
	require.Contains(t, labels, "nvidia.com/gfd.timestamp", "Missing timestamp")

	firstTimestamp := labels["nvidia.com/gfd.timestamp"]

	// Wait for second timestamp
	outFile, err = waitForFile(conf.OutputFilePath, 5, time.Second)
	require.NoError(t, err, "Open output file while searching for second timestamp")

	output, err = ioutil.ReadAll(outFile)
	require.NoError(t, err, "Read output file while searching for second timestamp")

	err = outFile.Close()
	require.NoError(t, err, "Close output file while searching for second timestamp")

	err = checkResult(output, "tests/expected-output.txt")
	require.NoError(t, err, "Checking result")

	err = os.Remove(conf.OutputFilePath)
	require.NoError(t, err, "Remove output file while searching for second timestamp")

	labels, err = buildLabelMapFromOutput(output)
	require.NoError(t, err, "Building map of labels from output file")
	require.Contains(t, labels, "nvidia.com/gfd.timestamp", "Missing timestamp")

	currentTimestamp := labels["nvidia.com/gfd.timestamp"]

	require.NotEqual(t, firstTimestamp, currentTimestamp, "Timestamp didn't change")
	require.NoError(t, runError, "Error from run")
}

func buildLabelMapFromOutput(output []byte) (map[string]string, error) {
	labels := make(map[string]string)

	lines := strings.Split(strings.TrimRight(string(output), "\n"), "\n")
	for _, line := range lines {
		split := strings.Split(line, "=")
		if len(split) != 2 {
			return nil, fmt.Errorf("Unexpected format in line: '%v'", line)
		}
		labels[split[0]] = split[1]
	}

	return labels, nil
}

func checkResult(result []byte, expectedOutputPath string) error {
	expected, err := ioutil.ReadFile(expectedOutputPath)
	if err != nil {
		return fmt.Errorf("Opening expected output file: %v", err)
	}

	var expectedRegexps []*regexp.Regexp
	for _, line := range strings.Split(strings.TrimRight(string(expected), "\n"), "\n") {
		expectedRegexps = append(expectedRegexps, regexp.MustCompile(line))
	}

LOOP:
	for _, line := range strings.Split(strings.TrimRight(string(result), "\n"), "\n") {
		for _, regex := range expectedRegexps {
			if regex.MatchString(line) {
				continue LOOP
			}
		}
		return fmt.Errorf("Line does not match any regexp: %v", string(line))
	}
	return nil
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
