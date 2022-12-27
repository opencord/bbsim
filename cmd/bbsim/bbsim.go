/*
* Copyright 2018-2023 Open Networking Foundation (ONF) and the ONF Contributors

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

	"github.com/opencord/bbsim/internal/bbsim/responders/webserver"

	"github.com/Shopify/sarama"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/api/legacy"
	"github.com/opencord/bbsim/internal/bbsim/api"
	"github.com/opencord/bbsim/internal/bbsim/devices"
	"github.com/opencord/bbsim/internal/bbsim/dmiserver"
	"github.com/opencord/bbsim/internal/common"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func startApiServer(apiDoneChannel chan bool, group *sync.WaitGroup) {
	address := common.Config.BBSim.ApiAddress
	log.Debugf("APIServer listening on %v", address)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("APIServer failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	bbsim.RegisterBBSimServer(grpcServer, api.BBSimServer{})

	reflection.Register(grpcServer)

	go func() { _ = grpcServer.Serve(lis) }()
	go startApiRestServer(apiDoneChannel, group, address)

	x := <-apiDoneChannel
	if x {
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

	address := common.Config.BBSim.RestApiAddress

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

	x := <-apiDoneChannel
	if x {
		log.Warnf("Stopping API REST server")
		_ = s.Shutdown(ctx)
	}

	group.Done()
}

// This server aims to provide compatibility with the previous BBSim version. It is deprecated and will be removed in the future.
func startLegacyApiServer(apiDoneChannel chan bool, group *sync.WaitGroup) {
	grpcAddress := common.Config.BBSim.LegacyApiAddress
	restAddress := common.Config.BBSim.LegacyRestApiAddress

	log.Debugf("Legacy APIServer listening on %v", grpcAddress)
	listener, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		log.Fatalf("Legacy APIServer failed to listen: %v", err)
		return
	}
	apiserver := grpc.NewServer()
	legacy.RegisterBBSimServiceServer(apiserver, api.BBSimLegacyServer{})

	go func() { _ = apiserver.Serve(listener) }()
	// Start rest gateway for BBSim server
	go api.StartRestGatewayService(apiDoneChannel, group, grpcAddress, restAddress)

	x := <-apiDoneChannel
	if x {
		// if the API channel is closed, stop the gRPC server
		log.Warnf("Stopping legacy API gRPC server")
		apiserver.Stop()
	}

	group.Done()
}

func main() {

	common.LoadConfig()

	common.SetLogLevel(log.StandardLogger(), common.Config.BBSim.LogLevel, common.Config.BBSim.LogCaller)

	if *common.Config.BBSim.CpuProfile != "" {
		// start profiling
		log.Infof("Creating profile file at: %s", *common.Config.BBSim.CpuProfile)
		f, err := os.Create(*common.Config.BBSim.CpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		_ = pprof.StartCPUProfile(f)
	}

	log.WithFields(log.Fields{
		"OltID":                       common.Config.Olt.ID,
		"NumNniPerOlt":                common.Config.Olt.NniPorts,
		"NumPonPerOlt":                common.Config.Olt.PonPorts,
		"NumOnuPerPon":                common.Config.Olt.OnusPonPort,
		"PonConfiguration":            *common.PonsConfig,
		"TotalOnus":                   common.Config.Olt.PonPorts * common.Config.Olt.OnusPonPort,
		"NniDhcpTrapVid":              common.Config.Olt.NniDhcpTrapVid,
		"Delay":                       common.Config.BBSim.Delay,
		"Events":                      common.Config.BBSim.Events,
		"KafkaEventTopic":             common.Config.BBSim.KafkaEventTopic,
		"ControlledActivation":        common.Config.BBSim.ControlledActivation,
		"EnablePerf":                  common.Config.BBSim.EnablePerf,
		"DhcpRetry":                   common.Config.BBSim.DhcpRetry,
		"AuthRetry":                   common.Config.BBSim.AuthRetry,
		"OltRebootDelay":              common.Config.Olt.OltRebootDelay,
		"OmciResponseRate":            common.Config.Olt.OmciResponseRate,
		"injectOmciUnknownMe":         common.Config.BBSim.InjectOmciUnknownMe,
		"injectOmciUnknownAttributes": common.Config.BBSim.InjectOmciUnknownAttributes,
		"omccVersion":                 common.Config.BBSim.OmccVersion,
	}).Info("BroadBand Simulator is on")

	// control channels, they are only closed when the goroutine needs to be terminated
	apiDoneChannel := make(chan bool)

	olt := devices.CreateOLT(
		*common.Config,
		common.Services,
		false,
	)

	log.Debugf("Created OLT with id: %d", common.Config.Olt.ID)

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
	if common.Config.BBSim.SadisServer {
		wg.Add(1)
		go webserver.StartRestServer(olt, &wg)
	}

	dms, dmserr := dmiserver.StartDmiAPIServer()
	if dmserr != nil {
		log.Errorf("Failed to start Device Management Interface Server %v", dmserr)
	}

	if common.Config.BBSim.Events {
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

		dms.Stop()

		if *common.Config.BBSim.CpuProfile != "" {
			log.Info("Stopping profiler")
			pprof.StopCPUProfile()
		}
	}()
}
