// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"text/template"
	"time"
)

const (
	ProgName   = "gpu-feature-discovery"
	// TODO: Get version from git
	Version    = "0.0.1-alpha"
	// TODO: Change path and get it by config
	OutputFilePath = "./output"
	// TODO: Change label format
	DEVICEINFO = `{{if .Model}}nvidia-model={{replace .Model " " "-" -1}}{{end}}
{{if .Memory}}nvidia-memory={{.Memory}}{{end}}
`
)

func run(nvmlInterface NvmlInterface, conf Conf, out io.Writer) {

	if err := nvmlInterface.Init(); err != nil {
		log.Fatal("Failed to initialize NVML: ", err)
	}
	defer nvmlInterface.Shutdown()

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

	t := template.Must(template.New("Device").Funcs(funcMap).Parse(DEVICEINFO))

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

		time.Sleep(conf.SleepInterval)
	}
}

func main() {

	log.SetPrefix(ProgName + ": ")

	log.Printf("Running %s in version %s", ProgName, Version)

	nvmlLib := NvmlLib{}

	log.Print("Load configuration")
	conf := Conf{}
	conf.getConfFromArgv(os.Args)
	conf.getConfFromEnv()
	log.Print("Oneshot: ", conf.Oneshot)
	log.Print("SleepInterval: ", conf.SleepInterval)
	log.Print("OutputFilePath: ", OutputFilePath)

	outputFile, err := os.Create(OutputFilePath)
	if err != nil {
		log.Printf("Fail to create output file: %v", err)
		os.Exit(1)
	}

	log.Print("Start running")
	run(nvmlLib, conf, outputFile)
	log.Print("Exiting")

	outputFile.Close()
}
