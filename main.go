// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"
	"time"
)

const (
	// Bin : Name of the binary
	Bin = "gpu-feature-discovery"
)

var (
	// Version : Version of the binary
	// This will be set using ldflags at compile time
	Version = ""
	// MachineTypePath : Path to the file describing the machine type
	// This will be override during unit testing
	MachineTypePath = "/sys/class/dmi/id/product_name"
)

func main() {

	log.SetPrefix(Bin + ": ")

	if Version == "" {
		log.Print("Version is not set.")
		log.Fatal("Be sure to compile with '-ldflags \"-X main.Version=${GFD_VERSION}\"' and to set $GFD_VERSION")
	}

	log.Printf("Running %s in version %s", Bin, Version)

	nvmlLib := NvmlLib{}

	conf := Conf{}
	conf.getConfFromArgv(os.Args)
	conf.getConfFromEnv()
	log.Print("Loaded configuration:")
	log.Print("Oneshot: ", conf.Oneshot)
	log.Print("SleepInterval: ", conf.SleepInterval)
	log.Print("OutputFilePath: ", conf.OutputFilePath)

	log.Print("Start running")
	err := run(nvmlLib, conf)
	if err != nil {
		log.Printf("Unexpected error: %v", err)
	}
	log.Print("Exiting")
}

func getArchFamily(cudaComputeMajor int) string {
	m := map[int]string{
		1: "tesla",
		2: "fermi",
		3: "kepler",
		5: "maxwell",
		6: "pascal",
	}

	f, ok := m[cudaComputeMajor]
	if !ok {
		return "undefined"
	}
	return f
}

func getMachineType() (string, error) {
	data, err := ioutil.ReadFile(MachineTypePath)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

func run(nvmlInterface NvmlInterface, conf Conf) error {

	if err := nvmlInterface.Init(); err != nil {
		log.Printf("Failed to initialize NVML: %s.", err)
		log.Printf("If this is a GPU node, did you set the docker default runtime to `nvidia`?")
		log.Printf("You can check the prerequisites at: https://github.com/NVIDIA/gpu-feature-discovery#prerequisites")
		log.Printf("You can learn how to set the runtime at: https://github.com/NVIDIA/gpu-feature-discovery#quick-start")
		return err
	}

	defer func() {
		err := nvmlInterface.Shutdown()
		if err != nil {
			log.Println("Shutdown of NVML returned:", nvmlInterface.Shutdown())
		}
	}()

	count, err := nvmlInterface.GetDeviceCount()
	if err != nil {
		return fmt.Errorf("Error getting device count: %v", err)
	}

	if count < 1 {
		return fmt.Errorf("Error: no device found on the node")
	}

	const deviceTemplate = `{{if .Model}}nvidia.com/gpu.product={{replace .Model " " "-" -1}}{{end}}
{{if .Memory}}nvidia.com/gpu.memory={{.Memory}}{{end}}
{{if .CudaComputeCapability.Major}}nvidia.com/gpu.family={{getArchFamily .CudaComputeCapability.Major}}
nvidia.com/gpu.compute.major={{.CudaComputeCapability.Major}}
nvidia.com/gpu.compute.minor={{.CudaComputeCapability.Minor}}{{end}}
`

	funcMap := template.FuncMap{
		"replace":       strings.Replace,
		"getArchFamily": getArchFamily,
	}

	t := template.Must(template.New("Device").Funcs(funcMap).Parse(deviceTemplate))

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

	outputFileAbsPath, err := filepath.Abs(conf.OutputFilePath)
	if err != nil {
		return fmt.Errorf("Failed to retrieve absolute path of output file: %v", err)
	}
	tmpDirPath := filepath.Dir(outputFileAbsPath) + "/gfd-tmp"

	err = os.Mkdir(tmpDirPath, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("Failed to create temporary directory: %v", err)
	}

L:
	for {
		tmpOutputFile, err := ioutil.TempFile(tmpDirPath, "gfd-")
		if err != nil {
			return fmt.Errorf("Fail to create temporary output file: %v", err)
		}

		device, err := nvmlInterface.NewDevice(0)
		if err != nil {
			return fmt.Errorf("Error getting device: %v", err)
		}

		driverVersion, err := nvmlInterface.GetDriverVersion()
		if err != nil {
			return fmt.Errorf("Error getting driver version: %v", err)
		}

		driverVersionSplit := strings.Split(driverVersion, ".")
		if len(driverVersionSplit) > 3 || len(driverVersionSplit) < 2 {
			return fmt.Errorf("Error getting driver version: Version \"%s\" does not match format \"X.Y[.Z]\"", driverVersion)
		}

		driverMajor := driverVersionSplit[0]
		driverMinor := driverVersionSplit[1]
		driverRev := ""
		if len(driverVersionSplit) > 2 {
			driverRev = driverVersionSplit[2]
		}

		cudaMajor, cudaMinor, err := nvmlInterface.GetCudaDriverVersion()
		if err != nil {
			return fmt.Errorf("Error getting cuda driver version: %v", err)
		}

		machineType, err := getMachineType()
		if err != nil {
			return fmt.Errorf("Error getting machine type: %v", err)
		}

		log.Print("Writing labels to output file")
		fmt.Fprintf(tmpOutputFile, "nvidia.com/gfd.timestamp=%d\n", time.Now().Unix())

		fmt.Fprintf(tmpOutputFile, "nvidia.com/cuda.driver.major=%s\n", driverMajor)
		fmt.Fprintf(tmpOutputFile, "nvidia.com/cuda.driver.minor=%s\n", driverMinor)
		fmt.Fprintf(tmpOutputFile, "nvidia.com/cuda.driver.rev=%s\n", driverRev)
		fmt.Fprintf(tmpOutputFile, "nvidia.com/cuda.runtime.major=%d\n", *cudaMajor)
		fmt.Fprintf(tmpOutputFile, "nvidia.com/cuda.runtime.minor=%d\n", *cudaMinor)
		fmt.Fprintf(tmpOutputFile, "nvidia.com/gpu.machine=%s\n", strings.Replace(machineType, " ", "-", -1))
		fmt.Fprintf(tmpOutputFile, "nvidia.com/gpu.count=%s\n", count)

		err = t.Execute(tmpOutputFile, device)
		if err != nil {
			return fmt.Errorf("Template error: %v", err)
		}

		err = tmpOutputFile.Chmod(0644)
		if err != nil {
			return fmt.Errorf("Error chmod temporary file: %v", err)
		}

		err = tmpOutputFile.Close()
		if err != nil {
			return fmt.Errorf("Error closing temporary file: %v", err)
		}

		err = os.Rename(tmpOutputFile.Name(), conf.OutputFilePath)
		if err != nil {
			return fmt.Errorf("Error moving temporary file '%s': %v", conf.OutputFilePath, err)
		}

		if conf.Oneshot {
			break
		}

		log.Print("Sleeping for ", conf.SleepInterval)

		select {
		case <-exitChan:
			break L
		case <-time.After(conf.SleepInterval):
			break
		}
	}

	return nil
}
