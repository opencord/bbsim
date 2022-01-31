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
	"math/rand"
	"time"

	"github.com/looplab/fsm"
	"github.com/opencord/bbsim/internal/common"
	"github.com/opencord/voltha-protos/v5/go/openolt"
)

type mode int

// Constants for Controlled Activation modes
const (
	Default mode = iota
	OnlyONU
	OnlyPON
	Both
)

// ControlledActivationModes maps string to int value of mode
var ControlledActivationModes = map[string]mode{
	"default":  Default,
	"only-onu": OnlyONU,
	"only-pon": OnlyPON,
	"both":     Both,
}

//Constants for openolt Flows
const (
	flowTypeUpstream   = "upstream"
	flowTypeDownstream = "downstream"
	flowTagTypeSingle  = "single_tag"
	flowTagTypeDouble  = "double_tag"
)

var newFSM = fsm.NewFSM

func getOperStateFSM(cb fsm.Callback) *fsm.FSM {
	return newFSM(
		"down",
		fsm.Events{
			{Name: "enable", Src: []string{"down"}, Dst: "up"},
			{Name: "disable", Src: []string{"up"}, Dst: "down"},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) {
				cb(e)
			},
		},
	)
}

func publishEvent(eventType string, intfID int32, onuID int32, onuSerial string) {
	if olt.PublishEvents {
		currentTime := time.Now()

		event := common.Event{
			EventType: eventType,
			OltID:     olt.ID,
			IntfID:    intfID,
			OnuID:     onuID,
			OnuSerial: onuSerial,
			Timestamp: currentTime.Format("2006-01-02 15:04:05.000000000"),
			EpochTime: currentTime.UnixNano() / 1000000,
		}
		olt.EventChannel <- event
	}
}

func getPortStats(packetCount uint64, incrementStat bool) (*openolt.PortStatistics, uint64) {
	// increment current packet count by random number
	if incrementStat {
		packetCount = packetCount + uint64(rand.Intn(50)+1*10)
	}

	// fill all other stats based on packet count
	portStats := &openolt.PortStatistics{
		RxBytes:        packetCount * 64,
		RxPackets:      packetCount,
		RxUcastPackets: packetCount * 40 / 100,
		RxMcastPackets: packetCount * 30 / 100,
		RxBcastPackets: packetCount * 30 / 100,
		RxErrorPackets: 0,
		TxBytes:        packetCount * 64,
		TxPackets:      packetCount,
		TxUcastPackets: packetCount * 40 / 100,
		TxMcastPackets: packetCount * 30 / 100,
		TxBcastPackets: packetCount * 30 / 100,
		TxErrorPackets: 0,
		RxCrcErrors:    0,
		BipErrors:      0,
		Timestamp:      uint32(time.Now().Unix()),
	}

	return portStats, packetCount
}

// InterfaceIDToPortNo converts InterfaceID to voltha PortID
// Refer openolt adapter code(master) voltha-openolt-adapter/adaptercore/olt_platform.go: IntfIDToPortNo()
func InterfaceIDToPortNo(intfID uint32, intfType string) uint32 {
	// Converts interface-id to port-numbers that can be understood by the VOLTHA
	if intfType == "nni" {
		// nni at voltha starts with 1,048,576
		// nni = 1,048,576 + InterfaceID
		return 0x1<<20 + intfID
	} else if intfType == "pon" {
		// pon = 536,870,912 + InterfaceID
		return (0x2 << 28) + intfID
	}
	return 0
}
