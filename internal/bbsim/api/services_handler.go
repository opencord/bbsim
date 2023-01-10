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
	"strconv"

	"github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsim/devices"
)

func convertBBSimServiceToProtoService(s *devices.Service) *bbsim.Service {
	return &bbsim.Service{
		Name:                s.Name,
		InternalState:       s.InternalState.Current(),
		HwAddress:           s.HwAddress.String(),
		OnuSn:               s.UniPort.Onu.Sn(),
		UniId:               s.UniPort.ID,
		UniTagMatch:         int32(s.UniTagMatch),
		CTag:                int32(s.CTag),
		UsCTagPriority:      uint32(s.UsPonCTagPriority),
		DsCTagPriority:      uint32(s.DsPonCTagPriority),
		STag:                int32(s.STag),
		UsSTagPriority:      uint32(s.UsPonSTagPriority),
		DsSTagPriority:      uint32(s.DsPonSTagPriority),
		NeedsEapol:          s.NeedsEapol,
		NeedsDhcp:           s.NeedsDhcp,
		NeedsIgmp:           s.NeedsIgmp,
		NeedsPPPoE:          s.NeedsPPPoE,
		ConfigureMacAddress: s.ConfigureMacAddress,
		EnableMacLearning:   s.EnableMacLearning,
		GemPort:             int32(s.GemPort),
		EapolState:          s.EapolState.Current(),
		DhcpState:           s.DHCPState.Current(),
		IGMPState:           s.IGMPState.Current(),
	}
}

func convertBBsimServicesToProtoServices(list []devices.ServiceIf) []*bbsim.Service {
	services := []*bbsim.Service{}
	for _, service := range list {
		s := service.(*devices.Service)
		services = append(services, convertBBSimServiceToProtoService(s))
	}
	return services
}

func (s BBSimServer) GetServices(ctx context.Context, req *bbsim.UNIRequest) (*bbsim.Services, error) {

	services := bbsim.Services{
		Items: []*bbsim.Service{},
	}

	olt := devices.GetOLT()

	for _, pon := range olt.Pons {
		for _, o := range pon.Onus {
			for _, u := range o.UniPorts {
				uni := u.(*devices.UniPort)
				s := convertBBsimServicesToProtoServices(uni.Services)
				for _, service := range s {
					intVar, err := strconv.Atoi(req.UniID)
					if req.UniID == "" && req.OnuSerialNumber == "" {
						services.Items = append(services.Items, service)
					} else if err == nil && service.UniId == uint32(intVar) && service.OnuSn == req.OnuSerialNumber {
						services.Items = append(services.Items, service)
					} else if req.UniID == "" && service.OnuSn == req.OnuSerialNumber {
						services.Items = append(services.Items, service)
					}
				}
			}
		}
	}

	return &services, nil
}
