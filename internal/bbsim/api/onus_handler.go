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

	"github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsim/alarmsim"
	"github.com/opencord/bbsim/internal/bbsim/devices"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
)

func (s BBSimServer) GetONUs(ctx context.Context, req *bbsim.Empty) (*bbsim.ONUs, error) {
	olt := devices.GetOLT()
	onus := bbsim.ONUs{
		Items: []*bbsim.ONU{},
	}

	for _, pon := range olt.Pons {
		for _, o := range pon.Onus {
			onu := bbsim.ONU{
				ID:            int32(o.ID),
				SerialNumber:  o.Sn(),
				OperState:     o.OperState.Current(),
				InternalState: o.InternalState.Current(),
				PonPortID:     int32(o.PonPortID),
				PortNo:        int32(o.PortNo),
				Services:      convertBBsimServicesToProtoServices(o.Services),
			}
			onus.Items = append(onus.Items, &onu)
		}
	}
	return &onus, nil
}

func (s BBSimServer) GetONU(ctx context.Context, req *bbsim.ONURequest) (*bbsim.ONU, error) {
	olt := devices.GetOLT()
	onu, err := olt.FindOnuBySn(req.SerialNumber)

	if err != nil {
		res := bbsim.ONU{}
		return &res, err
	}

	res := bbsim.ONU{
		ID:            int32(onu.ID),
		SerialNumber:  onu.Sn(),
		OperState:     onu.OperState.Current(),
		InternalState: onu.InternalState.Current(),
		PonPortID:     int32(onu.PonPortID),
		PortNo:        int32(onu.PortNo),
		Services:      convertBBsimServicesToProtoServices(onu.Services),
	}
	return &res, nil
}

// ShutdownONU sends DyingGasp indication for specified ONUs and mark ONUs as disabled.
func (s BBSimServer) ShutdownONU(ctx context.Context, req *bbsim.ONURequest) (*bbsim.Response, error) {
	// NOTE this method is now sending a Dying Gasp and then disabling the device (operState: down, adminState: up),
	// is this the only way to do? Should we address other cases?
	// Investigate what happens when:
	// - a fiber is pulled
	// - ONU malfunction
	// - ONU shutdown
	logger.WithFields(log.Fields{
		"OnuSn": req.SerialNumber,
	}).Infof("Received request to shutdown ONU")

	res := &bbsim.Response{}
	olt := devices.GetOLT()

	onu, err := olt.FindOnuBySn(req.SerialNumber)
	if err != nil {
		res.StatusCode = int32(codes.NotFound)
		res.Message = err.Error()
		return res, err
	}

	return handleShutdownONU(onu)
}

// ShutdownONUsOnPON sends DyingGasp indication for all ONUs under specified PON port
func (s BBSimServer) ShutdownONUsOnPON(ctx context.Context, req *bbsim.PONRequest) (*bbsim.Response, error) {
	logger.WithFields(log.Fields{
		"IntfId": req.PonPortId,
	}).Infof("Received request to shutdown all ONUs on PON")

	res := &bbsim.Response{}
	olt := devices.GetOLT()
	pon, _ := olt.GetPonById(req.PonPortId)

	go func() {
		for _, onu := range pon.Onus {
			res, _ = handleShutdownONU(onu)
		}
	}()
	res.StatusCode = int32(codes.OK)
	res.Message = fmt.Sprintf("Request accepted for shutdown all ONUs on PON port %d", pon.ID)

	return res, nil
}

// ShutdownAllONUs sends DyingGasp indication for all ONUs and mark ONUs as disabled.
func (s BBSimServer) ShutdownAllONUs(context.Context, *bbsim.Empty) (*bbsim.Response, error) {
	logger.Infof("Received request to shutdown all ONUs")
	res := &bbsim.Response{}
	olt := devices.GetOLT()

	go func() {
		for _, pon := range olt.Pons {
			for _, onu := range pon.Onus {
				res, _ = handleShutdownONU(onu)
			}
		}
	}()
	res.StatusCode = int32(codes.OK)
	res.Message = fmt.Sprintf("Request Accepted for shutdown all ONUs in OLT %d", olt.ID)

	return res, nil
}

// PoweronONU simulates ONU power on and start sending discovery indications to VOLTHA
func (s BBSimServer) PoweronONU(ctx context.Context, req *bbsim.ONURequest) (*bbsim.Response, error) {
	logger.WithFields(log.Fields{
		"OnuSn": req.SerialNumber,
	}).Infof("Received request to poweron ONU")

	res := &bbsim.Response{}
	olt := devices.GetOLT()

	onu, err := olt.FindOnuBySn(req.SerialNumber)
	if err != nil {
		res.StatusCode = int32(codes.NotFound)
		res.Message = err.Error()
		return res, err
	}

	pon, _ := olt.GetPonById(onu.PonPortID)
	if pon.InternalState.Current() != "enabled" {
		err := fmt.Errorf("PON port %d not enabled", onu.PonPortID)
		logger.WithFields(log.Fields{
			"OnuId":  onu.ID,
			"IntfId": onu.PonPortID,
			"OnuSn":  onu.Sn(),
		}).Errorf("Cannot poweron ONU: %s", err.Error())

		res.StatusCode = int32(codes.FailedPrecondition)
		res.Message = err.Error()
		return res, err
	}

	return handlePoweronONU(onu)
}

// PoweronONUsOnPON simulates ONU power on for all ONUs under specified PON port
func (s BBSimServer) PoweronONUsOnPON(ctx context.Context, req *bbsim.PONRequest) (*bbsim.Response, error) {
	logger.WithFields(log.Fields{
		"IntfId": req.PonPortId,
	}).Infof("Received request to poweron all ONUs on PON")

	res := &bbsim.Response{}
	olt := devices.GetOLT()

	pon, _ := olt.GetPonById(req.PonPortId)
	if pon.InternalState.Current() != "enabled" {
		err := fmt.Errorf("PON port %d not enabled", pon.ID)
		logger.WithFields(log.Fields{
			"IntfId": pon.ID,
		}).Errorf("Cannot poweron ONUs on PON: %s", err.Error())

		res.StatusCode = int32(codes.FailedPrecondition)
		res.Message = err.Error()
		return res, err
	}

	go func() {
		for _, onu := range pon.Onus {
			res, _ = handlePoweronONU(onu)
		}
	}()
	res.StatusCode = int32(codes.OK)
	res.Message = fmt.Sprintf("Request Accepted for power on all ONUs on PON port %d", pon.ID)

	return res, nil
}

// PoweronAllONUs simulates ONU power on for all ONUs on all PON ports
func (s BBSimServer) PoweronAllONUs(context.Context, *bbsim.Empty) (*bbsim.Response, error) {
	logger.Infof("Received request to poweron all ONUs")

	res := &bbsim.Response{}
	olt := devices.GetOLT()

	go func() {
		for _, pon := range olt.Pons {
			if pon.InternalState.Current() == "enabled" {
				for _, onu := range pon.Onus {
					res, _ = handlePoweronONU(onu)
				}
			}
		}
	}()
	res.StatusCode = int32(codes.OK)
	res.Message = fmt.Sprintf("Request Accepted for power on all ONUs in OLT %d", olt.ID)

	return res, nil
}

func (s BBSimServer) ChangeIgmpState(ctx context.Context, req *bbsim.IgmpRequest) (*bbsim.Response, error) {

	// TODO check that the ONU is enabled and the services are initialized before changing the state

	res := &bbsim.Response{}

	logger.WithFields(log.Fields{
		"OnuSn":     req.OnuReq.SerialNumber,
		"subAction": req.SubActionVal,
	}).Infof("Received igmp request for ONU")

	olt := devices.GetOLT()
	onu, err := olt.FindOnuBySn(req.OnuReq.SerialNumber)

	if err != nil {
		res.StatusCode = int32(codes.NotFound)
		res.Message = err.Error()
		fmt.Println("ONU not found for sending igmp packet.")
		return res, err
	} else {
		event := ""
		switch req.SubActionVal {
		case bbsim.SubActionTypes_JOIN:
			event = "igmp_join_start"
		case bbsim.SubActionTypes_LEAVE:
			event = "igmp_leave"
		case bbsim.SubActionTypes_JOINV3:
			event = "igmp_join_startv3"
		}

		errors := []string{}
		startedOn := []string{}
		success := true

		for _, s := range onu.Services {
			service := s.(*devices.Service)
			if service.NeedsIgmp {

				logger.WithFields(log.Fields{
					"OnuId":   onu.ID,
					"IntfId":  onu.PonPortID,
					"OnuSn":   onu.Sn(),
					"Service": service.Name,
				}).Debugf("Sending %s event on Service %s", event, service.Name)

				if err := service.IGMPState.Event(event); err != nil {
					logger.WithFields(log.Fields{
						"OnuId":   onu.ID,
						"IntfId":  onu.PonPortID,
						"OnuSn":   onu.Sn(),
						"Service": service.Name,
					}).Errorf("IGMP request failed: %s", err.Error())
					errors = append(errors, fmt.Sprintf("%s: %s", service.Name, err.Error()))
					success = false
				}
				startedOn = append(startedOn, service.Name)
			}
		}

		if success {
			res.StatusCode = int32(codes.OK)
			res.Message = fmt.Sprintf("Authentication restarted on Services %s for ONU %s.",
				fmt.Sprintf("%v", startedOn), onu.Sn())
		} else {
			res.StatusCode = int32(codes.FailedPrecondition)
			res.Message = fmt.Sprintf("%v", errors)
		}
	}

	return res, nil
}

func (s BBSimServer) RestartEapol(ctx context.Context, req *bbsim.ONURequest) (*bbsim.Response, error) {
	res := &bbsim.Response{}

	logger.WithFields(log.Fields{
		"OnuSn": req.SerialNumber,
	}).Infof("Received request to restart authentication ONU")

	olt := devices.GetOLT()

	onu, err := olt.FindOnuBySn(req.SerialNumber)

	if err != nil {
		res.StatusCode = int32(codes.NotFound)
		res.Message = err.Error()
		return res, err
	}

	errors := []string{}
	startedOn := []string{}
	success := true

	for _, s := range onu.Services {
		service := s.(*devices.Service)
		if service.NeedsEapol {
			if err := service.EapolState.Event("start_auth"); err != nil {
				logger.WithFields(log.Fields{
					"OnuId":   onu.ID,
					"IntfId":  onu.PonPortID,
					"OnuSn":   onu.Sn(),
					"Service": service.Name,
				}).Errorf("Cannot restart authenticaton for Service: %s", err.Error())
				errors = append(errors, fmt.Sprintf("%s: %s", service.Name, err.Error()))
				success = false
			}
			startedOn = append(startedOn, service.Name)
		}
	}

	if success {
		res.StatusCode = int32(codes.OK)
		res.Message = fmt.Sprintf("Authentication restarted on Services %s for ONU %s.",
			fmt.Sprintf("%v", startedOn), onu.Sn())
	} else {
		res.StatusCode = int32(codes.FailedPrecondition)
		res.Message = fmt.Sprintf("%v", errors)
	}

	return res, nil
}

func (s BBSimServer) RestartDhcp(ctx context.Context, req *bbsim.ONURequest) (*bbsim.Response, error) {
	res := &bbsim.Response{}

	logger.WithFields(log.Fields{
		"OnuSn": req.SerialNumber,
	}).Infof("Received request to restart DHCP on ONU")

	olt := devices.GetOLT()

	onu, err := olt.FindOnuBySn(req.SerialNumber)

	if err != nil {
		res.StatusCode = int32(codes.NotFound)
		res.Message = err.Error()
		return res, err
	}

	errors := []string{}
	startedOn := []string{}
	success := true

	for _, s := range onu.Services {
		service := s.(*devices.Service)
		if service.NeedsDhcp {

			if err := service.DHCPState.Event("start_dhcp"); err != nil {
				logger.WithFields(log.Fields{
					"OnuId":   onu.ID,
					"IntfId":  onu.PonPortID,
					"OnuSn":   onu.Sn(),
					"Service": service.Name,
				}).Errorf("Cannot restart DHCP for Service: %s", err.Error())
				errors = append(errors, fmt.Sprintf("%s: %s", service.Name, err.Error()))
				success = false
			}
			startedOn = append(startedOn, service.Name)
		}
	}

	if success {
		res.StatusCode = int32(codes.OK)
		res.Message = fmt.Sprintf("DHCP restarted on Services %s for ONU %s.",
			fmt.Sprintf("%v", startedOn), onu.Sn())
	} else {
		res.StatusCode = int32(codes.FailedPrecondition)
		res.Message = fmt.Sprintf("%v", errors)
	}

	return res, nil
}

// GetFlows for OLT/ONUs
func (s BBSimServer) GetFlows(ctx context.Context, req *bbsim.ONURequest) (*bbsim.Flows, error) {
	logger.WithFields(log.Fields{
		"OnuSn": req.SerialNumber,
	}).Info("Received GetFlows request")

	olt := devices.GetOLT()
	res := &bbsim.Flows{}

	if req.SerialNumber == "" {
		for flowKey := range olt.Flows {
			flow := olt.Flows[flowKey]
			res.Flows = append(res.Flows, &flow)
		}
		res.FlowCount = uint32(len(olt.Flows))
	} else {
		onu, err := olt.FindOnuBySn(req.SerialNumber)
		if err != nil {
			logger.WithFields(log.Fields{
				"OnuSn": req.SerialNumber,
			}).Error("Can't get ONU in GetFlows request")
			return nil, err
		}
		for _, flowKey := range onu.Flows {
			flow := olt.Flows[flowKey]
			res.Flows = append(res.Flows, &flow)
		}
		res.FlowCount = uint32(len(onu.Flows))
	}
	return res, nil
}

func (s BBSimServer) GetOnuTrafficSchedulers(ctx context.Context, req *bbsim.ONURequest) (*bbsim.ONUTrafficSchedulers, error) {
	olt := devices.GetOLT()
	ts := bbsim.ONUTrafficSchedulers{}

	onu, err := olt.FindOnuBySn(req.SerialNumber)
	if err != nil {
		return &ts, err
	}

	if onu.TrafficSchedulers != nil {
		ts.TraffSchedulers = onu.TrafficSchedulers
		return &ts, nil
	} else {
		ts.TraffSchedulers = nil
		return &ts, nil
	}
}

func handlePoweronONU(onu *devices.Onu) (*bbsim.Response, error) {
	res := &bbsim.Response{}
	olt := devices.GetOLT()
	intitalState := onu.InternalState.Current()
	if onu.InternalState.Current() == "created" || onu.InternalState.Current() == "disabled" {
		if err := onu.InternalState.Event("initialize"); err != nil {
			logger.WithFields(log.Fields{
				"OnuId":  onu.ID,
				"IntfId": onu.PonPortID,
				"OnuSn":  onu.Sn(),
			}).Errorf("Cannot poweron ONU: %s", err.Error())
			res.StatusCode = int32(codes.FailedPrecondition)
			res.Message = err.Error()
			return res, err
		}
	}

	losReq := bbsim.ONUAlarmRequest{
		AlarmType:    "ONU_ALARM_LOS",
		SerialNumber: onu.Sn(),
		Status:       "off",
	}

	if err := alarmsim.SimulateOnuAlarm(context.TODO(), &losReq, olt); err != nil {
		logger.WithFields(log.Fields{
			"OnuId":  onu.ID,
			"IntfId": onu.PonPortID,
			"OnuSn":  onu.Sn(),
		}).Errorf("Cannot send LOS: %s", err.Error())
		res.StatusCode = int32(codes.FailedPrecondition)
		res.Message = err.Error()
		return res, err
	}

	if err := onu.InternalState.Event("discover"); err != nil {
		logger.WithFields(log.Fields{
			"OnuId":  onu.ID,
			"IntfId": onu.PonPortID,
			"OnuSn":  onu.Sn(),
		}).Errorf("Cannot poweron ONU: %s", err.Error())
		res.StatusCode = int32(codes.FailedPrecondition)
		res.Message = err.Error()
		return res, err
	}
	// move onu directly to enable state only when its a powercycle case
	// in case of first time onu poweron onu will be moved to enable on
	// receiving ActivateOnu request from openolt adapter
	if intitalState == "disabled" {
		if err := onu.InternalState.Event("enable"); err != nil {
			logger.WithFields(log.Fields{
				"OnuId":  onu.ID,
				"IntfId": onu.PonPortID,
				"OnuSn":  onu.Sn(),
			}).Errorf("Cannot enable ONU: %s", err.Error())
			res.StatusCode = int32(codes.FailedPrecondition)
			res.Message = err.Error()
			return res, err
		}
	}

	res.StatusCode = int32(codes.OK)
	res.Message = fmt.Sprintf("ONU %s successfully powered on.", onu.Sn())

	return res, nil
}

func handleShutdownONU(onu *devices.Onu) (*bbsim.Response, error) {
	res := &bbsim.Response{}
	olt := devices.GetOLT()

	dyingGasp := bbsim.ONUAlarmRequest{
		AlarmType:    "DYING_GASP",
		SerialNumber: onu.Sn(),
		Status:       "on",
	}

	if err := alarmsim.SimulateOnuAlarm(context.TODO(), &dyingGasp, olt); err != nil {
		logger.WithFields(log.Fields{
			"OnuId":  onu.ID,
			"IntfId": onu.PonPortID,
			"OnuSn":  onu.Sn(),
		}).Errorf("Cannot send Dying Gasp: %s", err.Error())
		res.StatusCode = int32(codes.FailedPrecondition)
		res.Message = err.Error()
		return res, err
	}

	losReq := bbsim.ONUAlarmRequest{
		AlarmType:    "ONU_ALARM_LOS",
		SerialNumber: onu.Sn(),
		Status:       "on",
	}

	if err := alarmsim.SimulateOnuAlarm(context.TODO(), &losReq, olt); err != nil {
		logger.WithFields(log.Fields{
			"OnuId":  onu.ID,
			"IntfId": onu.PonPortID,
			"OnuSn":  onu.Sn(),
		}).Errorf("Cannot send LOS: %s", err.Error())
		res.StatusCode = int32(codes.FailedPrecondition)
		res.Message = err.Error()
		return res, err
	}

	// TODO if it's the last ONU on the PON, then send a PON LOS

	if err := onu.InternalState.Event("disable"); err != nil {
		logger.WithFields(log.Fields{
			"OnuId":  onu.ID,
			"IntfId": onu.PonPortID,
			"OnuSn":  onu.Sn(),
		}).Errorf("Cannot shutdown ONU: %s", err.Error())
		res.StatusCode = int32(codes.FailedPrecondition)
		res.Message = err.Error()
		return res, err
	}

	res.StatusCode = int32(codes.OK)
	res.Message = fmt.Sprintf("ONU %s successfully shut down.", onu.Sn())

	return res, nil
}
