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

	"github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsim/devices"
	"github.com/opencord/bbsim/internal/common"
	log "github.com/sirupsen/logrus"
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

	res := bbsim.Olt{
		ID:            int32(olt.ID),
		SerialNumber:  olt.SerialNumber,
		OperState:     olt.OperState.Current(),
		InternalState: olt.InternalState.Current(),
		NNIPorts:      nnis,
		PONPorts:      pons,
	}
	return &res, nil
}

func (s BBSimServer) SetLogLevel(ctx context.Context, req *bbsim.LogLevel) (*bbsim.LogLevel, error) {

	common.SetLogLevel(log.StandardLogger(), req.Level, req.Caller)

	return &bbsim.LogLevel{
		Level:  log.StandardLogger().Level.String(),
		Caller: log.StandardLogger().ReportCaller,
	}, nil
}
