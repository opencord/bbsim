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
				STag:          int32(o.STag),
				CTag:          int32(o.CTag),
				HwAddress:     o.HwAddress.String(),
				PortNo:        int32(o.PortNo),
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
		STag:          int32(onu.STag),
		CTag:          int32(onu.CTag),
		HwAddress:     onu.HwAddress.String(),
		PortNo:        int32(onu.PortNo),
	}
	return &res, nil
}

func (s BBSimServer) ShutdownONU(ctx context.Context, req *bbsim.ONURequest) (*bbsim.Response, error) {
	// NOTE this method is now sendying a Dying Gasp and then disabling the device (operState: down, adminState: up),
	// is this the only way to do? Should we address other cases?
	// Investigate what happens when:
	// - a fiber is pulled
	// - ONU malfunction
	// - ONU shutdown
	res := &bbsim.Response{}

	logger.WithFields(log.Fields{
		"OnuSn": req.SerialNumber,
	}).Infof("Received request to shutdown ONU")

	olt := devices.GetOLT()

	onu, err := olt.FindOnuBySn(req.SerialNumber)

	if err != nil {
		res.StatusCode = int32(codes.NotFound)
		res.Message = err.Error()
		return res, err
	}

	dyingGasp := devices.Message{
		Type: devices.DyingGaspIndication,
		Data: devices.DyingGaspIndicationMessage{
			OnuID:     onu.ID,
			PonPortID: onu.PonPortID,
			Status:    "on", // TODO do we need a type for Dying Gasp Indication?
		},
	}

	onu.Channel <- dyingGasp

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

func (s BBSimServer) PoweronONU(ctx context.Context, req *bbsim.ONURequest) (*bbsim.Response, error) {
	res := &bbsim.Response{}

	logger.WithFields(log.Fields{
		"OnuSn": req.SerialNumber,
	}).Infof("Received request to poweron ONU")

	olt := devices.GetOLT()

	onu, err := olt.FindOnuBySn(req.SerialNumber)

	if err != nil {
		res.StatusCode = int32(codes.NotFound)
		res.Message = err.Error()
		return res, err
	}

	if err := onu.InternalState.Event("enable"); err != nil {
		logger.WithFields(log.Fields{
			"OnuId":  onu.ID,
			"IntfId": onu.PonPortID,
			"OnuSn":  onu.Sn(),
		}).Errorf("Cannot poweron ONU: %s", err.Error())
		res.StatusCode = int32(codes.FailedPrecondition)
		res.Message = err.Error()
		return res, err
	}

	res.StatusCode = int32(codes.OK)
	res.Message = fmt.Sprintf("ONU %s successfully powered on.", onu.Sn())

	return res, nil
}

func (s BBSimServer) ChangeIgmpState(ctx context.Context, req *bbsim.IgmpRequest) (*bbsim.Response, error) {
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

		}
		igmpErr := onu.InternalState.Event(event)

		if igmpErr != nil {
			logger.WithFields(log.Fields{
				"OnuId":  onu.ID,
				"IntfId": onu.PonPortID,
				"OnuSn":  onu.Sn(),
			}).Errorf("IGMP request failed: %s", igmpErr.Error())
			res.StatusCode = int32(codes.FailedPrecondition)
			res.Message = err.Error()
			return res, igmpErr
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

	if err := onu.InternalState.Event("start_auth"); err != nil {
		logger.WithFields(log.Fields{
			"OnuId":  onu.ID,
			"IntfId": onu.PonPortID,
			"OnuSn":  onu.Sn(),
		}).Errorf("Cannot restart authenticaton for ONU: %s", err.Error())
		res.StatusCode = int32(codes.FailedPrecondition)
		res.Message = err.Error()
		return res, err
	}

	res.StatusCode = int32(codes.OK)
	res.Message = fmt.Sprintf("Authentication restarted for ONU %s.", onu.Sn())

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

	if err := onu.InternalState.Event("start_dhcp"); err != nil {
		logger.WithFields(log.Fields{
			"OnuId":  onu.ID,
			"IntfId": onu.PonPortID,
			"OnuSn":  onu.Sn(),
		}).Errorf("Cannot restart DHCP for ONU: %s", err.Error())
		res.StatusCode = int32(codes.FailedPrecondition)
		res.Message = err.Error()
		return res, err
	}

	res.StatusCode = int32(codes.OK)
	res.Message = fmt.Sprintf("DHCP restarted for ONU %s.", onu.Sn())

	return res, nil
}
