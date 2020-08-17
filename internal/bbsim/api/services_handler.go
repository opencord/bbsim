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
)

func convertBBSimServiceToProtoService(s *devices.Service) *bbsim.Service {
	return &bbsim.Service{
		Name:       s.Name,
		HwAddress:  s.HwAddress.String(),
		OnuSn:      s.Onu.Sn(),
		CTag:       int32(s.CTag),
		STag:       int32(s.STag),
		NeedsEapol: s.NeedsEapol,
		NeedsDhcp:  s.NeedsDhcp,
		NeedsIgmp:  s.NeedsIgmp,
		GemPort:    int32(s.GemPort),
		EapolState: s.EapolState.Current(),
		DhcpState:  s.DHCPState.Current(),
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

func (s BBSimServer) GetServices(ctx context.Context, req *bbsim.Empty) (*bbsim.Services, error) {

	services := bbsim.Services{
		Items: []*bbsim.Service{},
	}

	olt := devices.GetOLT()

	for _, pon := range olt.Pons {
		for _, o := range pon.Onus {
			s := convertBBsimServicesToProtoServices(o.Services)
			services.Items = append(services.Items, s...)
		}
	}

	return &services, nil
}

func (s BBSimServer) GetOnuServices(ctx context.Context, req *bbsim.ONURequest) (*bbsim.Services, error) {
	onu, err := s.GetONU(ctx, req)

	if err != nil {
		return nil, err
	}

	services := bbsim.Services{
		Items: onu.Services,
	}

	return &services, nil
}
