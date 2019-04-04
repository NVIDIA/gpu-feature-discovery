// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/template"
	"time"
)

const (
	// Bin : Name of the binary
	Bin            = "gpu-feature-discovery"
	// OutputFilePath : Path to the output file
	// TODO: Change path and get it by config
	OutputFilePath = "./output"
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
	log.Print("OutputFilePath: ", OutputFilePath)

	outputFile, err := os.Create(OutputFilePath)
	if err != nil {
		log.Fatalf("Fail to create output file: %v", err)
	}

	log.Print("Start running")
	run(nvmlLib, conf, outputFile)
	log.Print("Exiting")

	outputFile.Close()
}

func run(nvmlInterface NvmlInterface, conf Conf, out io.Writer) {

	if err := nvmlInterface.Init(); err != nil {
		// TODO: Update README and links
		log.Printf("Failed to initialize NVML: %s.", err)
		log.Printf("If this is a GPU node, did you set the docker default runtime to `nvidia`?")
		log.Printf("You can check the prerequisites at: https://github.com/NVIDIA/gpu-feature-discovery")
		log.Printf("You can learn how to set the runtime at: https://github.com/NVIDIA/gpu-feature-discovery#quick-start")
		return
	}
	defer func() {
		err := nvmlInterface.Shutdown()
		if err != nil {
			log.Println("Shutdown of NVML returned:", nvmlInterface.Shutdown())
		}
	}()

	count, err := nvmlInterface.GetDeviceCount()
	if err != nil {
		log.Fatal("Error getting device count: ", err)
	}

	if count < 1 {
		log.Fatal("Error: no device found on the node")
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


L:
	for {

		device, err := nvmlInterface.NewDevice(0)
		if err != nil {
			log.Fatal("Error getting device: ", err)
		}

		driverVersion, err := nvmlInterface.GetDriverVersion()
		if err != nil {
			log.Fatal("Error getting driver version: ", err)
		}
		// TODO: Change label format
		fmt.Fprintf(out, "nvidia-driver-version=%s\n", driverVersion)

		log.Print("Writing labels to output file")
		err = t.Execute(out, device)
		if err != nil {
			log.Fatal("Template error: ", err)
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
}
