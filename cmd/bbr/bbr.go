/*
 * Copyright 2018-present Open Networking Foundation

 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	bbrdevices "github.com/opencord/bbsim/internal/bbr/devices"
	"github.com/opencord/bbsim/internal/bbsim/devices"
	"github.com/opencord/bbsim/internal/common"
	log "github.com/sirupsen/logrus"
	"os"
	"runtime/pprof"
	"time"
)

func main() {
	options := common.GetBBROpts()

	common.SetLogLevel(log.StandardLogger(), options.LogLevel, options.LogCaller)

	if *options.ProfileCpu != "" {
		// start profiling
		log.Infof("Creating profile file at: %s", *options.ProfileCpu)
		f, err := os.Create(*options.ProfileCpu)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
	}

	log.WithFields(log.Fields{
		"OltID":        options.OltID,
		"NumNniPerOlt": options.NumNniPerOlt,
		"NumPonPerOlt": options.NumPonPerOlt,
		"NumOnuPerPon": options.NumOnuPerPon,
	}).Info("BroadBand Reflector is on")

	// NOTE this are probably useless in the MockOLT case, check if we can avoid using them in the CreateOlt method
	oltDoneChannel := make(chan bool)
	apiDoneChannel := make(chan bool)

	// create the OLT device
	olt := devices.CreateOLT(options.OltID, options.NumNniPerOlt, options.NumPonPerOlt, options.NumOnuPerPon, options.STag, options.CTagInit, &oltDoneChannel, &apiDoneChannel, true)
	oltMock := bbrdevices.OltMock{
		Olt:          olt,
		BBSimIp:      options.BBSimIp,
		BBSimPort:    options.BBSimPort,
		BBSimApiPort: options.BBSimApiPort,
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
