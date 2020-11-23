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

	dmi "github.com/opencord/device-management-interface/go/dmi"
)

//ListEvents lists the supported events for the passed device
func (dms *DmiAPIServer) ListEvents(ctx context.Context, req *dmi.HardwareID) (*dmi.ListEventsResponse, error) {
	logger.Debugf("ListEvents called with request %+v", req)
	//empty events
	events := []*dmi.EventCfg{{}}
	return &dmi.ListEventsResponse{
		Status: dmi.Status_OK_STATUS,
		Reason: dmi.Reason_UNDEFINED_REASON,
		Events: &dmi.EventsCfg{
			Items: events,
		},
	}, nil
}

//UpdateEventsConfiguration updates the configuration of the list of events in the request
func (dms *DmiAPIServer) UpdateEventsConfiguration(ctx context.Context, req *dmi.EventsConfigurationRequest) (*dmi.EventsConfigurationResponse, error) {
	logger.Debugf("UpdateEventsConfiguration called with request %+v", req)
	return &dmi.EventsConfigurationResponse{
		Status: dmi.Status_OK_STATUS,
	}, nil
}
