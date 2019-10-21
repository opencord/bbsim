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
	"runtime/pprof"
	"sync"

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

	wg := sync.WaitGroup{}
	wg.Add(1)

	go grpcServer.Serve(lis)
	log.Debug("here1")

	go startApiRestServer(address)

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

// startApiRestServer method starts the grpc gateway REST server for BBSim.
func startApiRestServer(grpcAddress string) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	address := "0.0.0.0:50071"
	log.Debug("here2")

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	err := bbsim.RegisterBBSimHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
	if err != nil {
		log.Error("%v", err)
		return
	}

	log.Infof("REST API server listening on %s ...", address)
	err = http.ListenAndServe(address, mux)
	if err != nil {
		log.Error("%v", err)
		return
	}
	return
}

func startLegacyApiServer() {
	defer log.Debug("Management API server exited")

	grpcAddressPort := "0.0.0.0:50072"
	restAddressPort := "0.0.0.0:50073"

	// Start rest gateway for BBSim server
	go api.StartRestGatewayService(grpcAddressPort, restAddressPort)

	listener, apiserver, err := api.NewGRPCServer(grpcAddressPort)
	if err != nil {
		log.Errorf("Unable to create management api server %v", err)
		return
	}

	legacy.RegisterBBSimServiceServer(apiserver, api.BBSimLegacyServer{})
	if e := apiserver.Serve(listener); e != nil {
		log.Errorf("Failed to run management api server %v", e)
		return
	}
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

	go startLegacyApiServer()

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
