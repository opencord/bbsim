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

package devices

import (
	"bytes"
	"fmt"

	"github.com/looplab/fsm"
	"github.com/opencord/voltha-protos/v4/go/openolt"
	log "github.com/sirupsen/logrus"
)

type PonPort struct {
	// BBSIM Internals
	ID            uint32
	NumOnu        int
	Onus          []*Onu
	Olt           *OltDevice
	PacketCount   uint64
	InternalState *fsm.FSM

	// PON Attributes
	OperState *fsm.FSM
	Type      string
}

// CreatePonPort creates pon port object
func CreatePonPort(olt *OltDevice, id uint32) *PonPort {

	ponPort := PonPort{
		NumOnu: olt.NumOnuPerPon,
		ID:     id,
		Type:   "pon",
		Olt:    olt,
		Onus:   []*Onu{},
	}

	ponPort.InternalState = fsm.NewFSM(
		"created",
		fsm.Events{
			{Name: "enable", Src: []string{"created", "disabled"}, Dst: "enabled"},
			{Name: "disable", Src: []string{"enabled"}, Dst: "disabled"},
		},
		fsm.Callbacks{
			"enter_enabled": func(e *fsm.Event) {
				oltLogger.WithFields(log.Fields{
					"ID": ponPort.ID,
				}).Debugf("Changing PON Port InternalState from %s to %s", e.Src, e.Dst)

				if e.Src == "created" {
					if olt.ControlledActivation == Default || olt.ControlledActivation == OnlyPON {
						for _, onu := range ponPort.Onus {
							if err := onu.InternalState.Event(OnuTxInitialize); err != nil {
								log.Errorf("Error initializing ONU: %v", err)
								continue
							}
							if err := onu.InternalState.Event(OnuTxDiscover); err != nil {
								log.Errorf("Error discover ONU: %v", err)
							}
						}
					}
				} else if e.Src == "disabled" {
					if ponPort.Olt.ControlledActivation == OnlyONU || ponPort.Olt.ControlledActivation == Both {
						// if ONUs are manually activated then only initialize them
						for _, onu := range ponPort.Onus {
							if err := onu.InternalState.Event(OnuTxInitialize); err != nil {
								log.WithFields(log.Fields{
									"Err":    err,
									"OnuSn":  onu.Sn(),
									"IntfId": onu.PonPortID,
								}).Error("Error initializing ONU")
								continue
							}
						}
					} else {
						for _, onu := range ponPort.Onus {
							if onu.InternalState.Current() == OnuStatePonDisabled {
								if err := onu.InternalState.Event(OnuStateEnabled); err != nil {
									log.WithFields(log.Fields{
										"Err":    err,
										"OnuSn":  onu.Sn(),
										"IntfId": onu.PonPortID,
									}).Error("Error enabling ONU")
								}
							} else if onu.InternalState.Current() == OnuStateDisabled {
								if err := onu.InternalState.Event(OnuTxInitialize); err != nil {
									log.WithFields(log.Fields{
										"Err":    err,
										"OnuSn":  onu.Sn(),
										"IntfId": onu.PonPortID,
									}).Error("Error initializing ONU")
									continue
								}
								if err := onu.InternalState.Event(OnuTxDiscover); err != nil {
									log.WithFields(log.Fields{
										"Err":    err,
										"OnuSn":  onu.Sn(),
										"IntfId": onu.PonPortID,
									}).Error("Error discovering ONU")
								}
							} else if onu.InternalState.Current() == OnuStateInitialized {
								if err := onu.InternalState.Event(OnuTxDiscover); err != nil {
									log.WithFields(log.Fields{
										"Err":    err,
										"OnuSn":  onu.Sn(),
										"IntfId": onu.PonPortID,
									}).Error("Error discovering ONU")
								}
							} else {
								// this is to loudly report unexpected states in order to address them
								log.WithFields(log.Fields{
									"OnuSn":         onu.Sn(),
									"IntfId":        onu.PonPortID,
									"InternalState": onu.InternalState.Current(),
								}).Error("Unexpected ONU state in PON enabling")
							}
						}
					}
				}
			},
			"enter_disabled": func(e *fsm.Event) {
				for _, onu := range ponPort.Onus {
					if onu.InternalState.Current() == OnuStateInitialized || onu.InternalState.Current() == OnuStateDisabled {
						continue
					}
					if err := onu.InternalState.Event(OnuTxPonDisable); err != nil {
						oltLogger.Errorf("Failed to move ONU in %s states: %v", OnuStatePonDisabled, err)
					}
				}
			},
		},
	)

	ponPort.OperState = fsm.NewFSM(
		"down",
		fsm.Events{
			{Name: "enable", Src: []string{"down"}, Dst: "up"},
			{Name: "disable", Src: []string{"up"}, Dst: "down"},
		},
		fsm.Callbacks{
			"enter_up": func(e *fsm.Event) {
				oltLogger.WithFields(log.Fields{
					"ID": ponPort.ID,
				}).Debugf("Changing PON Port OperState from %s to %s", e.Src, e.Dst)
				olt.sendPonIndication(ponPort.ID)
			},
			"enter_down": func(e *fsm.Event) {
				oltLogger.WithFields(log.Fields{
					"ID": ponPort.ID,
				}).Debugf("Changing PON Port OperState from %s to %s", e.Src, e.Dst)
				olt.sendPonIndication(ponPort.ID)
			},
		},
	)
	return &ponPort
}

func (p PonPort) GetOnuBySn(sn *openolt.SerialNumber) (*Onu, error) {
	for _, onu := range p.Onus {
		if bytes.Equal(onu.SerialNumber.VendorSpecific, sn.VendorSpecific) {
			return onu, nil
		}
	}
	return nil, fmt.Errorf("Cannot find Onu with serial number %d in PonPort %d", sn, p.ID)
}

func (p PonPort) GetOnuById(id uint32) (*Onu, error) {
	for _, onu := range p.Onus {
		if onu.ID == id {
			return onu, nil
		}
	}
	return nil, fmt.Errorf("Cannot find Onu with id %d in PonPort %d", id, p.ID)
}

// GetNumOfActiveOnus returns number of active ONUs for PON port
func (p PonPort) GetNumOfActiveOnus() uint32 {
	var count uint32 = 0
	for _, onu := range p.Onus {
		if onu.InternalState.Current() == OnuStateInitialized || onu.InternalState.Current() == OnuStateCreated || onu.InternalState.Current() == OnuStateDisabled {
			continue
		}
		count++
	}
	return count
}
