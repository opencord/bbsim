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
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/opencord/bbsim/internal/bbsim/devices"
	"github.com/opencord/bbsim/internal/common"
	dmi "github.com/opencord/device-management-interface/go/dmi"

	guuid "github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func getUUID(seed string) string {
	return guuid.NewMD5(guuid.Nil, []byte(seed)).String()
}

//StartManagingDevice establishes connection with the device and does checks to ascertain if the device with passed identity can be managed
func (dms *DmiAPIServer) StartManagingDevice(req *dmi.ModifiableComponent, stream dmi.NativeHWManagementService_StartManagingDeviceServer) error {
	//Get serial number and generate the UUID based on this serial number. Store this UUID in local cache
	logger.Debugf("StartManagingDevice() invoked with request  %+v", req)
	if req == nil {
		return status.Errorf(codes.FailedPrecondition, "request is empty")
	}

	if req.Name == "" {
		return status.Errorf(codes.InvalidArgument, "'Name' can not be empty in the request")
	}

	olt := devices.GetOLT()

	// Uri is the IP address
	dms.ipAddress = req.GetUri().GetUri()
	dms.deviceSerial = olt.SerialNumber
	dms.deviceName = fmt.Sprintf("%s-%s", common.Config.Olt.Vendor, dms.deviceSerial)

	dms.uuid = getUUID(dms.deviceSerial)

	dms.ponTransceiverUuids = make([]string, olt.NumPon)
	dms.ponTransceiverCageUuids = make([]string, olt.NumPon)

	var components []*dmi.Component

	// Create and store the component for transceivers and transceiver cages
	for i := 0; i < olt.NumPon; i++ {
		label := fmt.Sprintf("pon-%d", olt.Pons[i].ID)
		dms.ponTransceiverUuids[i] = getUUID(dms.deviceSerial + label)
		dms.ponTransceiverCageUuids[i] = getUUID(dms.deviceSerial + "cage" + label)

		transName := fmt.Sprintf("sfp-%d", i)
		cageName := fmt.Sprintf("cage-%d", i)

		trans := dmi.Component{
			Name:        transName,
			Class:       dmi.ComponentType_COMPONENT_TYPE_TRANSCEIVER,
			Description: "XGS-PON",
			Uuid: &dmi.Uuid{
				Uuid: dms.ponTransceiverUuids[i],
			},
			Parent: cageName,
		}

		cage := dmi.Component{
			Name:        cageName,
			Class:       dmi.ComponentType_COMPONENT_TYPE_CONTAINER,
			Description: "cage",
			Uuid: &dmi.Uuid{
				Uuid: dms.ponTransceiverCageUuids[i],
			},
			Parent:   dms.deviceName,
			Children: []*dmi.Component{&trans},
		}

		components = append(components, &cage)
	}

	dms.components = components

	logger.Debugf("Generated UUID for the uri %s is %s", dms.ipAddress, dms.uuid)
	response := &dmi.StartManagingDeviceResponse{
		Status: dmi.Status_OK,
		DeviceUuid: &dmi.Uuid{
			Uuid: dms.uuid,
		},
	}

	err := stream.Send(response)
	if err != nil {
		logger.Errorf("Error while sending response to client %v", err.Error())
		return status.Errorf(codes.Unknown, err.Error())
	}

	return nil
}

//StopManagingDevice stops management of a device and cleans up any context and caches for that device
func (dms *DmiAPIServer) StopManagingDevice(context.Context, *dmi.StopManagingDeviceRequest) (*dmi.StopManagingDeviceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "rpc StopManagingDevice not implemented")
}

//GetPhysicalInventory gets the HW inventory details of the Device
func (dms *DmiAPIServer) GetPhysicalInventory(req *dmi.PhysicalInventoryRequest, stream dmi.NativeHWManagementService_GetPhysicalInventoryServer) error {
	if req == nil || req.DeviceUuid == nil || req.DeviceUuid.Uuid == "" {
		return status.Errorf(codes.InvalidArgument, "device-UUID missing in the request")
	}

	// Function to send the response back on the stream
	sendResponseBackOnStream := func(stream dmi.NativeHWManagementService_GetPhysicalInventoryServer, msg *dmi.PhysicalInventoryResponse) error {
		err := stream.Send(msg)
		if err != nil {
			logger.Errorf("Error sending response to client, error: %v", err)
			return status.Errorf(codes.Internal, "Error sending response to client "+err.Error())
		}
		return nil
	}

	if req.DeviceUuid.Uuid != dms.uuid {
		logger.Errorf("Requested uuid =%s, uuid of existing device = %s", req.DeviceUuid.Uuid, dms.uuid)
		// Wrong uuid, return error
		errResponse := &dmi.PhysicalInventoryResponse{
			Status:    dmi.Status_ERROR,
			Reason:    dmi.Reason_UNKNOWN_DEVICE,
			Inventory: &dmi.Hardware{},
		}

		return sendResponseBackOnStream(stream, errResponse)
	}

	response := &dmi.PhysicalInventoryResponse{
		Status: dmi.Status_OK,
		Inventory: &dmi.Hardware{
			LastChange: &timestamp.Timestamp{
				Seconds: 0,
				Nanos:   0,
			},
			Root: &dmi.Component{
				Name:         dms.deviceName,
				Class:        0,
				Description:  "",
				Parent:       "",
				ParentRelPos: 0,
				Children:     dms.components,
				SerialNum:    dms.deviceSerial,
				MfgName:      common.Config.Olt.Vendor,
				IsFru:        false,
				Uri: &dmi.Uri{
					Uri: dms.ipAddress,
				},
				Uuid: &dmi.Uuid{
					Uuid: dms.uuid,
				},
				State: &dmi.ComponentState{},
			},
		},
	}
	return sendResponseBackOnStream(stream, response)
}

//Contains tells whether arr contains element.
func Contains(arr []string, element string) bool {
	for _, item := range arr {
		logger.Debugf("Checking in Contains slice elem = %v str = %s", item, element)
		if element == item {
			return true
		}
	}
	return false
}

func findComponent(l []*dmi.Component, compUUID string) *dmi.Component {
	for _, comp := range l {
		logger.Debugf("findComponent slice comp = %v compUUID = %s", comp, compUUID)
		if comp.Uuid.Uuid == compUUID {
			return comp
		}

		for _, child := range comp.GetChildren() {
			logger.Debugf("findComponent Child slice comp = %v compUUID = %s", comp, compUUID)
			if child.Uuid.Uuid == compUUID {
				return child
			}
		}
	}

	return nil
}

func sendGetHWComponentResponse(c *dmi.Component, stream dmi.NativeHWManagementService_GetHWComponentInfoServer) error {
	apiStatus := dmi.Status_OK
	reason := dmi.Reason_UNDEFINED_REASON

	if c == nil {
		apiStatus = dmi.Status_ERROR
		reason = dmi.Reason_UNKNOWN_DEVICE
	}

	response := &dmi.HWComponentInfoGetResponse{
		Status:    apiStatus,
		Reason:    reason,
		Component: c,
	}

	err := stream.Send(response)
	if err != nil {
		logger.Errorf("Error sending response to client, error: %v", err)
		return status.Errorf(codes.Internal, "Error sending response to client "+err.Error())
	}
	return nil
}

//GetHWComponentInfo gets the details of a particular HW component
func (dms *DmiAPIServer) GetHWComponentInfo(req *dmi.HWComponentInfoGetRequest, stream dmi.NativeHWManagementService_GetHWComponentInfoServer) error {
	logger.Debugf("GetHWComponentInfo() invoked with request %+v", req)

	if req == nil {
		return status.Errorf(codes.FailedPrecondition, "can not entertain nil request")
	}
	if stream == nil {
		logger.Errorf("stream to send is nil, not sending response from gRPC server ")
		return status.Errorf(codes.Internal, "stream to send is nil, can not send response from gRPC server")
	}

	componentFound := Contains(dms.ponTransceiverUuids, req.ComponentUuid.Uuid)
	if !componentFound {
		componentFound = Contains(dms.ponTransceiverCageUuids, req.ComponentUuid.Uuid)
	}

	if req.DeviceUuid.Uuid != dms.uuid || !componentFound {
		// Wrong uuid, return error
		return sendGetHWComponentResponse(nil, stream)
	}

	// Search for the component and return it
	c := findComponent(dms.components, req.ComponentUuid.Uuid)
	return sendGetHWComponentResponse(c, stream)
}

//SetHWComponentInfo sets the permissible attributes of a HW component
func (dms *DmiAPIServer) SetHWComponentInfo(context.Context, *dmi.HWComponentInfoSetRequest) (*dmi.HWComponentInfoSetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "rpc SetHWComponentInfo not implemented")
}

//SetLoggingEndpoint sets the location to which logs need to be shipped
func (dms *DmiAPIServer) SetLoggingEndpoint(context.Context, *dmi.SetLoggingEndpointRequest) (*dmi.SetRemoteEndpointResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "rpc SetLoggingEndpoint not implemented")
}

//GetLoggingEndpoint gets the configured location to which the logs are being shipped
func (dms *DmiAPIServer) GetLoggingEndpoint(context.Context, *dmi.Uuid) (*dmi.GetLoggingEndpointResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "rpc GetLoggingEndpoint not implemented")
}

//SetMsgBusEndpoint sets the location of the Message Bus to which events and metrics are shipped
func (dms *DmiAPIServer) SetMsgBusEndpoint(context.Context, *dmi.SetMsgBusEndpointRequest) (*dmi.SetRemoteEndpointResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "rpc SetMsgBusEndpoint not implemented")
}

//GetMsgBusEndpoint gets the configured location to which the events and metrics are being shipped
func (dms *DmiAPIServer) GetMsgBusEndpoint(context.Context, *empty.Empty) (*dmi.GetMsgBusEndpointResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "rpc GetMsgBusEndpoint not implemented")
}

//GetManagedDevices returns an object containing a list of devices managed by this entity
func (dms *DmiAPIServer) GetManagedDevices(context.Context, *empty.Empty) (*dmi.ManagedDevicesResponse, error) {
	retResponse := dmi.ManagedDevicesResponse{}
	//If our uuid is empty, we return empty list; else we fill details and return
	if dms.uuid != "" {
		root := dmi.ModifiableComponent{
			Name: dms.deviceName,
			Uri: &dmi.Uri{
				Uri: dms.ipAddress,
			},
		}

		retResponse.Devices = append(retResponse.Devices, &root)
	}

	return &retResponse, nil
}
