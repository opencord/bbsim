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
	"fmt"
	"sync"

	"github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/device-management-interface/go/dmi"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//DmiEventsGenerator has the attributes for generating events
type DmiEventsGenerator struct {
	apiSrv           *DmiAPIServer
	configuredEvents map[dmi.EventIds]dmi.EventCfg
	access           sync.Mutex
}

// func to generate the different types of events, there are two types one with thresholds and one without
type eventGenerationFunc func(dmi.EventIds, dmi.ComponentType)

// eventGenerationUtil contains the component and the func for a specific eventId
type eventGenerationUtil struct {
	componentType dmi.ComponentType
	genFunc       eventGenerationFunc
}

var dmiEG DmiEventsGenerator
var eventGenMap map[dmi.EventIds]eventGenerationUtil

func init() {
	eventGenMap = make(map[dmi.EventIds]eventGenerationUtil)
	eventGenMap[dmi.EventIds_EVENT_FAN_FAILURE] = eventGenerationUtil{
		componentType: dmi.ComponentType_COMPONENT_TYPE_FAN,
		genFunc:       noThresholdEventGenerationFunc,
	}
	eventGenMap[dmi.EventIds_EVENT_FAN_FAILURE_RECOVERED] = eventGenerationUtil{
		componentType: dmi.ComponentType_COMPONENT_TYPE_FAN,
		genFunc:       noThresholdEventGenerationFunc,
	}
	eventGenMap[dmi.EventIds_EVENT_TRANSCEIVER_PLUG_IN] = eventGenerationUtil{
		componentType: dmi.ComponentType_COMPONENT_TYPE_TRANSCEIVER,
		genFunc:       noThresholdEventGenerationFunc,
	}
	eventGenMap[dmi.EventIds_EVENT_TRANSCEIVER_PLUG_OUT] = eventGenerationUtil{
		componentType: dmi.ComponentType_COMPONENT_TYPE_TRANSCEIVER,
		genFunc:       noThresholdEventGenerationFunc,
	}

	eventGenMap[dmi.EventIds_EVENT_PSU_FAILURE] = eventGenerationUtil{
		componentType: dmi.ComponentType_COMPONENT_TYPE_POWER_SUPPLY,
		genFunc:       noThresholdEventGenerationFunc,
	}
	eventGenMap[dmi.EventIds_EVENT_PSU_FAILURE_RECOVERED] = eventGenerationUtil{
		componentType: dmi.ComponentType_COMPONENT_TYPE_POWER_SUPPLY,
		genFunc:       noThresholdEventGenerationFunc,
	}

	eventGenMap[dmi.EventIds_EVENT_HW_DEVICE_TEMPERATURE_ABOVE_CRITICAL] = eventGenerationUtil{
		componentType: dmi.ComponentType_COMPONENT_TYPE_SENSOR,
		genFunc:       thresholdEventGenerationFunc,
	}
	eventGenMap[dmi.EventIds_EVENT_HW_DEVICE_TEMPERATURE_ABOVE_CRITICAL_RECOVERED] = eventGenerationUtil{
		componentType: dmi.ComponentType_COMPONENT_TYPE_SENSOR,
		genFunc:       thresholdEventGenerationFunc,
	}
}

//StartEventsGenerator initializes the event generator
func StartEventsGenerator(apiSrv *DmiAPIServer) {
	log.Debugf("StartEventsGenerator invoked")

	dmiEG = DmiEventsGenerator{
		apiSrv: apiSrv,
	}
	dmiEG.configuredEvents = make(map[dmi.EventIds]dmi.EventCfg)

	// Add Fan Failure event configuration
	dmiEG.configuredEvents[dmi.EventIds_EVENT_FAN_FAILURE] = dmi.EventCfg{
		EventId:      dmi.EventIds_EVENT_FAN_FAILURE,
		IsConfigured: true,
	}

	// Add hardware device temp above critical event configuration
	dmiEG.configuredEvents[dmi.EventIds_EVENT_HW_DEVICE_TEMPERATURE_ABOVE_CRITICAL] = dmi.EventCfg{
		EventId:      dmi.EventIds_EVENT_HW_DEVICE_TEMPERATURE_ABOVE_CRITICAL,
		IsConfigured: true,
		Thresholds: &dmi.Thresholds{
			Threshold: &dmi.Thresholds_Upper{Upper: &dmi.WaterMarks{
				High: &dmi.ValueType{
					Val: &dmi.ValueType_IntVal{IntVal: 95},
				},
				Low: &dmi.ValueType{
					Val: &dmi.ValueType_IntVal{IntVal: 90},
				},
			}},
		},
	}

	// Add Power Supply Unit failure event configuration
	dmiEG.configuredEvents[dmi.EventIds_EVENT_PSU_FAILURE] = dmi.EventCfg{
		EventId:      dmi.EventIds_EVENT_PSU_FAILURE,
		IsConfigured: true,
	}

	// Add Transceiver Plug in and out event configuration
	dmiEG.configuredEvents[dmi.EventIds_EVENT_TRANSCEIVER_PLUG_IN] = dmi.EventCfg{
		EventId:      dmi.EventIds_EVENT_TRANSCEIVER_PLUG_IN,
		IsConfigured: true,
	}
	dmiEG.configuredEvents[dmi.EventIds_EVENT_TRANSCEIVER_PLUG_OUT] = dmi.EventCfg{
		EventId:      dmi.EventIds_EVENT_TRANSCEIVER_PLUG_OUT,
		IsConfigured: true,
	}
}

// get the events list
func getEventsList() []*dmi.EventCfg {
	events := make(map[dmi.EventIds]dmi.EventCfg)
	dmiEG.access.Lock()

	for key, value := range dmiEG.configuredEvents {
		events[key] = value
	}

	dmiEG.access.Unlock()

	var toRet []*dmi.EventCfg
	for _, v := range events {
		eventConfig := v
		toRet = append(toRet, &eventConfig)
	}
	logger.Debugf("Events list supported by device %+v", toRet)
	return toRet
}

//UpdateEventConfig Adds/Updates the passed event configuration
func UpdateEventConfig(newEventCfg *dmi.EventCfg) {
	dmiEG.access.Lock()
	dmiEG.configuredEvents[newEventCfg.GetEventId()] = *newEventCfg
	dmiEG.access.Unlock()
	logger.Infof("Events updated %v", newEventCfg)
}

// update Event MetaData
func updateEventMetaData(c *dmi.Component, apiSrv *DmiAPIServer, evt *dmi.Event) *dmi.Event {
	evt.EventMetadata = &dmi.EventMetaData{
		DeviceUuid:    apiSrv.uuid,
		ComponentUuid: c.Uuid,
		ComponentName: c.Name,
	}
	return evt
}

func sendOutEventOnKafka(event interface{}, apiSrv *DmiAPIServer) {
	select {
	case apiSrv.eventChannel <- event:
	default:
		logger.Debugf("Channel not ready dropping event")
	}
}

func noThresholdEventGenerationFunc(eventID dmi.EventIds, cType dmi.ComponentType) {
	for _, comp := range findComponentsOfType(dmiEG.apiSrv.root.Children, cType) {
		var evnt dmi.Event
		evnt.EventId = eventID
		evnt = *updateEventMetaData(comp, dmiEG.apiSrv, &evnt)
		evnt.RaisedTs = timestamppb.Now()
		logger.Debugf("Got a No Threshold event %+v", evnt)
		sendOutEventOnKafka(evnt, dmiEG.apiSrv)
		break
	}
}

func thresholdEventGenerationFunc(eventID dmi.EventIds, cType dmi.ComponentType) {
	eventGenerated := false
	for _, comp := range findComponentsOfType(dmiEG.apiSrv.root.Children, cType) {
		var evnt dmi.Event
		evnt.EventId = eventID
		evnt = *updateEventMetaData(comp, dmiEG.apiSrv, &evnt)
		evnt.RaisedTs = timestamppb.Now()
		configuredEvents := make(map[dmi.EventIds]dmi.EventCfg)

		dmiEG.access.Lock()
		for key, value := range dmiEG.configuredEvents {
			configuredEvents[key] = value
		}
		dmiEG.access.Unlock()

		for k, v := range configuredEvents {
			if k == eventID {
				evnt.ThresholdInfo = &dmi.ThresholdInformation{
					ObservedValue: &dmi.ValueType{
						Val: &dmi.ValueType_IntVal{IntVal: int64(generateRand(int32(v.Thresholds.GetUpper().GetLow().GetIntVal()), int32(v.Thresholds.GetUpper().GetHigh().GetIntVal())))},
					},
					Thresholds: v.GetThresholds(),
				}
			}
		}

		logger.Debugf("Got Threshold event %v", evnt)
		sendOutEventOnKafka(evnt, dmiEG.apiSrv)
		eventGenerated = true
		if eventGenerated {
			break
		}

	}
}

// CreateEvent creates and the passed event if it's valid and sends it to the msg bus
func (dms *DmiAPIServer) CreateEvent(ctx context.Context, evt *bbsim.DmiEvent) (*bbsim.DmiResponse, error) {
	retFunc := func(code codes.Code, msg string) (*bbsim.DmiResponse, error) {
		res := &bbsim.DmiResponse{}
		res.StatusCode = int32(code)
		res.Message = msg
		return res, nil
	}

	if dmiEG.apiSrv == nil || dmiEG.apiSrv.root == nil || dmiEG.apiSrv.root.Children == nil {
		// inventory might not yet be created
		return retFunc(codes.Internal, "Inventory does not exist")
	}

	eventID, exists := dmi.EventIds_value[evt.EventName]
	if !exists {
		return retFunc(codes.NotFound,
			fmt.Sprintf("DMI Alarm not supported. Permissible values are %s", getValidEventNames()))
	}

	genUtil, exists := eventGenMap[dmi.EventIds(eventID)]
	if !exists {
		return retFunc(codes.Unimplemented, "Generation of this event not yet implemented")
	}

	genUtil.genFunc(dmi.EventIds(eventID), genUtil.componentType)

	return retFunc(codes.OK, "DMI Event Indication Sent.")

}

func getValidEventNames() string {
	s := ""
	//keys := make([]string, len(dmi.EventIds_value)-1)
	for k, v := range dmi.EventIds_value {
		if v != 0 {
			s = s + "\n" + k
		}
	}
	return s
}
