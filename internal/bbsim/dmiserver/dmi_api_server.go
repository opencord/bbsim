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

package dmiserver

import (
	"context"
	"net"

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
	deviceSerial            string
	deviceName              string
	ipAddress               string
	uuid                    string
	ponTransceiverUuids     []string
	ponTransceiverCageUuids []string
	root                    *dmi.Component
	metricChannel           chan interface{}
	kafkaEndpoint           string
	mPublisherCancelFunc    context.CancelFunc
}

var dmiServ DmiAPIServer

//StartDmiAPIServer starts a new grpc server for the Device Manager Interface
func StartDmiAPIServer() (*grpc.Server, error) {
	dmiServ = DmiAPIServer{}

	return dmiServ.newDmiAPIServer()
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

	reflection.Register(grpcServer)

	go func() { _ = grpcServer.Serve(lis) }()
	logger.Debugf("DMI grpc Server listening on %v", address)

	return grpcServer, nil
}
