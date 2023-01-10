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

package dmiserver

import (
	"context"
	"fmt"
	"net"

	"github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsim/devices"
	"github.com/opencord/bbsim/internal/common"
	dmi "github.com/opencord/device-management-interface/go/dmi"
	log "github.com/sirupsen/logrus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var logger = log.WithFields(log.Fields{
	"module": "DmiServer",
})

//DmiAPIServer has the attributes for the Server handling the Device Management Interface
type DmiAPIServer struct {
	ipAddress            string
	uuid                 *dmi.Uuid
	root                 *dmi.Component
	Transceivers         []*Transceiver
	metricChannel        chan interface{}
	eventChannel         chan interface{}
	kafkaEndpoint        string
	loggingEndpoint      string
	loggingProtocol      string
	mPublisherCancelFunc context.CancelFunc
}

var dmiServ DmiAPIServer

//StartDmiAPIServer starts a new grpc server for the Device Manager Interface
func StartDmiAPIServer() (*grpc.Server, error) {
	dmiServ = DmiAPIServer{}

	// Create the mapping between transceivers and PONS
	// TODO: at the moment we create one transceiver for each PON,
	// but in the case of COMBO PON this will not always be the case.
	// This should be expanded to cover that scenario
	logger.Debug("Creating transceivers from DMI")
	dmiServ.Transceivers = []*Transceiver{}
	for _, pon := range devices.GetOLT().Pons {
		trans := newTransceiver(pon.ID, []*devices.PonPort{pon})
		dmiServ.Transceivers = append(dmiServ.Transceivers, trans)
	}

	return dmiServ.newDmiAPIServer()
}

func getDmiAPIServer() (*DmiAPIServer, error) {
	if dmiServ.root == nil {
		return nil, fmt.Errorf("Device management not started")
	}
	return &dmiServ, nil
}

// newDmiAPIServer launches a new grpc server for the Device Manager Interface
func (dms *DmiAPIServer) newDmiAPIServer() (*grpc.Server, error) {
	address := common.Config.BBSim.DmiServerAddress
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer()

	dmi.RegisterNativeHWManagementServiceServer(grpcServer, dms)
	dmi.RegisterNativeSoftwareManagementServiceServer(grpcServer, dms)
	dmi.RegisterNativeEventsManagementServiceServer(grpcServer, dms)
	dmi.RegisterNativeMetricsManagementServiceServer(grpcServer, dms)
	bbsim.RegisterBBsimDmiServer(grpcServer, dms)

	reflection.Register(grpcServer)

	go func() { _ = grpcServer.Serve(lis) }()
	logger.Debugf("DMI grpc Server listening on %v", address)

	return grpcServer, nil
}
