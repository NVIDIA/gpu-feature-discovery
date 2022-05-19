// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/NVIDIA/gpu-feature-discovery/internal/nvml"
	"github.com/NVIDIA/gpu-feature-discovery/internal/vgpu"
	config "github.com/NVIDIA/k8s-device-plugin/api/config/v1"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	root string
}

var cfg *testConfig

func TestMain(m *testing.M) {
	// TEST SETUP
	// Determine the module root and the test binary path
	var err error
	moduleRoot, err := getModuleRoot()
	if err != nil {
		log.Printf("error in test setup: could not get module root: %v", err)
		os.Exit(1)
	}

	// Store the root and binary paths in the test Config
	cfg = &testConfig{
		root: moduleRoot,
	}

	// RUN TESTS
	exitCode := m.Run()

	os.Exit(exitCode)
}

func getModuleRoot() (string, error) {
	_, filename, _, _ := runtime.Caller(0)

	return hasGoMod(filename)
}

func hasGoMod(dir string) (string, error) {
	if dir == "" || dir == "/" {
		return "", fmt.Errorf("module root not found")
	}

	_, err := os.Stat(filepath.Join(dir, "go.mod"))
	if err != nil {
		return hasGoMod(filepath.Dir(dir))
	}
	return dir, nil
}

func (t testConfig) Path(path string) string {
	return filepath.Join(t.root, path)
}

func NewTestNvmlMock() *nvml.NvmlMock {
	device := nvml.NvmlMockDevice{
		Model:        "MOCKMODEL",
		ComputeMajor: 1,
		ComputeMinor: 1,
		TotalMemory:  uint64(128),
	}

	return &nvml.NvmlMock{
		Devices: []nvml.NvmlMockDevice{
			device,
		},
		DriverVersion: "400.300",
		CudaMajor:     1,
		CudaMinor:     1,
	}
}

func NewTestVGPUMock() vgpu.Interface {
	return vgpu.NewMockVGPU()
}

func TestRunOneshot(t *testing.T) {
	nvmlMock := NewTestNvmlMock()
	vgpuMock := NewTestVGPUMock()
	conf := &config.Config{
		Flags: config.Flags{
			CommandLineFlags: config.CommandLineFlags{
				MigStrategy:     "none",
				FailOnInitError: true,
				GFD: config.GFDCommandLineFlags{
					Oneshot:       true,
					OutputFile:    "./gfd-test-oneshot",
					SleepInterval: time.Second,
					NoTimestamp:   false,
				},
			},
		},
	}

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

	outFile, err := os.Open(conf.Flags.GFD.OutputFile)
	require.NoError(t, err, "Opening output file")

	defer func() {
		err = outFile.Close()
		require.NoError(t, err, "Closing output file")
		err = os.Remove(conf.Flags.GFD.OutputFile)
		require.NoError(t, err, "Removing output file")
	}()

	result, err := ioutil.ReadAll(outFile)
	require.NoError(t, err, "Reading output file")

	err = checkResult(result, cfg.Path("tests/expected-output.txt"), false)
	require.NoError(t, err, "Checking result")

	err = checkResult(result, cfg.Path("tests/expected-output-vgpu.txt"), true)
	require.NoError(t, err, "Checking result for vgpu labels")
}

func TestRunWithNoTimestamp(t *testing.T) {
	nvmlMock := NewTestNvmlMock()
	vgpuMock := NewTestVGPUMock()
	conf := &config.Config{
		Flags: config.Flags{
			CommandLineFlags: config.CommandLineFlags{
				MigStrategy:     "none",
				FailOnInitError: true,
				GFD: config.GFDCommandLineFlags{
					Oneshot:       true,
					OutputFile:    "./gfd-test-with-no-timestamp",
					SleepInterval: time.Second,
					NoTimestamp:   true,
				},
			},
		},
	}

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

	outFile, err := os.Open(conf.Flags.GFD.OutputFile)
	require.NoError(t, err, "Opening output file")

	defer func() {
		err = outFile.Close()
		require.NoError(t, err, "Closing output file")
		err = os.Remove(conf.Flags.GFD.OutputFile)
		require.NoError(t, err, "Removing output file")
	}()

	result, err := ioutil.ReadAll(outFile)
	require.NoError(t, err, "Reading output file")

	err = checkResult(result, cfg.Path("tests/expected-output.txt"), false)
	require.NoError(t, err, "Checking result")
	require.NotContains(t, string(result), "nvidia.com/gfd.timestamp=", "Checking absent timestamp")

	err = checkResult(result, cfg.Path("tests/expected-output-vgpu.txt"), true)
	require.NoError(t, err, "Checking result for vgpu labels")
}

func TestRunSleep(t *testing.T) {
	nvmlMock := NewTestNvmlMock()
	vgpuMock := NewTestVGPUMock()
	conf := &config.Config{
		Flags: config.Flags{
			CommandLineFlags: config.CommandLineFlags{
				MigStrategy:     "none",
				FailOnInitError: true,
				GFD: config.GFDCommandLineFlags{
					Oneshot:       false,
					OutputFile:    "./gfd-test-loop",
					SleepInterval: time.Second,
					NoTimestamp:   false,
				},
			},
		},
	}

	MachineTypePath = "/tmp/machine-type"
	machineType := []byte("product-name\n")
	err := ioutil.WriteFile("/tmp/machine-type", machineType, 0644)
	require.NoError(t, err, "Write machine type mock file")

	defer func() {
		err = os.Remove(MachineTypePath)
		require.NoError(t, err, "Removing machine type mock file")
		err = os.Remove(conf.Flags.GFD.OutputFile)
		require.NoError(t, err, "Removing output file")
	}()

	var runError error
	go func() {
		runError = run(nvmlMock, vgpuMock, conf)
	}()

	outFileModificationTime := make([]int64, 2)
	timestampLabels := make([]string, 2)
	// Read two iterations of the output file
	for i := 0; i < 2; i++ {
		outFile, err := waitForFile(conf.Flags.GFD.OutputFile, 5, time.Second)
		require.NoErrorf(t, err, "Open output file: %d", i)

		var outFileStat os.FileInfo
		var ts int64

		for attempt := 0; i > 0 && attempt < 3; attempt++ {
			// We ensure that the output file has been modified. Note, we expect the contents to remain the
			// same so we check the modification timestamp of the file.
			outFileStat, err = os.Stat(conf.Flags.GFD.OutputFile)
			require.NoError(t, err, "Getting output file info")

			ts = outFileStat.ModTime().Unix()
			if ts > outFileModificationTime[0] {
				break
			}
			// We wait for conf.SleepInterval, as the labels should be updated at least once in that period
			time.Sleep(conf.Flags.GFD.SleepInterval)
		}
		outFileModificationTime[i] = ts

		output, err := ioutil.ReadAll(outFile)
		require.NoErrorf(t, err, "Read output file: %d", i)

		err = outFile.Close()
		require.NoErrorf(t, err, "Close output file: %d", i)

		err = checkResult(output, cfg.Path("tests/expected-output.txt"), false)
		require.NoErrorf(t, err, "Checking result: %d", i)
		err = checkResult(output, cfg.Path("tests/expected-output-vgpu.txt"), true)
		require.NoErrorf(t, err, "Checking result for vgpu labels: %d", i)

		labels, err := buildLabelMapFromOutput(output)
		require.NoErrorf(t, err, "Building map of labels from output file: %d", i)

		require.Containsf(t, labels, "nvidia.com/gfd.timestamp", "Missing timestamp: %d", i)
		timestampLabels[i] = labels["nvidia.com/gfd.timestamp"]

		require.Containsf(t, labels, "nvidia.com/vgpu.present", "Missing vgpu present label: %d", i)
		require.Containsf(t, labels, "nvidia.com/vgpu.host-driver-version", "Missing vGPU host driver version label: %d", i)
		require.Containsf(t, labels, "nvidia.com/vgpu.host-driver-branch", "Missing vGPU host driver branch label: %d", i)
	}
	require.Greater(t, outFileModificationTime[1], outFileModificationTime[0], "Output file not modified")
	require.Equal(t, timestampLabels[1], timestampLabels[0], "Timestamp label changed")

	require.NoError(t, runError, "Error from run")
}

func TestFailOnNVMLInitError(t *testing.T) {
	nvmlMock := NewTestNvmlMock()
	vgpuMock := NewTestVGPUMock()
	conf := &config.Config{
		Flags: config.Flags{
			CommandLineFlags: config.CommandLineFlags{
				MigStrategy:     "none",
				FailOnInitError: true,
				GFD: config.GFDCommandLineFlags{
					Oneshot:       true,
					OutputFile:    "./gfd-test-fail-on-nvml-init",
					SleepInterval: 500 * time.Millisecond,
					NoTimestamp:   false,
				},
			},
		},
	}

	MachineTypePath = "/tmp/machine-type"
	machineType := []byte("product-name\n")
	err := ioutil.WriteFile("/tmp/machine-type", machineType, 0644)
	require.NoError(t, err, "Write machine type mock file")

	defer func() {
		err = os.Remove(MachineTypePath)
		require.NoError(t, err, "Removing machine type mock file")
	}()

	defer func() {
		// Remove the output file created by any "success" cases below
		err = os.Remove(conf.Flags.GFD.OutputFile)
		require.NoError(t, err, "Removing output file")
	}()

	// Test for case (errorOnInit = true, failOnInitError = true, no other errors)
	nvmlMock.ErrorOnInit = true
	conf.Flags.FailOnInitError = true
	conf.Flags.MigStrategy = "none"
	err = run(nvmlMock, vgpuMock, conf)
	require.Error(t, err, "Expected error from NVML Init")

	// Test for case (errorOnInit = true, failOnInitError = true, some other error)
	nvmlMock.ErrorOnInit = true
	conf.Flags.FailOnInitError = true
	conf.Flags.MigStrategy = "bogus"
	err = run(nvmlMock, vgpuMock, conf)
	require.Error(t, err, "Expected error from NVML Init")

	// Test for case (errorOnInit = true, failOnInitError = false, no other errors)
	nvmlMock.ErrorOnInit = true
	conf.Flags.FailOnInitError = false
	conf.Flags.MigStrategy = "none"
	err = run(nvmlMock, vgpuMock, conf)
	require.NoError(t, err, "Expected to skip error from NVML Init")

	// Test for case (errorOnInit = true, failOnInitError = false, some other error)
	nvmlMock.ErrorOnInit = true
	conf.Flags.FailOnInitError = false
	conf.Flags.MigStrategy = "bogus"
	err = run(nvmlMock, vgpuMock, conf)
	require.NoError(t, err, "Expected to skip error from NVML Init")

	// Test for case (errorOnInit = false, failOnInitError = true, no other errors)
	nvmlMock.ErrorOnInit = false
	conf.Flags.FailOnInitError = true
	conf.Flags.MigStrategy = "none"
	err = run(nvmlMock, vgpuMock, conf)
	require.NoError(t, err, "Expected no errors")

	// Test for case (errorOnInit = false, failOnInitError = true, some other error)
	nvmlMock.ErrorOnInit = false
	conf.Flags.FailOnInitError = true
	conf.Flags.MigStrategy = "bogus"
	err = run(nvmlMock, vgpuMock, conf)
	require.Error(t, err, "Expected error since MIGStrategy is 'bogus'")

	// Test for case (errorOnInit = false, failOnInitError = false, no other errors)
	nvmlMock.ErrorOnInit = false
	conf.Flags.FailOnInitError = false
	conf.Flags.MigStrategy = "none"
	err = run(nvmlMock, vgpuMock, conf)
	require.NoError(t, err, "Expected no errors")

	// Test for case (errorOnInit = false, failOnInitError = false, some other error)
	nvmlMock.ErrorOnInit = false
	conf.Flags.FailOnInitError = false
	conf.Flags.MigStrategy = "bogus"
	err = run(nvmlMock, vgpuMock, conf)
	require.Error(t, err, "Expected error since MIGStrategy is 'bogus'")
}

func buildLabelMapFromOutput(output []byte) (map[string]string, error) {
	labels := make(map[string]string)

	lines := strings.Split(strings.TrimRight(string(output), "\n"), "\n")
	for _, line := range lines {
		split := strings.Split(line, "=")
		if len(split) != 2 {
			return nil, fmt.Errorf("unexpected format in line: '%v'", line)
		}
		key := split[0]
		value := split[1]

		if v, ok := labels[key]; ok {
			return nil, fmt.Errorf("duplicate label '%v': %v (overwrites %v)", key, v, value)
		}
		labels[key] = value
	}

	return labels, nil
}

func checkResult(result []byte, expectedOutputPath string, isVGPU bool) error {
	expected, err := ioutil.ReadFile(expectedOutputPath)
	if err != nil {
		return fmt.Errorf("opening expected output file: %v", err)
	}

	var expectedRegexps []*regexp.Regexp
	for _, line := range strings.Split(strings.TrimRight(string(expected), "\n"), "\n") {
		expectedRegexps = append(expectedRegexps, regexp.MustCompile(line))
	}

LOOP:
	for _, line := range strings.Split(strings.TrimRight(string(result), "\n"), "\n") {
		if isVGPU {
			if !strings.Contains(line, "vgpu") {
				// ignore other labels when vgpu file is specified
				continue
			}
		} else {
			if strings.Contains(line, "vgpu") {
				// ignore vgpu labels when non vgpu file is specified
				continue
			}
		}
		for _, regex := range expectedRegexps {
			if regex.MatchString(line) {
				continue LOOP
			}
		}
		return fmt.Errorf("line does not match any regexp: %v", string(line))
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
