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
	"flag"
	"github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsim/devices"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"sync"
)

func getOpts() *CliOptions {

	olt_id := flag.Int("olt_id", 0, "Number of OLT devices to be emulated (default is 1)")
	nni := flag.Int("nni", 1, "Number of NNI ports per OLT device to be emulated (default is 1)")
	pon := flag.Int("pon", 1, "Number of PON ports per OLT device to be emulated (default is 1)")
	onu := flag.Int("onu", 1, "Number of ONU devices per PON port to be emulated (default is 1)")
	flag.Parse()

	o := new(CliOptions)

	o.OltID = int(*olt_id)
	o.NumNniPerOlt = int(*nni)
	o.NumPonPerOlt = int(*pon)
	o.NumOnuPerPon = int(*onu)

	return o
}

func startApiServer()  {
	// TODO make configurable
	address :=  "0.0.0.0:50070"
	log.Debugf("APIServer Listening on: %v", address)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("APIServer failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	bbsim.RegisterBBSimServer(grpcServer, BBSimServer{})

	reflection.Register(grpcServer)

	go grpcServer.Serve(lis)
}

func init() {
	log.SetLevel(log.DebugLevel)
	//log.SetLevel(log.TraceLevel)
	//log.SetReportCaller(true)
}

func main() {

	options := getOpts()

	log.WithFields(log.Fields{
		"OltID": options.OltID,
		"NumNniPerOlt": options.NumNniPerOlt,
		"NumPonPerOlt": options.NumPonPerOlt,
		"NumOnuPerPon": options.NumOnuPerPon,
	}).Info("BroadBand Simulator is on")

	wg := sync.WaitGroup{}
	wg.Add(2)


	go devices.CreateOLT(options.OltID, options.NumNniPerOlt, options.NumPonPerOlt, options.NumOnuPerPon)
	log.Debugf("Created OLT with id: %d", options.OltID)
	go startApiServer()
	log.Debugf("Started APIService")

	wg.Wait()

	defer func() {
		log.Info("BroadBand Simulator is off")
	}()
}