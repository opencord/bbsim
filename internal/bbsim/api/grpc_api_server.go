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

package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsim/alarmsim"
	"github.com/opencord/bbsim/internal/bbsim/devices"
	"github.com/opencord/bbsim/internal/common"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
)

var logger = log.WithFields(log.Fields{
	"module": "GrpcApiServer",
})

var (
	version    string
	buildTime  string
	commitHash string
	gitStatus  string
)

type BBSimServer struct {
}

func (s BBSimServer) Version(ctx context.Context, req *bbsim.Empty) (*bbsim.VersionNumber, error) {
	// TODO add a flag to specify whether the tree was clean at this commit or not
	return &bbsim.VersionNumber{
		Version:    version,
		BuildTime:  buildTime,
		CommitHash: commitHash,
		GitStatus:  gitStatus,
	}, nil
}

func (s BBSimServer) GetOlt(ctx context.Context, req *bbsim.Empty) (*bbsim.Olt, error) {
	olt := devices.GetOLT()
	nnis := []*bbsim.NNIPort{}
	pons := []*bbsim.PONPort{}

	for _, nni := range olt.Nnis {
		n := bbsim.NNIPort{
			ID:        int32(nni.ID),
			OperState: nni.OperState.Current(),
		}
		nnis = append(nnis, &n)
	}

	for _, pon := range olt.Pons {
		p := bbsim.PONPort{
			ID:        int32(pon.ID),
			OperState: pon.OperState.Current(),
		}
		pons = append(pons, &p)
	}

	oltAddress := strings.Split(common.Config.BBSim.OpenOltAddress, ":")[0]
	if oltAddress == "" {
		oltAddress = getOltIP().String()
	}

	res := bbsim.Olt{
		ID:            int32(olt.ID),
		SerialNumber:  olt.SerialNumber,
		OperState:     olt.OperState.Current(),
		InternalState: olt.InternalState.Current(),
		IP:            oltAddress,
		NNIPorts:      nnis,
		PONPorts:      pons,
	}
	return &res, nil
}

func (s BBSimServer) PoweronOlt(ctx context.Context, req *bbsim.Empty) (*bbsim.Response, error) {
	res := &bbsim.Response{}
	o := devices.GetOLT()

	if err := o.InternalState.Event("initialize"); err != nil {
		log.Errorf("Error initializing OLT: %v", err)
		res.StatusCode = int32(codes.FailedPrecondition)
		return res, err
	}

	res.StatusCode = int32(codes.OK)
	return res, nil
}

func (s BBSimServer) ShutdownOlt(ctx context.Context, req *bbsim.Empty) (*bbsim.Response, error) {
	res := &bbsim.Response{}
	o := devices.GetOLT()

	if err := o.InternalState.Event("disable"); err != nil {
		log.Errorf("Error disabling OLT: %v", err)
		res.StatusCode = int32(codes.FailedPrecondition)
		return res, err
	}

	res.StatusCode = int32(codes.OK)
	return res, nil
}

func (s BBSimServer) RebootOlt(ctx context.Context, req *bbsim.Empty) (*bbsim.Response, error) {
	res := &bbsim.Response{}
	o := devices.GetOLT()
	go func() { _ = o.RestartOLT() }()
	res.StatusCode = int32(codes.OK)
	res.Message = fmt.Sprintf("OLT restart triggered.")
	return res, nil
}

func (s BBSimServer) StopgRPCServer(ctx context.Context, req *bbsim.Empty) (*bbsim.Response, error) {
	res := &bbsim.Response{}
	res.StatusCode = int32(codes.OK)
	res.Message = fmt.Sprintf("Openolt gRPC server stopped")
	o := devices.GetOLT()

	logger.Infof("Received request to stop Openolt gRPC Server")

	o.StopOltServer()

	return res, nil
}

func (s BBSimServer) StartgRPCServer(ctx context.Context, req *bbsim.Empty) (*bbsim.Response, error) {
	res := &bbsim.Response{}
	res.StatusCode = int32(codes.OK)
	res.Message = fmt.Sprintf("Openolt gRPC server started")
	o := devices.GetOLT()

	logger.Infof("Received request to start Openolt gRPC Server")

	_, err := o.StartOltServer()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s BBSimServer) RestartgRPCServer(ctx context.Context, req *bbsim.Timeout) (*bbsim.Response, error) {
	o := devices.GetOLT()
	logger.Infof("Received request to restart Openolt gRPC Server in %v seconds", req.Delay)
	o.StopOltServer()

	res := &bbsim.Response{}
	res.StatusCode = int32(codes.OK)
	res.Message = fmt.Sprintf("Openolt gRPC server stopped, restarting in %v", req.Delay)

	go func() {
		time.Sleep(time.Duration(req.Delay) * time.Second)
		_, err := o.StartOltServer()
		if err != nil {
			logger.WithFields(log.Fields{
				"err": err,
			}).Error("Cannot restart Openolt gRPC server")
		}
		logger.Infof("Openolt gRPC Server restarted after %v seconds", req.Delay)
	}()

	return res, nil
}

func (s BBSimServer) SetLogLevel(ctx context.Context, req *bbsim.LogLevel) (*bbsim.LogLevel, error) {

	common.SetLogLevel(log.StandardLogger(), req.Level, req.Caller)

	return &bbsim.LogLevel{
		Level:  log.StandardLogger().Level.String(),
		Caller: log.StandardLogger().ReportCaller,
	}, nil
}

func (s BBSimServer) SetOnuAlarmIndication(ctx context.Context, req *bbsim.ONUAlarmRequest) (*bbsim.Response, error) {
	o := devices.GetOLT()
	err := alarmsim.SimulateOnuAlarm(ctx, req, o)
	if err != nil {
		return nil, err
	}

	res := &bbsim.Response{}
	res.StatusCode = int32(codes.OK)
	res.Message = fmt.Sprintf("Onu Alarm Indication Sent.")
	return res, nil
}

// SetOltAlarmIndication generates OLT Alarms for LOS
func (s BBSimServer) SetOltAlarmIndication(ctx context.Context, req *bbsim.OLTAlarmRequest) (*bbsim.Response, error) {
	o := devices.GetOLT()
	err := alarmsim.SimulateOltAlarm(ctx, req, o)
	if err != nil {
		return nil, err
	}

	res := &bbsim.Response{}
	res.StatusCode = int32(codes.OK)
	res.Message = fmt.Sprintf("Olt Alarm Indication Sent.")
	return res, nil
}
