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

package api

import (
	"context"
	"net/http"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/opencord/bbsim/api/legacy"
	"github.com/opencord/bbsim/internal/bbsim/devices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BBSimLegacyServer struct {
}

// Response Constants
const (
	RequestAccepted = "API request accepted"
	OLTNotEnabled   = "OLT not enabled"
	RequestFailed   = "API request failed"
	SoftReboot      = "soft-reboot"
	HardReboot      = "hard-reboot"
	DeviceTypeOlt   = "olt"
	DeviceTypeOnu   = "onu"
)

// OLTStatus method returns the OLT status.
func (s BBSimLegacyServer) OLTStatus(ctx context.Context, in *legacy.Empty) (*legacy.OLTStatusResponse, error) {
	logger.Trace("OLTStatus request received")

	olt := devices.GetOLT()
	oltInfo := &legacy.OLTStatusResponse{
		Olt: &legacy.OLTInfo{
			OltId:     int64(olt.ID),
			OltVendor: "BBSIM",
			OltSerial: olt.SerialNumber,
			//OltIp:     getOltIP().String(),
			OltState: olt.OperState.Current(),
		},
	}

	for _, nni := range olt.Nnis {
		nniPortInfo := &legacy.PortInfo{
			PortType:  "nni",
			PortId:    uint32(nni.ID),
			PortState: nni.OperState.Current(),
		}
		oltInfo.Ports = append(oltInfo.Ports, nniPortInfo)
	}

	for _, pon := range olt.Pons {
		ponPortInfo := &legacy.PortInfo{
			PortType:          "pon",
			PortId:            uint32(pon.ID),
			PonPortMaxOnus:    uint32(olt.NumOnuPerPon),
			PonPortActiveOnus: uint32(olt.NumPon),
			PortState:         pon.OperState.Current(),
		}
		oltInfo.Ports = append(oltInfo.Ports, ponPortInfo)
	}

	return oltInfo, nil
}

// PortStatus method returns Port status.
func (s BBSimLegacyServer) PortStatus(ctx context.Context, in *legacy.PortInfo) (*legacy.Ports, error) {
	logger.Trace("PortStatus() invoked")
	ports := &legacy.Ports{}
	switch portType := in.GetPortType(); portType {
	case "pon":
		portInfo, err := s.fetchPortDetail(in.PortId, portType)
		if err != nil {
			return ports, err
		}
		ports.Ports = append(ports.Ports, portInfo)
	case "nni":
		portInfo, _ := s.fetchPortDetail(in.PortId, portType)
		ports.Ports = append(ports.Ports, portInfo)
	default:
		return &legacy.Ports{}, status.Errorf(codes.InvalidArgument, "Invalid port type")
	}

	return ports, nil
}

// ONUStatus method returns ONU status.
func (s BBSimLegacyServer) ONUStatus(ctx context.Context, in *legacy.ONURequest) (*legacy.ONUs, error) {
	logger.Trace("ONUStatus request received")
	onuInfo := &legacy.ONUs{}

	if in.GetOnu() != nil {
		logger.Debugf("Received single ONU: %+v, %d\n", in.GetOnu().OnuId, in.GetOnu().PonPortId)
		return s.handleONUStatusRequest(in.GetOnu())
	}

	logger.Debug("Received all ONUs status request")

	// Get status of all ONUs
	olt := devices.GetOLT()
	for _, p := range olt.Pons {
		for _, o := range p.Onus {
			onuInfo.Onus = append(onuInfo.Onus, copyONUInfo(o))
		}
	}

	return onuInfo, nil
}

// ONUActivate method handles ONU activate requests from user.
func (s BBSimLegacyServer) ONUActivate(ctx context.Context, in *legacy.ONURequest) (*legacy.BBSimResponse, error) {
	logger.Trace("ONUActivate request received")
	logger.Error("Not implemented")

	return &legacy.BBSimResponse{StatusMsg: RequestAccepted}, nil
}

// ONUDeactivate method handles ONU deactivation request.
func (s BBSimLegacyServer) ONUDeactivate(ctx context.Context, in *legacy.ONURequest) (*legacy.BBSimResponse, error) {
	logger.Info("ONUDeactivate request received")
	logger.Error("Not implemented")

	return &legacy.BBSimResponse{StatusMsg: RequestAccepted}, nil
}

// GenerateONUAlarm RPC generates alarm for the onu
func (s BBSimLegacyServer) GenerateONUAlarm(ctx context.Context, in *legacy.ONUAlarmRequest) (*legacy.BBSimResponse, error) {
	logger.Trace("GenerateONUAlarms() invoked")
	logger.Error("Not implemented")

	return nil, nil
}

// GenerateOLTAlarm RPC generates alarm for the OLT
func (s BBSimLegacyServer) GenerateOLTAlarm(ctx context.Context, in *legacy.OLTAlarmRequest) (*legacy.BBSimResponse, error) {
	logger.Trace("GenerateOLTAlarm() invoked")
	logger.Error("Not implemented")

	return &legacy.BBSimResponse{StatusMsg: RequestAccepted}, nil
}

// PerformDeviceAction rpc take the device request and performs OLT and ONU hard and soft reboot
func (s BBSimLegacyServer) PerformDeviceAction(ctx context.Context, in *legacy.DeviceAction) (*legacy.BBSimResponse, error) {
	logger.Trace("PerformDeviceAction() invoked")
	logger.Error("Not implemented")

	return &legacy.BBSimResponse{StatusMsg: RequestAccepted}, nil
}

// GetFlows returns all flows or flows for specified ONU
func (s BBSimLegacyServer) GetFlows(ctx context.Context, in *legacy.ONUInfo) (*legacy.Flows, error) {
	logger.Info("GetFlow request received")
	logger.Error("Not implemented")

	return &legacy.Flows{}, nil
}

// StartRestGatewayService method starts REST server for BBSim.
func StartRestGatewayService(channel chan bool, group *sync.WaitGroup, grpcAddress string, hostandport string) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	// Register REST endpoints
	err := legacy.RegisterBBSimServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
	if err != nil {
		logger.Errorf("%v", err)
		return
	}

	s := &http.Server{Addr: hostandport, Handler: mux}

	go func() {
		logger.Infof("legacy REST API server listening on %s", hostandport)
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("Could not start legacy API server: %v", err)
			return
		}
	}()

	x := <-channel
	if x {
		logger.Warnf("Stopping legacy API REST server")
		if err := s.Shutdown(ctx); err != nil {
			logger.Errorf("Could not stop server: %v", err)
		}
		group.Done()
	}
}
