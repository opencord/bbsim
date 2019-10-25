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
	"github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsim/api"
	"github.com/opencord/bbsim/internal/bbsim/devices"
	"github.com/opencord/bbsim/internal/common"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"os"
	"runtime/pprof"
	"sync"
)

func startApiServer(channel chan bool, group *sync.WaitGroup) {
	// TODO make configurable
	address := "0.0.0.0:50070"
	log.Debugf("APIServer Listening on: %v", address)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("APIServer failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	bbsim.RegisterBBSimServer(grpcServer, api.BBSimServer{})

	reflection.Register(grpcServer)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go grpcServer.Serve(lis)

	for {
		_, ok := <-channel
		if !ok {
			// if the olt channel is closed, stop the gRPC server
			log.Warnf("Stopping API gRPC server")
			grpcServer.Stop()
			wg.Done()
			break
		}
	}

	wg.Wait()
	group.Done()
	return
}

func main() {
	options := common.GetBBSimOpts()

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
	}).Info("BroadBand Simulator is on")

	// control channels, they are only closed when the goroutine needs to be terminated
	oltDoneChannel := make(chan bool)
	apiDoneChannel := make(chan bool)

	wg := sync.WaitGroup{}
	wg.Add(2)

	olt := devices.CreateOLT(options.OltID, options.NumNniPerOlt, options.NumPonPerOlt, options.NumOnuPerPon, options.STag, options.CTagInit, &oltDoneChannel, &apiDoneChannel, false)
	go devices.StartOlt(olt, &wg)
	log.Debugf("Created OLT with id: %d", options.OltID)
	go startApiServer(apiDoneChannel, &wg)
	log.Debugf("Started APIService")

	wg.Wait()

	defer func() {
		log.Info("BroadBand Simulator is off")
		if *options.ProfileCpu != "" {
			log.Info("Stopping profiler")
			pprof.StopCPUProfile()
		}
	}()
}
