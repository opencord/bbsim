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

	"github.com/Shopify/sarama"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/opencord/bbsim/internal/bbsim/devices"
	"github.com/opencord/bbsim/internal/common"
	dmi "github.com/opencord/device-management-interface/go/dmi"

	guuid "github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	kafkaChannelSize = 100
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

	// Start device metrics generator
	dms.metricChannel = make(chan interface{}, kafkaChannelSize)
	StartMetricGenerator(dms)

	// Start device event generator
	dms.eventChannel = make(chan interface{}, kafkaChannelSize)
	StartEventsGenerator(dms)

	var components []*dmi.Component

	// Create and store the component for transceivers and transceiver cages
	for i := 0; i < olt.NumPon; i++ {
		label := fmt.Sprintf("pon-%d", olt.Pons[i].ID)
		dms.ponTransceiverUuids[i] = getUUID(dms.deviceSerial + label)
		dms.ponTransceiverCageUuids[i] = getUUID(dms.deviceSerial + "cage" + label)

		transName := fmt.Sprintf("sfp-%d", i)
		cageName := fmt.Sprintf("sfp-plus-transceiver-cage-pon-%d", i)

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

	// create the fans
	numFans := 2
	fans := make([]*dmi.Component, numFans)

	for i := 0; i < numFans; i++ {
		fans[i] = createFanComponent(i + 1)
	}
	components = append(components, fans...)

	// Create 1 disk, 1 processor, 1 ram, 1 temperature sensor and power supply unit
	components = append(components, createDiskComponent(0))
	components = append(components, createProcessorComponent(0))
	components = append(components, createMemoryComponent(0))
	components = append(components, createInnerSurroundingTempComponentSensor(0))
	components = append(components, createPowerSupplyComponent(0))

	// create the root component
	dms.root = &dmi.Component{
		Name:         dms.deviceName,
		Class:        0,
		Description:  "",
		Parent:       "",
		ParentRelPos: 0,
		Children:     components,
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
	}

	logger.Debugf("Generated UUID for the uri %s is %s", dms.ipAddress, dms.uuid)
	response := &dmi.StartManagingDeviceResponse{
		Status: dmi.Status_OK_STATUS,
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

func createFanComponent(fanIdx int) *dmi.Component {
	fanName := fmt.Sprintf("Thermal/Fans/System Fan/%d", fanIdx)
	fanSerial := fmt.Sprintf("bbsim-fan-serial-%d", fanIdx)
	return &dmi.Component{
		Name:         fanName,
		Class:        dmi.ComponentType_COMPONENT_TYPE_FAN,
		Description:  "bbsim-fan",
		Parent:       "",
		ParentRelPos: 0,
		SerialNum:    fanSerial,
		MfgName:      "bbsim-fan",
		IsFru:        false,
		Uuid: &dmi.Uuid{
			Uuid: getUUID(fanName),
		},
		State: &dmi.ComponentState{},
	}
}

func createProcessorComponent(cpuIdx int) *dmi.Component {
	cpuName := fmt.Sprintf("Systems/1/Processors/%d", cpuIdx)
	cpuSerial := fmt.Sprintf("bbsim-cpu-serial-%d", cpuIdx)
	return &dmi.Component{
		Name:         cpuName,
		Class:        dmi.ComponentType_COMPONENT_TYPE_CPU,
		Description:  "bbsim-cpu",
		Parent:       "",
		ParentRelPos: 0,
		SerialNum:    cpuSerial,
		MfgName:      "bbsim-cpu",
		IsFru:        false,
		Uuid: &dmi.Uuid{
			Uuid: getUUID(cpuName),
		},
		State: &dmi.ComponentState{},
	}
}

func createMemoryComponent(memIdx int) *dmi.Component {
	memName := fmt.Sprintf("Systems/1/Memory/%d", memIdx)
	memSerial := fmt.Sprintf("bbsim-ram-serial-%d", memIdx)
	return &dmi.Component{
		Name:         memName,
		Class:        dmi.ComponentType_COMPONENT_TYPE_MEMORY,
		Description:  "bbsim-ram",
		Parent:       "",
		ParentRelPos: 0,
		SerialNum:    memSerial,
		MfgName:      "bbsim-ram",
		IsFru:        false,
		Uuid: &dmi.Uuid{
			Uuid: getUUID(memName),
		},
		State: &dmi.ComponentState{},
	}
}

func createDiskComponent(diskIdx int) *dmi.Component {
	diskName := fmt.Sprintf("Systems/1/Disk/%d", diskIdx)
	diskSerial := fmt.Sprintf("bbsim-disk-serial-%d", diskIdx)
	return &dmi.Component{
		Name:         diskName,
		Class:        dmi.ComponentType_COMPONENT_TYPE_STORAGE,
		Description:  "bbsim-disk",
		Parent:       "",
		ParentRelPos: 0,
		SerialNum:    diskSerial,
		MfgName:      "bbsim-disk",
		IsFru:        false,
		Uuid: &dmi.Uuid{
			Uuid: getUUID(diskName),
		},
		State: &dmi.ComponentState{},
	}
}

func createInnerSurroundingTempComponentSensor(sensorIdx int) *dmi.Component {
	sensorName := fmt.Sprintf("Systems/1/Sensor/%d", sensorIdx)
	sensorSerial := fmt.Sprintf("bbsim-sensor-istemp-serial-%d", sensorIdx)
	return &dmi.Component{
		Name:         sensorName,
		Class:        dmi.ComponentType_COMPONENT_TYPE_SENSOR,
		Description:  "bbsim-istemp",
		Parent:       "",
		ParentRelPos: 0,
		SerialNum:    sensorSerial,
		MfgName:      "bbsim-istemp",
		IsFru:        false,
		Uuid: &dmi.Uuid{
			Uuid: getUUID(sensorName),
		},
		State: &dmi.ComponentState{},
	}
}

func createPowerSupplyComponent(psuIdx int) *dmi.Component {
	psuName := fmt.Sprintf("Thermal/PSU/SystemPSU/%d", psuIdx)
	psuSerial := fmt.Sprintf("bbsim-psu-serial-%d", psuIdx)
	return &dmi.Component{
		Name:         psuName,
		Class:        dmi.ComponentType_COMPONENT_TYPE_POWER_SUPPLY,
		Description:  "bbsim-psu",
		Parent:       "",
		ParentRelPos: 0,
		SerialNum:    psuSerial,
		MfgName:      "bbsim-psu",
		IsFru:        false,
		Uuid: &dmi.Uuid{
			Uuid: getUUID(psuName),
		},
		State: &dmi.ComponentState{},
	}
}

//StopManagingDevice stops management of a device and cleans up any context and caches for that device
func (dms *DmiAPIServer) StopManagingDevice(ctx context.Context, req *dmi.StopManagingDeviceRequest) (*dmi.StopManagingDeviceResponse, error) {
	logger.Debugf("StopManagingDevice API invoked")
	if req == nil {
		return &dmi.StopManagingDeviceResponse{Status: dmi.Status_ERROR_STATUS, Reason: dmi.StopManagingDeviceResponse_UNDEFINED_REASON}, status.Errorf(codes.FailedPrecondition, "request is empty")
	}

	if req.Name == "" {
		return &dmi.StopManagingDeviceResponse{Status: dmi.Status_ERROR_STATUS, Reason: dmi.StopManagingDeviceResponse_UNKNOWN_DEVICE},
			status.Errorf(codes.InvalidArgument, "'Name' can not be empty in the request")
	}

	// Stop the components/go routines created
	StopMetricGenerator()

	if dms.mPublisherCancelFunc != nil {
		dms.mPublisherCancelFunc()
	}

	dms.deviceName = ""
	dms.kafkaEndpoint = ""
	dms.ipAddress = ""
	dms.deviceSerial = ""
	dms.ponTransceiverUuids = nil
	dms.ponTransceiverCageUuids = nil
	dms.uuid = ""
	dms.root = nil
	dms.metricChannel = nil

	logger.Infof("Stopped managing the device")
	return &dmi.StopManagingDeviceResponse{Status: dmi.Status_OK_STATUS}, nil
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
			Status:    dmi.Status_ERROR_STATUS,
			Reason:    dmi.PhysicalInventoryResponse_UNKNOWN_DEVICE,
			Inventory: &dmi.Hardware{},
		}

		return sendResponseBackOnStream(stream, errResponse)
	}

	response := &dmi.PhysicalInventoryResponse{
		Status: dmi.Status_OK_STATUS,
		Inventory: &dmi.Hardware{
			LastChange: &timestamp.Timestamp{
				Seconds: 0,
				Nanos:   0,
			},
			Root: dms.root,
		},
	}
	return sendResponseBackOnStream(stream, response)
}

//Contains tells whether arr contains element.
func Contains(arr []string, element string) bool {
	for _, item := range arr {
		if element == item {
			return true
		}
	}
	return false
}

func findComponent(l []*dmi.Component, compUUID string) *dmi.Component {
	var foundComp *dmi.Component

	for _, comp := range l {
		logger.Debugf("findComponent slice comp = %v compUUID = %s", comp, compUUID)
		if comp.Uuid.Uuid == compUUID {
			return comp
		}

		foundComp = findComponent(comp.GetChildren(), compUUID)
		if foundComp != nil {
			return foundComp
		}
	}

	return nil
}

func findComponentsOfType(l []*dmi.Component, compType dmi.ComponentType) []*dmi.Component {
	var comps []*dmi.Component
	findComponents(l, compType, &comps)
	return comps
}

func findComponents(l []*dmi.Component, compType dmi.ComponentType, collector *[]*dmi.Component) {

	for _, comp := range l {
		if comp.Class == compType {
			*collector = append(*collector, comp)
			//logger.Debugf("Added collector = %v", *collector)
		}

		findComponents(comp.GetChildren(), compType, collector)
	}
}

func sendGetHWComponentResponse(c *dmi.Component, stream dmi.NativeHWManagementService_GetHWComponentInfoServer) error {
	apiStatus := dmi.Status_OK_STATUS
	reason := dmi.HWComponentInfoGetResponse_UNDEFINED_REASON

	if c == nil {
		apiStatus = dmi.Status_ERROR_STATUS
		reason = dmi.HWComponentInfoGetResponse_UNKNOWN_DEVICE
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

	//if component list is empty, return error
	if dms.root == nil {
		logger.Errorf("Error occurred, device is not managed")
		return status.Errorf(codes.Internal, "Error occurred, device is not managed, please start managing device")
	}
	// Search for the component and return it
	c := findComponent(dms.root.Children, req.ComponentUuid.Uuid)

	return sendGetHWComponentResponse(c, stream)
}

//SetHWComponentInfo sets the permissible attributes of a HW component
func (dms *DmiAPIServer) SetHWComponentInfo(context.Context, *dmi.HWComponentInfoSetRequest) (*dmi.HWComponentInfoSetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "rpc SetHWComponentInfo not implemented")
}

//SetLoggingEndpoint sets the location to which logs need to be shipped
func (dms *DmiAPIServer) SetLoggingEndpoint(_ context.Context, request *dmi.SetLoggingEndpointRequest) (*dmi.SetRemoteEndpointResponse, error) {
	logger.Debugf("SetLoggingEndpoint called with request %+v", request)
	errRetFunc := func(stat dmi.Status, reason dmi.SetRemoteEndpointResponse_Reason) (*dmi.SetRemoteEndpointResponse, error) {
		return &dmi.SetRemoteEndpointResponse{
			Status: stat,
			Reason: reason,
		}, status.Errorf(codes.InvalidArgument, "invalid request")
	}

	//check the validity of the request
	if request == nil {
		return errRetFunc(dmi.Status_ERROR_STATUS, dmi.SetRemoteEndpointResponse_UNKNOWN_DEVICE)
	}
	if request.LoggingEndpoint == "" {
		return errRetFunc(dmi.Status_ERROR_STATUS, dmi.SetRemoteEndpointResponse_LOGGING_ENDPOINT_ERROR)
	}
	if request.LoggingProtocol == "" {
		return errRetFunc(dmi.Status_ERROR_STATUS, dmi.SetRemoteEndpointResponse_LOGGING_ENDPOINT_PROTOCOL_ERROR)
	}
	if request.DeviceUuid == nil || request.DeviceUuid.Uuid != dms.uuid {
		return errRetFunc(dmi.Status_ERROR_STATUS, dmi.SetRemoteEndpointResponse_UNKNOWN_DEVICE)
	}

	dms.loggingEndpoint = request.LoggingEndpoint
	dms.loggingProtocol = request.LoggingProtocol

	return &dmi.SetRemoteEndpointResponse{
		Status: dmi.Status_OK_STATUS,
	}, nil
}

//GetLoggingEndpoint gets the configured location to which the logs are being shipped
func (dms *DmiAPIServer) GetLoggingEndpoint(_ context.Context, request *dmi.HardwareID) (*dmi.GetLoggingEndpointResponse, error) {
	logger.Debugf("GetLoggingEndpoint called with request %+v", request)
	if request == nil || request.Uuid == nil || request.Uuid.Uuid == "" {
		return &dmi.GetLoggingEndpointResponse{
			Status: dmi.Status_ERROR_STATUS,
			Reason: dmi.GetLoggingEndpointResponse_UNKNOWN_DEVICE,
		}, status.Errorf(codes.InvalidArgument, "invalid request")
	}
	if request.Uuid.Uuid != dms.uuid {
		return &dmi.GetLoggingEndpointResponse{
			Status: dmi.Status_ERROR_STATUS,
			Reason: dmi.GetLoggingEndpointResponse_UNKNOWN_DEVICE,
		}, nil
	}

	return &dmi.GetLoggingEndpointResponse{
		Status:          dmi.Status_OK_STATUS,
		LoggingEndpoint: dms.loggingEndpoint,
		LoggingProtocol: dms.loggingProtocol,
	}, nil
}

//SetMsgBusEndpoint sets the location of the Message Bus to which events and metrics are shipped
func (dms *DmiAPIServer) SetMsgBusEndpoint(ctx context.Context, request *dmi.SetMsgBusEndpointRequest) (*dmi.SetRemoteEndpointResponse, error) {
	logger.Debugf("SetMsgBusEndpoint() invoked with request: %+v and context: %v", request, ctx)
	if request == nil || request.MsgbusEndpoint == "" {
		return &dmi.SetRemoteEndpointResponse{Status: dmi.Status_ERROR_STATUS, Reason: dmi.SetRemoteEndpointResponse_MSGBUS_ENDPOINT_ERROR},
			status.Errorf(codes.FailedPrecondition, "request is nil")
	}
	olt := devices.GetOLT()
	dms.kafkaEndpoint = request.MsgbusEndpoint

	// close the old publisher
	if dms.mPublisherCancelFunc != nil {
		dms.mPublisherCancelFunc()
	}

	// initialize a new publisher
	var nCtx context.Context
	nCtx, dms.mPublisherCancelFunc = context.WithCancel(context.Background())
	// initialize a publisher
	if err := InitializeDMKafkaPublishers(sarama.NewAsyncProducer, olt.ID, dms.kafkaEndpoint); err == nil {
		// start a go routine which will read from channel and publish on kafka topic dm.metrics
		go DMKafkaPublisher(nCtx, dms.metricChannel, "dm.metrics")
		// start a go routine which will read from channel and publish on kafka topic dm.events
		go DMKafkaPublisher(nCtx, dms.eventChannel, "dm.events")
	} else {
		logger.Errorf("Failed to start metric kafka publisher: %v", err)
		return &dmi.SetRemoteEndpointResponse{Status: dmi.Status_ERROR_STATUS, Reason: dmi.SetRemoteEndpointResponse_MSGBUS_ENDPOINT_ERROR}, err
	}

	return &dmi.SetRemoteEndpointResponse{Status: dmi.Status_OK_STATUS, Reason: dmi.SetRemoteEndpointResponse_UNDEFINED_REASON}, nil
}

//GetMsgBusEndpoint gets the configured location to which the events and metrics are being shipped
func (dms *DmiAPIServer) GetMsgBusEndpoint(context.Context, *empty.Empty) (*dmi.GetMsgBusEndpointResponse, error) {
	logger.Debugf("GetMsgBusEndpoint() invoked")
	if dms.kafkaEndpoint != "" {
		return &dmi.GetMsgBusEndpointResponse{
			Status:         dmi.Status_OK_STATUS,
			Reason:         dmi.GetMsgBusEndpointResponse_UNDEFINED_REASON,
			MsgbusEndpoint: dms.kafkaEndpoint,
		}, nil
	}
	return &dmi.GetMsgBusEndpointResponse{
		Status:         dmi.Status_ERROR_STATUS,
		Reason:         dmi.GetMsgBusEndpointResponse_INTERNAL_ERROR,
		MsgbusEndpoint: "",
	}, nil
}

//GetManagedDevices returns an object containing a list of devices managed by this entity
func (dms *DmiAPIServer) GetManagedDevices(context.Context, *empty.Empty) (*dmi.ManagedDevicesResponse, error) {
	retResponse := dmi.ManagedDevicesResponse{}
	//If our uuid is empty, we return empty list; else we fill details and return
	if dms.uuid != "" {
		root := dmi.ManagedDeviceInfo{
			Info: &dmi.ModifiableComponent{
				Name: dms.deviceName,
				Uri: &dmi.Uri{
					Uri: dms.ipAddress,
				},
			},
			DeviceUuid: &dmi.Uuid{
				Uuid: dms.uuid,
			},
		}

		retResponse.Devices = append(retResponse.Devices, &root)
	}

	return &retResponse, nil
}

//GetLogLevel Gets the configured log level for a certain entity on a certain device.
func (dms *DmiAPIServer) GetLogLevel(context.Context, *dmi.GetLogLevelRequest) (*dmi.GetLogLevelResponse, error) {
	return &dmi.GetLogLevelResponse{
		Status: dmi.Status_OK_STATUS,
		DeviceUuid: &dmi.Uuid{
			Uuid: dms.uuid,
		},
		LogLevels: []*dmi.EntitiesLogLevel{},
	}, nil
}

// SetLogLevel Sets the log level of the device, for each given entity to a certain level.
func (dms *DmiAPIServer) SetLogLevel(context.Context, *dmi.SetLogLevelRequest) (*dmi.SetLogLevelResponse, error) {
	return &dmi.SetLogLevelResponse{
		Status: dmi.Status_OK_STATUS,
		DeviceUuid: &dmi.Uuid{
			Uuid: dms.uuid,
		},
	}, nil
}

// GetLoggableEntities Gets the entities of a device on which log can be configured.
func (dms *DmiAPIServer) GetLoggableEntities(context.Context, *dmi.GetLoggableEntitiesRequest) (*dmi.GetLogLevelResponse, error) {
	return &dmi.GetLogLevelResponse{
		Status: dmi.Status_OK_STATUS,
		DeviceUuid: &dmi.Uuid{
			Uuid: dms.uuid,
		},
		LogLevels: []*dmi.EntitiesLogLevel{},
	}, nil
}

// Performs the heartbeat check
func (dms *DmiAPIServer) HeartbeatCheck(context.Context, *empty.Empty) (*dmi.Heartbeat, error) {
	return nil, status.Errorf(codes.Unimplemented, "rpc HeartbeatCheck not implemented")
}

// Performs the reboot of the device
func (dms *DmiAPIServer) RebootDevice(context.Context, *dmi.RebootDeviceRequest) (*dmi.RebootDeviceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "rpc RebootDevice not implemented")
}
