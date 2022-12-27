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
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	dmi "github.com/opencord/device-management-interface/go/dmi"
)

//ListEvents lists the supported events for the passed device
func (dms *DmiAPIServer) ListEvents(ctx context.Context, req *dmi.HardwareID) (*dmi.ListEventsResponse, error) {
	logger.Debugf("ListEvents called with request %+v", req)
	events := getEventsList()

	return &dmi.ListEventsResponse{
		Status: dmi.Status_OK_STATUS,
		Events: &dmi.EventsCfg{
			Items: events,
		},
	}, nil
}

//UpdateEventsConfiguration updates the configuration of the list of events in the request
func (dms *DmiAPIServer) UpdateEventsConfiguration(ctx context.Context, req *dmi.EventsConfigurationRequest) (*dmi.EventsConfigurationResponse, error) {
	logger.Debugf("UpdateEventsConfiguration called with request %+v", req)

	if req == nil || req.Operation == nil {
		return &dmi.EventsConfigurationResponse{
			Status: dmi.Status_ERROR_STATUS,
			//TODO reason must be INVALID_PARAMS, currently this is not available in Device Management interface (DMI),
			// change below reason with type INVALID_PARAMS once DMI is updated
			Reason: dmi.EventsConfigurationResponse_UNDEFINED_REASON,
		}, status.Errorf(codes.FailedPrecondition, "request is nil")
	}

	switch x := req.Operation.(type) {
	case *dmi.EventsConfigurationRequest_Changes:
		for _, eventConfig := range x.Changes.Items {
			UpdateEventConfig(eventConfig)
		}
	case *dmi.EventsConfigurationRequest_ResetToDefault:
		logger.Debugf("To be implemented later")
	case nil:
		// The field is not set.
		logger.Debugf("Update request operation type is nil")
		return &dmi.EventsConfigurationResponse{
			Status: dmi.Status_UNDEFINED_STATUS,
		}, nil
	}

	return &dmi.EventsConfigurationResponse{
		Status: dmi.Status_OK_STATUS,
	}, nil
}

// Initiates the server streaming of the events
func (dms *DmiAPIServer) StreamEvents(req *empty.Empty, srv dmi.NativeEventsManagementService_StreamEventsServer) error {
	return status.Errorf(codes.Unimplemented, "rpc StreamEvents not implemented")
}
