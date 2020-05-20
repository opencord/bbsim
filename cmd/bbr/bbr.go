// Copyright 2018-present Open Networking Foundation
// SPDX-FileCopyrightText: 2019 Open Networking Foundation <info@opennetworking.org>
// SPDX-FileCopyrightText: 2020 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

// BBR (BBSim Reflector) is a tool designed to scale test BBSim
// It shared most of the code with BBSim itself and replies to messages
// pretending to be VOLTHA.
// The idea behind it is that, given that the BBSim and BBR are based on the same
// codebase, BBR is acting as a wall for BBSim. And you can't beat the wall.
package main

import (
	"os"
	"runtime/pprof"
	"time"

	bbrdevices "github.com/opencord/bbsim/internal/bbr/devices"
	"github.com/opencord/bbsim/internal/bbsim/devices"
	"github.com/opencord/bbsim/internal/common"
	log "github.com/sirupsen/logrus"
)

// usage
func main() {
	options := common.GetBBROpts()

	common.SetLogLevel(log.StandardLogger(), options.BBR.LogLevel, options.BBR.LogCaller)

	if options.LogFile != "" {
		file, err := os.OpenFile(options.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			log.StandardLogger().Out = file
		} else {
			log.Fatal("Failed to log to file, using default stderr")
		}
	}

	if *options.BBSim.CpuProfile != "" {
		// start profiling
		log.Infof("Creating profile file at: %s", *options.BBSim.CpuProfile)
		f, err := os.Create(*options.BBSim.CpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
	}

	log.WithFields(log.Fields{
		"OltID":        options.Olt.ID,
		"NumNniPerOlt": options.Olt.NniPorts,
		"NumPonPerOlt": options.Olt.PonPorts,
		"NumOnuPerPon": options.Olt.OnusPonPort,
		"BBSimIp":      options.BBSimIp,
		"BBSimPort":    options.BBSimPort,
		"LogLevel":     options.BBR.LogLevel,
		"TotalOnus":    options.Olt.PonPorts * options.Olt.OnusPonPort,
	}).Info("BroadBand Reflector is on")

	// create the OLT device
	olt := devices.CreateOLT(
		*options.BBSimYamlConfig,
		true,
	)

	onuIdMap := make(map[uint32]uint32, options.Olt.PonPorts)

	oltMock := bbrdevices.OltMock{
		Olt:           olt,
		TargetOnus:    int(options.Olt.PonPorts * options.Olt.OnusPonPort),
		CompletedOnus: 0,
		BBSimIp:       options.BBSimIp,
		BBSimPort:     options.BBSimPort,
		BBSimApiPort:  options.BBSimApiPort,
		LastUsedOnuId: onuIdMap,
	}

	// start the enable sequence
	startTime := time.Now()
	defer func() {
		endTime := time.Now()
		runTime := endTime.Sub(startTime)
		log.WithField("Duration", runTime).Info("BBR done!")
	}()
	oltMock.Start()
}
