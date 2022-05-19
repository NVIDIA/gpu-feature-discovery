// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/NVIDIA/gpu-feature-discovery/internal/lm"
	"github.com/NVIDIA/gpu-feature-discovery/internal/nvml"
	"github.com/NVIDIA/gpu-feature-discovery/internal/vgpu"
	spec "github.com/NVIDIA/k8s-device-plugin/api/config/v1"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

const (
	// Bin : Name of the binary
	Bin = "gpu-feature-discovery"
)

var (
	// Version : Version of the binary
	// This will be set using ldflags at compile time
	version = ""
	// MachineTypePath : Path to the file describing the machine type
	// This will be override during unit testing
	MachineTypePath = "/sys/class/dmi/id/product_name"
)

func main() {
	var config spec.Config
	// TODO: Change this one we switch to loading values from config files.
	flags := &config.Flags
	var configFile string

	c := cli.NewApp()
	c.Name = "GPU Feature Discovery"
	c.Usage = "generate labels for NVIDIA devices"
	c.Version = version
	c.Action = func(ctx *cli.Context) error {
		return start(ctx, &config)
	}

	c.Flags = []cli.Flag{
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:        "mig-strategy",
				Value:       spec.MigStrategyNone,
				Usage:       "the desired strategy for exposing MIG devices on GPUs that support it:\n\t\t[none | single | mixed]",
				Destination: &flags.MigStrategy,
				EnvVars:     []string{"GFD_MIG_STRATEGY", "MIG_STRATEGY"},
			},
		),
		altsrc.NewBoolFlag(
			&cli.BoolFlag{
				Name:        "fail-on-init-error",
				Value:       true,
				Usage:       "fail the plugin if an error is encountered during initialization, otherwise block indefinitely",
				Destination: &flags.FailOnInitError,
				EnvVars:     []string{"GFD_FAIL_ON_INIT_ERROR", "FAIL_ON_INIT_ERROR"},
			},
		),
		altsrc.NewBoolFlag(
			&cli.BoolFlag{
				Name:        "oneshot",
				Value:       false,
				Usage:       "Label once and exit",
				Destination: &flags.GFD.Oneshot,
				EnvVars:     []string{"GFD_ONESHOT"},
			},
		),
		altsrc.NewBoolFlag(
			&cli.BoolFlag{
				Name:        "no-timestamp",
				Value:       false,
				Usage:       "Do not add the timestamp to the labels",
				Destination: &flags.GFD.NoTimestamp,
				EnvVars:     []string{"GFD_NO_TIMESTAMP"},
			},
		),
		altsrc.NewDurationFlag(
			&cli.DurationFlag{
				Name:        "sleep-interval",
				Value:       60 * time.Second,
				Usage:       "Time to sleep between labeling",
				Destination: &flags.GFD.SleepInterval,
				EnvVars:     []string{"GFD_SLEEP_INTERVAL"},
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:        "output",
				Aliases:     []string{"o"},
				Value:       "/etc/kubernetes/node-feature-discovery/features.d/gfd",
				Destination: &flags.GFD.OutputFile,
				EnvVars:     []string{"GFD_OUTPUT_FILE"},
			},
		),
		&cli.StringFlag{
			Name:        "config-file",
			Usage:       "the path to a config file as an alternative to command line options or environment variables",
			Destination: &configFile,
			EnvVars:     []string{"GFD_CONFIG_FILE", "CONFIG_FILE"},
		},
	}

	err := c.Run(os.Args)
	if err != nil {
		log.SetOutput(os.Stderr)
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func start(ctx *cli.Context, config *spec.Config) error {
	log.SetPrefix(Bin + ": ")

	if version == "" {
		log.Print("Version is not set.")
		log.Fatal("Be sure to compile with '-ldflags \"-X main.version=${GFD_VERSION}\"' and to set $GFD_VERSION")
	}

	log.Printf("Running %s in version %s", Bin, version)

	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %v", err)
	}
	log.Printf("\nRunning with config:\n%v", string(configJSON))

	nvml := nvml.Lib{}

	vgpul := vgpu.NewVGPULib(vgpu.NewNvidiaPCILib())

	log.Print("Start running")
	err = run(nvml, vgpul, config)
	if err != nil {
		log.Printf("Unexpected error: %v", err)
	}
	log.Print("Exiting")
	return err
}

func run(nvml nvml.Nvml, vgpu vgpu.Interface, config *spec.Config) error {
	defer func() {
		if !config.Flags.GFD.Oneshot {
			err := removeOutputFile(config.Flags.GFD.OutputFile)
			if err != nil {
				log.Printf("Warning: Error removing output file: %v", err)
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	exitChan := make(chan bool)

	go func() {
		select {
		case s := <-sigChan:
			log.Printf("Received signal \"%v\", shutting down.", s)
			exitChan <- true
		}
	}()

	gfdLabels, err := lm.NewTimestampLabeler(config).Labels()
	if err != nil {
		return fmt.Errorf("error generating timestamp labels: %v", err)
	}

L:
	for {
		nvmlLabels, err := getNVMLLabels(nvml, config.Flags.MigStrategy)
		if err != nil {
			isInitError := nvml.IsInitError(err)
			if !isInitError || (isInitError && config.Flags.FailOnInitError) {
				return fmt.Errorf("error generating NVML labels: %v", err)
			}
			log.Printf("Warning: Error generating NVML labels: %v", err)
		}

		vGPULabels, err := lm.NewVGPULabeler(vgpu).Labels()
		if err != nil {
			return fmt.Errorf("error generating vGPU labels: %v", err)
		}

		if len(nvmlLabels) == 0 && len(vGPULabels) == 0 {
			log.Printf("Warning: no labels generated from any source")
		}

		allLabels := lm.AsSet(
			gfdLabels,
			vGPULabels,
			nvmlLabels,
		)
		log.Print("Writing labels to output file")
		err = allLabels.WriteToFile(config.Flags.GFD.OutputFile)
		if err != nil {
			return fmt.Errorf("error writing file '%s': %v", config.Flags.GFD.OutputFile, err)
		}

		if config.Flags.GFD.Oneshot {
			break
		}

		log.Print("Sleeping for ", config.Flags.GFD.SleepInterval)

		select {
		case <-exitChan:
			break L
		case <-time.After(config.Flags.GFD.SleepInterval):
			break
		}
	}

	return nil
}

func getNVMLLabels(nvml nvml.Nvml, MigStrategy string) (lm.Labels, error) {
	if err := nvml.Init(); err != nil {
		return nil, nvml.AsInitError(fmt.Errorf("failed to initialize NVML: %v", err))
	}

	defer func() {
		err := nvml.Shutdown()
		if err != nil {
			fmt.Printf("Warning: Shutdown of NVML returned: %v", err)
		}
	}()

	count, err := nvml.GetDeviceCount()
	if err != nil {
		return nil, fmt.Errorf("error getting device count: %v", err)
	}

	if count == 0 {
		return nil, nil
	}

	commonLabels, err := generateCommonLabels(nvml)
	if err != nil {
		return nil, fmt.Errorf("error generating common labels: %v", err)
	}

	migStrategy, err := lm.NewMigStrategy(MigStrategy, nvml)
	if err != nil {
		return nil, fmt.Errorf("error creating MIG strategy: %v", err)
	}

	migStrategyLabels, err := migStrategy.Labels()
	if err != nil {
		return nil, fmt.Errorf("error generating labels from MIG strategy: %v", err)
	}

	allLabels := make(lm.Labels)
	for k, v := range commonLabels {
		allLabels[k] = v
	}
	for k, v := range migStrategyLabels {
		allLabels[k] = v
	}

	return allLabels, nil
}

func generateCommonLabels(nvml nvml.Nvml) (lm.Labels, error) {
	driverVersion, err := nvml.GetDriverVersion()
	if err != nil {
		return nil, fmt.Errorf("error getting driver version: %v", err)
	}

	driverVersionSplit := strings.Split(driverVersion, ".")
	if len(driverVersionSplit) > 3 || len(driverVersionSplit) < 2 {
		return nil, fmt.Errorf("error getting driver version: Version \"%s\" does not match format \"X.Y[.Z]\"", driverVersion)
	}

	driverMajor := driverVersionSplit[0]
	driverMinor := driverVersionSplit[1]
	driverRev := ""
	if len(driverVersionSplit) > 2 {
		driverRev = driverVersionSplit[2]
	}

	cudaMajor, cudaMinor, err := nvml.GetCudaDriverVersion()
	if err != nil {
		return nil, fmt.Errorf("error getting cuda driver version: %v", err)
	}

	machineType, err := getMachineType(MachineTypePath)
	if err != nil {
		return nil, fmt.Errorf("error getting machine type: %v", err)
	}

	device, err := nvml.NewDevice(0)
	if err != nil {
		return nil, fmt.Errorf("error getting device: %v", err)
	}

	computeMajor, computeMinor, err := device.GetCudaComputeCapability()
	if err != nil {
		return nil, fmt.Errorf("failed to determine CUDA compute capability: %v", err)
	}

	labels := make(lm.Labels)
	labels["nvidia.com/cuda.driver.major"] = driverMajor
	labels["nvidia.com/cuda.driver.minor"] = driverMinor
	labels["nvidia.com/cuda.driver.rev"] = driverRev
	labels["nvidia.com/cuda.runtime.major"] = fmt.Sprintf("%d", *cudaMajor)
	labels["nvidia.com/cuda.runtime.minor"] = fmt.Sprintf("%d", *cudaMinor)
	labels["nvidia.com/gpu.machine"] = strings.Replace(machineType, " ", "-", -1)
	if computeMajor != 0 {
		family, _ := device.GetArchFamily()
		labels["nvidia.com/gpu.family"] = family
		labels["nvidia.com/gpu.compute.major"] = fmt.Sprintf("%d", computeMajor)
		labels["nvidia.com/gpu.compute.minor"] = fmt.Sprintf("%d", computeMinor)
	}

	return labels, nil
}

func getMachineType(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func removeOutputFile(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to retrieve absolute path of output file: %v", err)
	}

	absDir := filepath.Dir(absPath)
	tmpDir := filepath.Join(absDir, "gfd-tmp")

	err = os.RemoveAll(tmpDir)
	if err != nil {
		return fmt.Errorf("failed to remove temporary output directory: %v", err)
	}

	err = os.Remove(absPath)
	if err != nil {
		return fmt.Errorf("failed to remove output file: %v", err)
	}

	return nil
}
