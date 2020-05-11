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
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/api/legacy"
	"github.com/opencord/bbsim/internal/bbsim/api"
	"github.com/opencord/bbsim/internal/bbsim/devices"
	"github.com/opencord/bbsim/internal/bbsim/responders/sadis"
	"github.com/opencord/bbsim/internal/common"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func startApiServer(apiDoneChannel chan bool, group *sync.WaitGroup) {
	address := common.Options.BBSim.ApiAddress
	log.Debugf("APIServer listening on %v", address)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("APIServer failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	bbsim.RegisterBBSimServer(grpcServer, api.BBSimServer{})

	reflection.Register(grpcServer)

	go grpcServer.Serve(lis)
	go startApiRestServer(apiDoneChannel, group, address)

	select {
	case <-apiDoneChannel:
		// if the API channel is closed, stop the gRPC server
		grpcServer.Stop()
		log.Warnf("Stopping API gRPC server")
	}

	group.Done()
}

// startApiRestServer method starts the REST server (grpc gateway) for BBSim.
func startApiRestServer(apiDoneChannel chan bool, group *sync.WaitGroup, grpcAddress string) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	address := common.Options.BBSim.RestApiAddress

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	if err := bbsim.RegisterBBSimHandlerFromEndpoint(ctx, mux, grpcAddress, opts); err != nil {
		log.Errorf("Could not register API server: %v", err)
		return
	}

	s := &http.Server{Addr: address, Handler: mux}

	go func() {
		log.Infof("REST API server listening on %s", address)
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Errorf("Could not start API server: %v", err)
			return
		}
	}()

	select {
	case <-apiDoneChannel:
		log.Warnf("Stopping API REST server")
		s.Shutdown(ctx)
	}

	group.Done()
}

// This server aims to provide compatibility with the previous BBSim version. It is deprecated and will be removed in the future.
func startLegacyApiServer(apiDoneChannel chan bool, group *sync.WaitGroup) {
	grpcAddress := common.Options.BBSim.LegacyApiAddress
	restAddress := common.Options.BBSim.LegacyRestApiAddress

	log.Debugf("Legacy APIServer listening on %v", grpcAddress)
	listener, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		log.Fatalf("Legacy APIServer failed to listen: %v", err)
		return
	}
	apiserver := grpc.NewServer()
	legacy.RegisterBBSimServiceServer(apiserver, api.BBSimLegacyServer{})

	go apiserver.Serve(listener)
	// Start rest gateway for BBSim server
	go api.StartRestGatewayService(apiDoneChannel, group, grpcAddress, restAddress)

	select {
	case <-apiDoneChannel:
		// if the API channel is closed, stop the gRPC server
		log.Warnf("Stopping legacy API gRPC server")
		apiserver.Stop()
		break

	}

	group.Done()
}

func main() {

	options := common.GetBBSimOpts()

	common.SetLogLevel(log.StandardLogger(), options.BBSim.LogLevel, options.BBSim.LogCaller)
	log.Tracef("BBSim options: %+v", options)

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
		"OltID":                options.Olt.ID,
		"NumNniPerOlt":         options.Olt.NniPorts,
		"NumPonPerOlt":         options.Olt.PonPorts,
		"NumOnuPerPon":         options.Olt.OnusPonPort,
		"TotalOnus":            options.Olt.PonPorts * options.Olt.OnusPonPort,
		"EnableAuth":           options.BBSim.EnableAuth,
		"Dhcp":                 options.BBSim.EnableDhcp,
		"Igmp":                 options.BBSim.EnableIgmp,
		"Delay":                options.BBSim.Delay,
		"Events":               options.BBSim.Events,
		"KafkaEventTopic":      options.BBSim.KafkaEventTopic,
		"ControlledActivation": options.BBSim.ControlledActivation,
		"EnablePerf":           options.BBSim.EnablePerf,
		"CTag":                 options.BBSim.CTag,
		"CTagAllocation":       options.BBSim.CTagAllocation,
		"STag":                 options.BBSim.STag,
		"STagAllocation":       options.BBSim.STagAllocation,
		"SadisFormat":          options.BBSim.SadisFormat,
	}).Info("BroadBand Simulator is on")

	// control channels, they are only closed when the goroutine needs to be terminated
	apiDoneChannel := make(chan bool)

	olt := devices.CreateOLT(
		*options,
		false,
	)

	log.Debugf("Created OLT with id: %d", options.Olt.ID)

	sigs := make(chan os.Signal, 1)
	// stop API servers on SIGTERM
	signal.Notify(sigs, syscall.SIGTERM)

	go func() {
		<-sigs
		close(apiDoneChannel)
	}()

	wg := sync.WaitGroup{}
	wg.Add(4)

	go startApiServer(apiDoneChannel, &wg)
	go startLegacyApiServer(apiDoneChannel, &wg)
	log.Debugf("Started APIService")
	if common.Options.BBSim.SadisServer != false {
		wg.Add(1)
		go sadis.StartRestServer(olt, &wg)
	}

	if options.BBSim.Events {
		// initialize a publisher
		if err := common.InitializePublisher(sarama.NewAsyncProducer, olt.ID); err == nil {
			// start a go routine which will read from channel and publish on kafka
			go common.KafkaPublisher(olt.EventChannel)
		} else {
			log.Errorf("Failed to start kafka publisher: %v", err)
		}
	}

	wg.Wait()

	defer func() {
		log.Info("BroadBand Simulator is off")
		if *options.BBSim.CpuProfile != "" {
			log.Info("Stopping profiler")
			pprof.StopCPUProfile()
		}
	}()
}
