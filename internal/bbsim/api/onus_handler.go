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
	"fmt"
	"strconv"

	"github.com/opencord/bbsim/internal/bbsim/types"
	"github.com/opencord/voltha-protos/v5/go/openolt"

	"github.com/opencord/bbsim/api/bbsim"
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
				ID:                            int32(o.ID),
				SerialNumber:                  o.Sn(),
				OperState:                     o.OperState.Current(),
				InternalState:                 o.InternalState.Current(),
				PonPortID:                     int32(o.PonPortID),
				ImageSoftwareReceivedSections: int32(o.ImageSoftwareReceivedSections),
				ImageSoftwareExpectedSections: int32(o.ImageSoftwareExpectedSections),
				ActiveImageEntityId:           int32(o.ActiveImageEntityId),
				CommittedImageEntityId:        int32(o.CommittedImageEntityId),
				Unis:                          append(convertBBsimUniPortsToProtoUniPorts(o.UniPorts), convertBBsimPotsPortsToProtoUniPorts(o.PotsPorts)...),
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
		Unis:          append(convertBBsimUniPortsToProtoUniPorts(onu.UniPorts), convertBBsimPotsPortsToProtoUniPorts(onu.PotsPorts)...),
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

	// NOTE this API will change the IGMP state for all UNIs on the requested ONU
	// TODO a new API needs to be created to individually manage the UNIs

	res := &bbsim.Response{}

	logger.WithFields(log.Fields{
		"OnuSn":        req.OnuReq.SerialNumber,
		"subAction":    req.SubActionVal,
		"GroupAddress": req.GroupAddress,
	}).Info("Received igmp request for ONU")

	olt := devices.GetOLT()
	onu, err := olt.FindOnuBySn(req.OnuReq.SerialNumber)

	if err != nil {
		res.StatusCode = int32(codes.NotFound)
		res.Message = err.Error()
		logger.WithFields(log.Fields{
			"OnuSn":        req.OnuReq.SerialNumber,
			"subAction":    req.SubActionVal,
			"GroupAddress": req.GroupAddress,
		}).Warn("ONU not found for sending igmp packet.")
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

		for _, u := range onu.UniPorts {
			uni := u.(*devices.UniPort)
			if !uni.OperState.Is(devices.UniStateUp) {
				// if the UNI is disabled, ignore it
				continue
			}
			for _, s := range uni.Services {
				service := s.(*devices.Service)
				serviceKey := fmt.Sprintf("uni[%d]%s", uni.ID, service.Name)
				if service.NeedsIgmp {
					if !service.InternalState.Is(devices.ServiceStateInitialized) {
						logger.WithFields(log.Fields{
							"OnuId":   onu.ID,
							"UniId":   uni.ID,
							"IntfId":  onu.PonPortID,
							"OnuSn":   onu.Sn(),
							"Service": service.Name,
						}).Warn("service-not-initialized-skipping-event")
						continue
					}
					logger.WithFields(log.Fields{
						"OnuId":   onu.ID,
						"UniId":   uni.ID,
						"IntfId":  onu.PonPortID,
						"OnuSn":   onu.Sn(),
						"Service": service.Name,
						"Uni":     uni.ID,
					}).Debugf("Sending %s event on Service %s", event, service.Name)

					if err := service.IGMPState.Event(event, types.IgmpMessage{GroupAddress: req.GroupAddress}); err != nil {
						logger.WithFields(log.Fields{
							"OnuId":   onu.ID,
							"UniId":   uni.ID,
							"IntfId":  onu.PonPortID,
							"OnuSn":   onu.Sn(),
							"Service": service.Name,
						}).Errorf("IGMP request failed: %s", err.Error())
						errors = append(errors, fmt.Sprintf("%s: %s", serviceKey, err.Error()))
						success = false
					}
					startedOn = append(startedOn, serviceKey)
				}
			}
		}

		if success {
			res.StatusCode = int32(codes.OK)
			if len(startedOn) > 0 {
				res.Message = fmt.Sprintf("IGMP %s sent on Services %s for ONU %s.",
					event, fmt.Sprintf("%v", startedOn), onu.Sn())
			} else {
				res.Message = "No service requires IGMP"
			}
			logger.WithFields(log.Fields{
				"OnuSn":        req.OnuReq.SerialNumber,
				"subAction":    req.SubActionVal,
				"GroupAddress": req.GroupAddress,
				"Message":      res.Message,
			}).Info("Processed IGMP request for ONU")
		} else {
			res.StatusCode = int32(codes.FailedPrecondition)
			res.Message = fmt.Sprintf("%v", errors)
			logger.WithFields(log.Fields{
				"OnuSn":        req.OnuReq.SerialNumber,
				"subAction":    req.SubActionVal,
				"GroupAddress": req.GroupAddress,
				"Message":      res.Message,
			}).Error("Error while processing IGMP request for ONU")
		}

	}

	return res, nil
}

func (s BBSimServer) RestartEapol(ctx context.Context, req *bbsim.UNIRequest) (*bbsim.Response, error) {
	// NOTE this API will change the EAPOL state for all UNIs on the requested ONU if no UNI is specified
	// Otherwise, it will change the EAPOL state for only the specified UNI on the ONU

	res := &bbsim.Response{}

	if req.UniID == "" {
		logger.WithFields(log.Fields{
			"OnuSn": req.OnuSerialNumber,
		}).Infof("Received request to restart authentication for all UNIs on the ONU")
	} else {
		logger.WithFields(log.Fields{
			"OnuSn": req.OnuSerialNumber,
			"UniId": req.UniID,
		}).Infof("Received request to restart authentication on UNI")
	}

	olt := devices.GetOLT()

	onu, err := olt.FindOnuBySn(req.OnuSerialNumber)

	if err != nil {
		res.StatusCode = int32(codes.NotFound)
		res.Message = err.Error()
		return res, err
	}

	errors := []string{}
	startedOn := []string{}
	success := true

	uniIDint, err := strconv.Atoi(req.UniID)
	for _, u := range onu.UniPorts {
		uni := u.(*devices.UniPort)
		//if a specific uni is specified, only restart it
		if err == nil && req.UniID != "" && uni.ID != uint32(uniIDint) {
			continue
		}
		if !uni.OperState.Is(devices.UniStateUp) {
			// if the UNI is disabled, ignore it
			continue
		}
		for _, s := range uni.Services {
			service := s.(*devices.Service)
			serviceKey := fmt.Sprintf("uni[%d]%s", uni.ID, service.Name)
			if service.NeedsEapol {
				if !service.InternalState.Is(devices.ServiceStateInitialized) {
					logger.WithFields(log.Fields{
						"OnuId":   onu.ID,
						"UniId":   uni.ID,
						"IntfId":  onu.PonPortID,
						"OnuSn":   onu.Sn(),
						"Service": service.Name,
					}).Warn("service-not-initialized-skipping-event")
					continue
				}
				if err := service.EapolState.Event("start_auth"); err != nil {
					logger.WithFields(log.Fields{
						"OnuId":   onu.ID,
						"IntfId":  onu.PonPortID,
						"OnuSn":   onu.Sn(),
						"UniId":   uni.ID,
						"Service": service.Name,
					}).Errorf("Cannot restart authenticaton for Service: %s", err.Error())
					errors = append(errors, fmt.Sprintf("%s: %s", serviceKey, err.Error()))
					success = false
				}
				startedOn = append(startedOn, serviceKey)
			}
		}
	}

	if success {
		res.StatusCode = int32(codes.OK)
		if len(startedOn) > 0 {
			res.Message = fmt.Sprintf("Authentication restarted on Services %s for ONU %s.",
				fmt.Sprintf("%v", startedOn), onu.Sn())
		} else {
			res.Message = "No service requires EAPOL"
		}

		if req.UniID == "" {
			logger.WithFields(log.Fields{
				"OnuSn":   req.OnuSerialNumber,
				"Message": res.Message,
			}).Info("Processed EAPOL restart request for all UNIs on the ONU")
		} else {
			logger.WithFields(log.Fields{
				"OnuSn":   req.OnuSerialNumber,
				"Message": res.Message,
				"UniId":   req.UniID,
			}).Info("Processed EAPOL restart request on UNI")
		}

	} else {
		res.StatusCode = int32(codes.FailedPrecondition)
		res.Message = fmt.Sprintf("%v", errors)
		logger.WithFields(log.Fields{
			"OnuSn":   req.OnuSerialNumber,
			"Message": res.Message,
		}).Error("Error while processing EAPOL restart request for ONU")

		if req.UniID == "" {
			logger.WithFields(log.Fields{
				"OnuSn":   req.OnuSerialNumber,
				"Message": res.Message,
			}).Error("Error while processing EAPOL restart request for all UNIs on the ONU")
		} else {
			logger.WithFields(log.Fields{
				"OnuSn":   req.OnuSerialNumber,
				"Message": res.Message,
				"UniId":   req.UniID,
			}).Error("Error while processing EAPOL restart request on UNI")
		}

	}

	return res, nil
}

func (s BBSimServer) RestartDhcp(ctx context.Context, req *bbsim.UNIRequest) (*bbsim.Response, error) {
	// NOTE this API will change the DHCP state for all UNIs on the requested ONU if no UNI is specified
	// Otherwise, it will change the DHCP state for only the specified UNI on the ONU

	res := &bbsim.Response{}

	if req.UniID == "" {
		logger.WithFields(log.Fields{
			"OnuSn": req.OnuSerialNumber,
		}).Infof("Received request to restart authentication for all UNIs on the ONU")
	} else {
		logger.WithFields(log.Fields{
			"OnuSn": req.OnuSerialNumber,
			"UniId": req.UniID,
		}).Infof("Received request to restart authentication on UNI")
	}

	olt := devices.GetOLT()

	onu, err := olt.FindOnuBySn(req.OnuSerialNumber)

	if err != nil {
		res.StatusCode = int32(codes.NotFound)
		res.Message = err.Error()
		return res, err
	}

	errors := []string{}
	startedOn := []string{}
	success := true

	uniIDint, err := strconv.Atoi(req.UniID)
	for _, u := range onu.UniPorts {
		uni := u.(*devices.UniPort)
		//if a specific uni is specified, only restart it
		if err == nil && req.UniID != "" && uni.ID != uint32(uniIDint) {
			continue
		}
		if !uni.OperState.Is(devices.UniStateUp) {
			// if the UNI is disabled, ignore it
			continue
		}
		for _, s := range uni.Services {
			service := s.(*devices.Service)
			serviceKey := fmt.Sprintf("uni[%d]%s", uni.ID, service.Name)
			if service.NeedsDhcp {
				if err := service.DHCPState.Event("start_dhcp"); err != nil {
					logger.WithFields(log.Fields{
						"OnuId":   onu.ID,
						"IntfId":  onu.PonPortID,
						"OnuSn":   onu.Sn(),
						"UniId":   uni.ID,
						"Service": service.Name,
					}).Errorf("Cannot restart DHCP for Service: %s", err.Error())
					errors = append(errors, fmt.Sprintf("%s: %s", serviceKey, err.Error()))
					success = false
				}
				startedOn = append(startedOn, serviceKey)
			}
		}
	}

	if success {
		res.StatusCode = int32(codes.OK)
		if len(startedOn) > 0 {
			res.Message = fmt.Sprintf("DHCP restarted on Services %s for ONU %s.",
				fmt.Sprintf("%v", startedOn), onu.Sn())

		} else {
			res.Message = "No service requires DHCP"
		}

		if req.UniID == "" {
			logger.WithFields(log.Fields{
				"OnuSn":   req.OnuSerialNumber,
				"Message": res.Message,
			}).Info("Processed DHCP restart request for all UNIs on the ONU")
		} else {
			logger.WithFields(log.Fields{
				"OnuSn":   req.OnuSerialNumber,
				"Message": res.Message,
				"UniId":   req.UniID,
			}).Info("Processed DHCP restart request on UNI")
		}
	} else {
		res.StatusCode = int32(codes.FailedPrecondition)
		res.Message = fmt.Sprintf("%v", errors)

		if req.UniID == "" {
			logger.WithFields(log.Fields{
				"OnuSn":   req.OnuSerialNumber,
				"Message": res.Message,
			}).Error("Error while processing DHCP restart request for all UNIs on the ONU")
		} else {
			logger.WithFields(log.Fields{
				"OnuSn":   req.OnuSerialNumber,
				"Message": res.Message,
				"UniId":   req.UniID,
			}).Error("Error while processing DHCP restart request on UNI")
		}
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
		olt.Flows.Range(func(flowKey, flow interface{}) bool {
			flowObj := flow.(openolt.Flow)
			res.Flows = append(res.Flows, &flowObj)
			return true
		})
		res.FlowCount = uint32(len(res.Flows))
	} else {
		onu, err := olt.FindOnuBySn(req.SerialNumber)
		if err != nil {
			logger.WithFields(log.Fields{
				"OnuSn": req.SerialNumber,
			}).Error("Can't get ONU in GetFlows request")
			return nil, err
		}
		for _, flowKey := range onu.Flows {
			flow, _ := olt.Flows.Load(flowKey)
			flowObj := flow.(openolt.Flow)
			res.Flows = append(res.Flows, &flowObj)
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

	if err := onu.HandlePowerOnONU(); err != nil {
		res.StatusCode = int32(codes.FailedPrecondition)
		res.Message = err.Error()
		return res, err
	}

	res.StatusCode = int32(codes.OK)
	res.Message = fmt.Sprintf("ONU %s successfully powered on.", onu.Sn())

	return res, nil
}

func handleShutdownONU(onu *devices.Onu) (*bbsim.Response, error) {
	res := &bbsim.Response{}

	if err := onu.HandleShutdownONU(); err != nil {
		res.StatusCode = int32(codes.FailedPrecondition)
		res.Message = err.Error()
		return res, err
	}

	res.StatusCode = int32(codes.OK)
	res.Message = fmt.Sprintf("ONU %s successfully shut down.", onu.Sn())

	return res, nil
}

func (s BBSimServer) GetUnis(ctx context.Context, req *bbsim.Empty) (*bbsim.UNIs, error) {
	onus, err := s.GetONUs(ctx, req)

	if err != nil {
		return nil, err
	}
	unis := []*bbsim.UNI{}
	for _, onu := range onus.Items {
		unis = append(unis, onu.Unis...)
	}
	unis_ret := bbsim.UNIs{
		Items: unis,
	}
	return &unis_ret, nil
}

// Invalidate the MDS counter of the ONU
func (s BBSimServer) InvalidateMds(ctx context.Context, req *bbsim.ONURequest) (*bbsim.Response, error) {
	logger.WithFields(log.Fields{
		"OnuSn": req.SerialNumber,
	}).Infof("Received request to invalidate the MDS counter of the ONU")

	res := &bbsim.Response{}
	olt := devices.GetOLT()

	onu, err := olt.FindOnuBySn(req.SerialNumber)
	if err != nil {
		res.StatusCode = int32(codes.NotFound)
		res.Message = err.Error()
		return res, err
	}

	previous := onu.MibDataSync
	onu.InvalidateMibDataSync()

	res.StatusCode = int32(codes.OK)
	res.Message = fmt.Sprintf("MDS counter of ONU %s was %d, set to %d).", onu.Sn(), previous, onu.MibDataSync)

	return res, nil
}
