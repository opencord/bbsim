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
			}
			onus.Items = append(onus.Items, &onu)
		}
	}
	return &onus, nil
}

func (s BBSimServer) DisableONU(ctx context.Context, req *bbsim.ONURequest) (*bbsim.Response, error) {
	// NOTE this method is now sendying a Dying Gasp and then disabling the device,
	// is this the only way to do? Should we address other cases?
	// Investigate what happens when:
	// - a fiber is pulled
	// - ONU malfunction
	// - ONU shutdown
	res := &bbsim.Response{}

	logger.WithFields(log.Fields{
		"OnuSn": req.SerialNumber,
	}).Infof("Received request to disable ONU")

	olt := devices.GetOLT()

	onu, err := olt.FindOnu(req.SerialNumber)

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
		}).Errorf("Cannot disable ONU: %s", err.Error())
		res.StatusCode = int32(codes.FailedPrecondition)
		res.Message = err.Error()
		return res, err
	}

	res.StatusCode = int32(codes.OK)
	res.Message = "Success!"

	return res, nil
}
