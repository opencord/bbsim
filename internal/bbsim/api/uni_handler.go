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

	"github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsim/devices"
)

func convertBBSimUniPortToProtoUniPort(u *devices.UniPort) *bbsim.UNI {
	return &bbsim.UNI{
		ID:        int32(u.ID),
		OnuID:     int32(u.Onu.ID),
		OnuSn:     u.Onu.Sn(),
		MeID:      uint32(u.MeId.ToUint16()),
		PortNo:    int32(u.PortNo),
		OperState: u.OperState.Current(),
		Services:  convertBBsimServicesToProtoServices(u.Services),
		Type:      bbsim.UniType_ETH,
	}
}

func convertBBSimPotsPortToProtoUniPort(u *devices.PotsPort) *bbsim.UNI {
	return &bbsim.UNI{
		ID:        int32(u.ID),
		OnuID:     int32(u.Onu.ID),
		OnuSn:     u.Onu.Sn(),
		MeID:      uint32(u.MeId.ToUint16()),
		PortNo:    int32(u.PortNo),
		OperState: u.OperState.Current(),
		Type:      bbsim.UniType_POTS,
	}
}

func convertBBsimUniPortsToProtoUniPorts(list []devices.UniPortIf) []*bbsim.UNI {
	unis := []*bbsim.UNI{}
	for _, u := range list {
		uni := u.(*devices.UniPort)
		unis = append(unis, convertBBSimUniPortToProtoUniPort(uni))
	}
	return unis
}

func convertBBsimPotsPortsToProtoUniPorts(list []devices.PotsPortIf) []*bbsim.UNI {
	unis := []*bbsim.UNI{}
	for _, u := range list {
		uni := u.(*devices.PotsPort)
		unis = append(unis, convertBBSimPotsPortToProtoUniPort(uni))
	}
	return unis
}

func (s BBSimServer) GetOnuUnis(ctx context.Context, req *bbsim.ONURequest) (*bbsim.UNIs, error) {
	onu, err := s.GetONU(ctx, req)

	if err != nil {
		return nil, err
	}

	unis := bbsim.UNIs{
		Items: onu.Unis,
	}

	return &unis, nil
}
