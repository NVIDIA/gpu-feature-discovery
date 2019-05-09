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

func run(nvmlInterface NvmlInterface, conf Conf) error {

	if err := nvmlInterface.Init(); err != nil {
		// TODO: Update README and links
		log.Printf("Failed to initialize NVML: %s.", err)
		log.Printf("If this is a GPU node, did you set the docker default runtime to `nvidia`?")
		log.Printf("You can check the prerequisites at: https://github.com/NVIDIA/gpu-feature-discovery")
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

	// TODO: Change label format
	const deviceTemplate = `{{if .Model}}nvidia-model={{replace .Model " " "-" -1}}{{end}}
{{if .Memory}}nvidia-memory={{.Memory}}{{end}}
`

	funcMap := template.FuncMap{
		"replace": strings.Replace,
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

		log.Print("Writing labels to output file")
		fmt.Fprintf(tmpOutputFile, "nvidia-timestamp=%d\n", time.Now().Unix())

		// TODO: Change label format
		fmt.Fprintf(tmpOutputFile, "nvidia-driver-version=%s\n", driverVersion)

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
