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
	"encoding/hex"
	"errors"
	"net"
	"strconv"

	"github.com/opencord/bbsim/api/legacy"
	api "github.com/opencord/bbsim/api/legacy"
	"github.com/opencord/bbsim/internal/bbsim/devices"
	"github.com/opencord/voltha-protos/v2/go/openolt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Constants for reboot delays
const (
	OltRebootDelay     = 40
	OnuSoftRebootDelay = 10
	OnuHardRebootDelay = 30
)

// handleONUStatusRequest process ONU status request
func (s BBSimLegacyServer) handleONUStatusRequest(in *api.ONUInfo) (*api.ONUs, error) {
	logger.Trace("handleONUStatusRequest() invoked")
	onuInfo := &api.ONUs{}
	olt := devices.GetOLT()

	if in.OnuSerial != "" { // Get status of a single ONU by SerialNumber
		onu, err := olt.FindOnuBySn(in.OnuSerial)
		if err != nil {
			logger.Errorf("ONU error: %+v", err)
			return onuInfo, status.Errorf(codes.NotFound, "Unable to retrieve ONU %s", in.OnuSerial)
		}
		onuInfo.Onus = append(onuInfo.Onus, copyONUInfo(onu))
	} else {
		if in.OnuId != 0 { // Get status of a single ONU by ONU-ID
			onu, err := olt.FindOnuById(in.PonPortId, in.OnuId)
			if err != nil {
				logger.Errorf("ONU error: %+v", err)
				return onuInfo, status.Errorf(codes.NotFound, "Unable to retrieve ONU with ID %d", in.OnuId)
			}

			onuInfo.Onus = append(onuInfo.Onus, copyONUInfo(onu))
		} else { // Get status of all ONUs
			for _, p := range olt.Pons {
				for _, o := range p.Onus {
					onuInfo.Onus = append(onuInfo.Onus, copyONUInfo(o))
				}
			}
		}
	}

	return onuInfo, nil
}

func copyONUInfo(onu *devices.Onu) *api.ONUInfo {
	onuData := &api.ONUInfo{
		OnuId:     onu.ID,
		PonPortId: onu.PonPortID,
		OnuSerial: onu.Sn(),
		OnuState:  onu.InternalState.Current(),
		OperState: onu.OperState.Current(),
	}

	return onuData
}

func (s BBSimLegacyServer) fetchPortDetail(intfID uint32, portType string) (*api.PortInfo, error) {
	logger.Tracef("fetchPortDetail() invoked %s-%d", portType, intfID)
	olt := devices.GetOLT()

	switch portType {
	case "pon":
		for _, pon := range olt.Pons {
			if pon.ID == intfID {
				ponPortInfo := &api.PortInfo{
					PortType:          "pon",
					PortId:            uint32(pon.ID),
					PonPortMaxOnus:    uint32(olt.NumOnuPerPon),
					PonPortActiveOnus: uint32(olt.NumPon),
					PortState:         pon.OperState.Current(),
				}
				return ponPortInfo, nil
			}
		}
	case "nni":
		for _, nni := range olt.Nnis {
			if nni.ID == intfID {
				nniPortInfo := &legacy.PortInfo{
					PortType:  "nni",
					PortId:    uint32(nni.ID),
					PortState: nni.OperState.Current(),
				}
				return nniPortInfo, nil
			}
		}
	}

	portInfo := &api.PortInfo{}
	return portInfo, status.Errorf(codes.NotFound, "Invalid port %s-%d", portType, intfID)
}

func getOpenoltSerialNumber(SerialNumber string) (*openolt.SerialNumber, error) {
	var VendorIDLength = 4
	var SerialNumberLength = 12

	if len(SerialNumber) != SerialNumberLength {
		logger.Error("Invalid serial number %s", SerialNumber)
		return nil, errors.New("invalid serial number")
	}
	// First four characters are vendorId
	vendorID := SerialNumber[:VendorIDLength]
	vendorSpecific := SerialNumber[VendorIDLength:]

	vsbyte, _ := hex.DecodeString(vendorSpecific)

	// Convert to Openolt serial number
	serialNum := new(openolt.SerialNumber)
	serialNum.VendorId = []byte(vendorID)
	serialNum.VendorSpecific = vsbyte

	return serialNum, nil
}

// ConvB2S converts byte array to string
func ConvB2S(b []byte) string {
	s := ""
	for _, i := range b {
		s = s + strconv.FormatInt(int64(i/16), 16) + strconv.FormatInt(int64(i%16), 16)
	}
	return s
}

func getOltIP() net.IP {
	// TODO make this better
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		logger.Error(err.Error())
		return net.IP{}
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			logger.Error(err.Error())
		}
	}()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
