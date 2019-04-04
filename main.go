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
	bin            = "gpu-feature-discovery"
	// TODO: Change path and get it by config
	outputFilePath = "./output"
	// TODO: Change label format
	deviceTemplate = `{{if .Model}}nvidia-model={{replace .Model " " "-" -1}}{{end}}
{{if .Memory}}nvidia-memory={{.Memory}}{{end}}
`
)

var (
	// This will be set using ldflags at compile time
	version = ""
)

func main() {

	log.SetPrefix(bin + ": ")

	if version == "" {
		log.Print("Version is not set.")
		log.Print("Be sure to compile with '-ldflags \"-X main.version=${GFD_VERSION}\"' and to set $GFD_VERSION")
		os.Exit(1)
	}

	log.Printf("Running %s in version %s", bin, version)

	nvmlLib := NvmlLib{}

	conf := Conf{}
	conf.getConfFromArgv(os.Args)
	conf.getConfFromEnv()
	log.Print("Loaded configuration:")
	log.Print("Oneshot: ", conf.Oneshot)
	log.Print("SleepInterval: ", conf.SleepInterval)
	log.Print("OutputFilePath: ", outputFilePath)

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		log.Printf("Fail to create output file: %v", err)
		os.Exit(1)
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

	funcMap := template.FuncMap{
		"replace": strings.Replace,
	}

	t := template.Must(template.New("Device").Funcs(funcMap).Parse(deviceTemplate))

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

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

		err = t.Execute(out, device)
		if err != nil {
			log.Fatal("Template error: ", err)
		}

		if conf.Oneshot {
			break
		}

		select {
			case s := <-sigs:
				log.Printf("Received signal \"%v\", shutting down.", s)
				break L
		}

		time.Sleep(conf.SleepInterval)
	}
}
