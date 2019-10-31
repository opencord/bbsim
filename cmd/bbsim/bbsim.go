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

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/api/legacy"
	"github.com/opencord/bbsim/internal/bbsim/api"
	"github.com/opencord/bbsim/internal/bbsim/devices"
	"github.com/opencord/bbsim/internal/common"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func startApiServer(channel chan bool, group *sync.WaitGroup) {
	// TODO make configurable
	address := "0.0.0.0:50070"
	log.Debugf("APIServer listening on: %v", address)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("APIServer failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	bbsim.RegisterBBSimServer(grpcServer, api.BBSimServer{})

	reflection.Register(grpcServer)

	go grpcServer.Serve(lis)
	go startApiRestServer(channel, group, address)

	select {
	case <-channel:
		// if the api channel is closed, stop the gRPC server
		grpcServer.Stop()
		log.Warnf("Stopping API gRPC server")
	}

	group.Done()
}

// startApiRestServer method starts the REST server (grpc gateway) for BBSim.
func startApiRestServer(channel chan bool, group *sync.WaitGroup, grpcAddress string) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// TODO make configurable
	address := "0.0.0.0:50071"

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	if err := bbsim.RegisterBBSimHandlerFromEndpoint(ctx, mux, grpcAddress, opts); err != nil {
		log.Errorf("Could not register API server: %v", err)
		return
	}

	s := &http.Server{Addr: address, Handler: mux}

	go func() {
		log.Infof("REST API server listening on %s ...", address)
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Errorf("Could not start API server: %v", err)
			return
		}
	}()

	select {
	case <-channel:
		log.Warnf("Stopping API REST server")
		s.Shutdown(ctx)
	}

	group.Done()
}

// This server aims to provide compatibility with the previous BBSim version. It is deprecated and will be removed in the future.
func startLegacyApiServer(channel chan bool, group *sync.WaitGroup) {
	// TODO make configurable
	grpcAddress := "0.0.0.0:50072"
	restAddress := "0.0.0.0:50073"

	log.Debugf("Legacy APIServer listening on: %v", grpcAddress)
	listener, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		log.Fatalf("Legacy APIServer failed to listen: %v", err)
		return
	}
	apiserver := grpc.NewServer()
	legacy.RegisterBBSimServiceServer(apiserver, api.BBSimLegacyServer{})

	go apiserver.Serve(listener)
	// Start rest gateway for BBSim server
	go api.StartRestGatewayService(channel, group, grpcAddress, restAddress)

	select {
	case <-channel:
		// if the olt channel is closed, stop the gRPC server
		log.Warnf("Stopping legacy API gRPC server")
		apiserver.Stop()
		break

	}

	group.Done()
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
		"TotalOnus":    options.NumPonPerOlt * options.NumOnuPerPon,
		"Auth":         options.Auth,
		"Dhcp":         options.Dhcp,
		"Delay":    options.Delay,
	}).Info("BroadBand Simulator is on")

	// control channels, they are only closed when the goroutine needs to be terminated
	oltDoneChannel := make(chan bool)
	apiDoneChannel := make(chan bool)

	sigs := make(chan os.Signal, 1)
	// stop API and OLT servers on SIGTERM
	signal.Notify(sigs, syscall.SIGTERM)

	go func() {
		<-sigs
		// TODO check when these servers should be shutdown
		close(apiDoneChannel)
		close(oltDoneChannel)
	}()


	wg := sync.WaitGroup{}
	wg.Add(5)

	olt := devices.CreateOLT(
		options.OltID,
		options.NumNniPerOlt,
		options.NumPonPerOlt,
		options.NumOnuPerPon,
		options.STag,
		options.CTagInit,
		&oltDoneChannel,
		&apiDoneChannel,
		options.Auth,
		options.Dhcp,
		options.Delay,
		false,
	)

	go devices.StartOlt(olt, &wg)
	log.Debugf("Created OLT with id: %d", options.OltID)
	go startApiServer(apiDoneChannel, &wg)
	go startLegacyApiServer(apiDoneChannel, &wg)

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
