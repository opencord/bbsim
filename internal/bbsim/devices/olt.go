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
	"context"
	"encoding/hex"
	"fmt"
	"github.com/opencord/bbsim/internal/bbsim/responders/dhcp"
	"github.com/opencord/voltha-protos/v4/go/ext/config"
	"net"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/looplab/fsm"
	"github.com/opencord/bbsim/internal/bbsim/packetHandlers"
	"github.com/opencord/bbsim/internal/common"
	omcisim "github.com/opencord/omci-sim"
	common_protos "github.com/opencord/voltha-protos/v4/go/common"
	"github.com/opencord/voltha-protos/v4/go/openolt"
	"github.com/opencord/voltha-protos/v4/go/tech_profile"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var oltLogger = log.WithFields(log.Fields{
	"module": "OLT",
})

type OltDevice struct {
	sync.Mutex
	oltServer *grpc.Server

	// BBSIM Internals
	ID                   int
	SerialNumber         string
	NumNni               int
	NumPon               int
	NumOnuPerPon         int
	InternalState        *fsm.FSM
	channel              chan Message
	dhcpServer           dhcp.DHCPServerIf
	Flows                map[FlowKey]openolt.Flow
	Delay                int
	ControlledActivation mode
	EventChannel         chan common.Event
	PublishEvents        bool
	PortStatsInterval    int

	Pons []*PonPort
	Nnis []*NniPort

	// OLT Attributes
	OperState *fsm.FSM

	enableContext       context.Context
	enableContextCancel context.CancelFunc

	OpenoltStream openolt.Openolt_EnableIndicationServer
	enablePerf    bool
}

var olt OltDevice

func GetOLT() *OltDevice {
	return &olt
}

func CreateOLT(options common.GlobalConfig, services []common.ServiceYaml, isMock bool) *OltDevice {
	oltLogger.WithFields(log.Fields{
		"ID":           options.Olt.ID,
		"NumNni":       options.Olt.NniPorts,
		"NumPon":       options.Olt.PonPorts,
		"NumOnuPerPon": options.Olt.OnusPonPort,
	}).Debug("CreateOLT")

	olt = OltDevice{
		ID:           options.Olt.ID,
		SerialNumber: fmt.Sprintf("BBSIM_OLT_%d", options.Olt.ID),
		OperState: getOperStateFSM(func(e *fsm.Event) {
			oltLogger.Debugf("Changing OLT OperState from %s to %s", e.Src, e.Dst)
		}),
		NumNni:            int(options.Olt.NniPorts),
		NumPon:            int(options.Olt.PonPorts),
		NumOnuPerPon:      int(options.Olt.OnusPonPort),
		Pons:              []*PonPort{},
		Nnis:              []*NniPort{},
		Delay:             options.BBSim.Delay,
		Flows:             make(map[FlowKey]openolt.Flow),
		enablePerf:        options.BBSim.EnablePerf,
		PublishEvents:     options.BBSim.Events,
		PortStatsInterval: options.Olt.PortStatsInterval,
		dhcpServer:        dhcp.NewDHCPServer(),
	}

	if val, ok := ControlledActivationModes[options.BBSim.ControlledActivation]; ok {
		olt.ControlledActivation = val
	} else {
		// FIXME throw an error if the ControlledActivation is not valid
		oltLogger.Warn("Unknown ControlledActivation Mode given, running in Default mode")
		olt.ControlledActivation = Default
	}

	// OLT State machine
	// NOTE do we need 2 state machines for the OLT? (InternalState and OperState)
	olt.InternalState = fsm.NewFSM(
		"created",
		fsm.Events{
			{Name: "initialize", Src: []string{"created", "deleted"}, Dst: "initialized"},
			{Name: "enable", Src: []string{"initialized", "disabled"}, Dst: "enabled"},
			{Name: "disable", Src: []string{"enabled"}, Dst: "disabled"},
			// delete event in enabled state below is for reboot OLT case.
			{Name: "delete", Src: []string{"disabled", "enabled"}, Dst: "deleted"},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) {
				oltLogger.Debugf("Changing OLT InternalState from %s to %s", e.Src, e.Dst)
			},
			"enter_initialized": func(e *fsm.Event) { olt.InitOlt() },
		},
	)

	if !isMock {
		// create NNI Port
		nniPort, err := CreateNNI(&olt)
		if err != nil {
			oltLogger.Fatalf("Couldn't create NNI Port: %v", err)
		}

		olt.Nnis = append(olt.Nnis, &nniPort)
	}

	// Create device and Services

	nextCtag := map[string]int{}
	nextStag := map[string]int{}

	// create PON ports
	for i := 0; i < olt.NumPon; i++ {
		p := CreatePonPort(&olt, uint32(i))

		// create ONU devices
		for j := 0; j < olt.NumOnuPerPon; j++ {
			delay := time.Duration(olt.Delay*j) * time.Millisecond
			o := CreateONU(&olt, p, uint32(j+1), delay, isMock)

			for k, s := range common.Services {

				// find the correct cTag for this service
				if _, ok := nextCtag[s.Name]; !ok {
					// it's the first time we iterate over this service,
					// so we start from the config value
					nextCtag[s.Name] = s.CTag
				} else {
					// we have a previous value, so we check it
					// if Allocation is unique, we increment,
					// otherwise (shared) we do nothing
					if s.CTagAllocation == common.TagAllocationUnique.String() {
						nextCtag[s.Name] = nextCtag[s.Name] + 1
					}
				}

				// find the correct sTag for this service
				if _, ok := nextStag[s.Name]; !ok {
					nextStag[s.Name] = s.STag
				} else {
					if s.STagAllocation == common.TagAllocationUnique.String() {
						nextStag[s.Name] = nextStag[s.Name] + 1
					}
				}

				mac := net.HardwareAddr{0x2e, 0x60, byte(olt.ID), byte(p.ID), byte(o.ID), byte(k)}
				service, err := NewService(s.Name, mac, o, nextCtag[s.Name], nextStag[s.Name],
					s.NeedsEapol, s.NeedsDchp, s.NeedsIgmp, s.TechnologyProfileID, s.UniTagMatch,
					s.ConfigureMacAddress, s.UsPonCTagPriority, s.UsPonSTagPriority, s.DsPonCTagPriority, s.DsPonSTagPriority)

				if err != nil {
					oltLogger.WithFields(log.Fields{
						"Err": err.Error(),
					}).Fatal("Can't create Service")
				}

				o.Services = append(o.Services, service)
			}
			p.Onus = append(p.Onus, o)
		}
		olt.Pons = append(olt.Pons, p)
	}

	if !isMock {
		if err := olt.InternalState.Event("initialize"); err != nil {
			log.Errorf("Error initializing OLT: %v", err)
			return nil
		}
	}

	if olt.PublishEvents {
		log.Debugf("BBSim event publishing is enabled")
		// Create a channel to write event messages
		olt.EventChannel = make(chan common.Event, 100)
	}

	return &olt
}

func (o *OltDevice) InitOlt() {

	if o.oltServer == nil {
		o.oltServer, _ = o.StartOltServer()
	} else {
		oltLogger.Fatal("OLT server already running.")
	}

	// create new channel for processOltMessages Go routine
	o.channel = make(chan Message)

	// FIXME we are assuming we have only one NNI
	if o.Nnis[0] != nil {
		// NOTE we want to make sure the state is down when we initialize the OLT,
		// the NNI may be in a bad state after a disable/reboot as we are not disabling it for
		// in-band management
		o.Nnis[0].OperState.SetState("down")
	}
}

func (o *OltDevice) RestartOLT() error {

	softReboot := false
	rebootDelay := common.Config.Olt.OltRebootDelay

	oltLogger.WithFields(log.Fields{
		"oltId": o.ID,
	}).Infof("Simulating OLT restart... (%ds)", rebootDelay)

	if o.InternalState.Is("enabled") {
		oltLogger.WithFields(log.Fields{
			"oltId": o.ID,
		}).Info("This is an OLT soft reboot")
		softReboot = true
	}

	// transition internal state to deleted
	if err := o.InternalState.Event("delete"); err != nil {
		oltLogger.WithFields(log.Fields{
			"oltId": o.ID,
		}).Errorf("Error deleting OLT: %v", err)
		return err
	}

	time.Sleep(1 * time.Second) // we need to give the OLT the time to respond to all the pending gRPC request before stopping the server
	o.StopOltServer()

	if softReboot {
		for _, pon := range o.Pons {
			if pon.InternalState.Current() == "enabled" {
				// disable PONs
				msg := Message{
					Type: PonIndication,
					Data: PonIndicationMessage{
						OperState: DOWN,
						PonPortID: pon.ID,
					},
				}
				o.channel <- msg
			}

			for _, onu := range pon.Onus {
				_ = onu.InternalState.Event("disable")
			}
		}
	} else {
		// PONs are already handled in the Disable call
		for _, pon := range olt.Pons {
			// ONUs are not automatically disabled when a PON goes down
			// as it's possible that it's an admin down and in that case the ONUs need to keep their state
			for _, onu := range pon.Onus {
				_ = onu.InternalState.Event("disable")
			}
		}
	}

	// terminate the OLT's processOltMessages go routine
	close(o.channel)

	o.enableContextCancel()

	time.Sleep(time.Duration(rebootDelay) * time.Second)

	if err := o.InternalState.Event("initialize"); err != nil {
		oltLogger.WithFields(log.Fields{
			"oltId": o.ID,
		}).Errorf("Error initializing OLT: %v", err)
		return err
	}
	oltLogger.WithFields(log.Fields{
		"oltId": o.ID,
	}).Info("OLT restart completed")
	return nil
}

// newOltServer launches a new grpc server for OpenOLT
func (o *OltDevice) newOltServer() (*grpc.Server, error) {
	address := common.Config.BBSim.OpenOltAddress
	lis, err := net.Listen("tcp", address)
	if err != nil {
		oltLogger.Fatalf("OLT failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()

	openolt.RegisterOpenoltServer(grpcServer, o)

	reflection.Register(grpcServer)

	go func() { _ = grpcServer.Serve(lis) }()
	oltLogger.Debugf("OLT listening on %v", address)

	return grpcServer, nil
}

// StartOltServer will create the grpc server that VOLTHA uses
// to communicate with the device
func (o *OltDevice) StartOltServer() (*grpc.Server, error) {
	oltServer, err := o.newOltServer()
	if err != nil {
		oltLogger.WithFields(log.Fields{
			"err": err,
		}).Error("Cannot OLT gRPC server")
		return nil, err
	}
	return oltServer, nil
}

// StopOltServer stops the OpenOLT grpc server
func (o *OltDevice) StopOltServer() {
	if o.oltServer != nil {
		oltLogger.WithFields(log.Fields{
			"oltId": o.SerialNumber,
		}).Warnf("Stopping OLT gRPC server")
		o.oltServer.Stop()
		o.oltServer = nil
	}
}

// Device Methods

// Enable implements the OpenOLT EnableIndicationServer functionality
func (o *OltDevice) Enable(stream openolt.Openolt_EnableIndicationServer) {
	oltLogger.Debug("Enable OLT called")
	rebootFlag := false

	// If enabled has already been called then an enabled context has
	// been created. If this is the case then we want to cancel all the
	// proessing loops associated with that enable before we recreate
	// new ones
	o.Lock()
	if o.enableContext != nil && o.enableContextCancel != nil {
		oltLogger.Info("This is an OLT reboot")
		o.enableContextCancel()
		rebootFlag = true
	}
	o.enableContext, o.enableContextCancel = context.WithCancel(context.TODO())
	o.Unlock()

	wg := sync.WaitGroup{}
	wg.Add(3)

	o.OpenoltStream = stream

	// create Go routine to process all OLT events
	go o.processOltMessages(o.enableContext, stream, &wg)

	// enable the OLT
	oltMsg := Message{
		Type: OltIndication,
		Data: OltIndicationMessage{
			OperState: UP,
		},
	}
	o.channel <- oltMsg

	// send NNI Port Indications
	for _, nni := range o.Nnis {
		msg := Message{
			Type: NniIndication,
			Data: NniIndicationMessage{
				OperState: UP,
				NniPortID: nni.ID,
			},
		}
		o.channel <- msg
	}

	go o.processOmciMessages(o.enableContext, stream, &wg)

	if rebootFlag {
		for _, pon := range o.Pons {
			if pon.InternalState.Current() == "disabled" {
				msg := Message{
					Type: PonIndication,
					Data: PonIndicationMessage{
						OperState: UP,
						PonPortID: pon.ID,
					},
				}
				o.channel <- msg
			}
		}
	} else {

		// 1. controlledActivation == Default: Send both PON and ONUs indications
		// 2. controlledActivation == only-onu: that means only ONUs will be controlled activated, so auto send PON indications

		if o.ControlledActivation == Default || o.ControlledActivation == OnlyONU {
			// send PON Port indications
			for _, pon := range o.Pons {
				msg := Message{
					Type: PonIndication,
					Data: PonIndicationMessage{
						OperState: UP,
						PonPortID: pon.ID,
					},
				}
				o.channel <- msg
			}
		}
	}

	oltLogger.Debug("Enable OLT Done")

	if !o.enablePerf {
		// Start a go routine to send periodic port stats to openolt adapter
		go o.periodicPortStats(o.enableContext)
	}

	wg.Wait()
}

func (o *OltDevice) processOmciMessages(ctx context.Context, stream openolt.Openolt_EnableIndicationServer, wg *sync.WaitGroup) {
	ch := omcisim.GetChannel()

	oltLogger.Debug("Starting OMCI Indication Channel")

loop:
	for {
		select {
		case <-ctx.Done():
			oltLogger.Debug("OMCI processing canceled via context")
			break loop
		case message, ok := <-ch:
			if !ok || ctx.Err() != nil {
				oltLogger.Debug("OMCI processing canceled via channel close")
				break loop
			}

			oltLogger.WithFields(log.Fields{
				"messageType": message.Type,
				"OnuId":       message.Data.OnuId,
				"IntfId":      message.Data.IntfId,
			}).Debug("Received message on OMCI Sim channel")

			onuId := message.Data.OnuId
			intfId := message.Data.IntfId
			onu, err := o.FindOnuById(intfId, onuId)
			if err != nil {
				oltLogger.Errorf("Failed to find onu: %v", err)
				continue
			}
			go onu.processOmciMessage(message, stream)
		}
	}

	wg.Done()
}

func (o *OltDevice) periodicPortStats(ctx context.Context) {
	var portStats *openolt.PortStatistics
	for {
		select {
		case <-time.After(time.Duration(o.PortStatsInterval) * time.Second):
			// send NNI port stats
			for _, port := range o.Nnis {
				incrementStat := true
				if port.OperState.Current() == "down" {
					incrementStat = false
				}
				portStats, port.PacketCount = getPortStats(port.PacketCount, incrementStat)
				o.sendPortStatsIndication(portStats, port.ID, port.Type)
			}

			// send PON port stats
			for _, port := range o.Pons {
				incrementStat := true
				// do not increment port stats if PON port is down or no ONU is activated on PON port
				if port.OperState.Current() == "down" || port.GetNumOfActiveOnus() < 1 {
					incrementStat = false
				}
				portStats, port.PacketCount = getPortStats(port.PacketCount, incrementStat)
				o.sendPortStatsIndication(portStats, port.ID, port.Type)
			}
		case <-ctx.Done():
			log.Debug("Stop sending port stats")
			return
		}

	}
}

// Helpers method

func (o *OltDevice) GetPonById(id uint32) (*PonPort, error) {
	for _, pon := range o.Pons {
		if pon.ID == id {
			return pon, nil
		}
	}
	return nil, fmt.Errorf("Cannot find PonPort with id %d in OLT %d", id, o.ID)
}

func (o *OltDevice) getNniById(id uint32) (*NniPort, error) {
	for _, nni := range o.Nnis {
		if nni.ID == id {
			return nni, nil
		}
	}
	return nil, fmt.Errorf("Cannot find NniPort with id %d in OLT %d", id, o.ID)
}

func (o *OltDevice) sendAlarmIndication(alarmInd *openolt.AlarmIndication, stream openolt.Openolt_EnableIndicationServer) {
	data := &openolt.Indication_AlarmInd{AlarmInd: alarmInd}
	if err := stream.Send(&openolt.Indication{Data: data}); err != nil {
		oltLogger.Errorf("Failed to send Alarm Indication: %v", err)
		return
	}

	oltLogger.WithFields(log.Fields{
		"AlarmIndication": alarmInd,
	}).Debug("Sent Indication_AlarmInd")
}

func (o *OltDevice) sendOltIndication(msg OltIndicationMessage, stream openolt.Openolt_EnableIndicationServer) {
	data := &openolt.Indication_OltInd{OltInd: &openolt.OltIndication{OperState: msg.OperState.String()}}
	if err := stream.Send(&openolt.Indication{Data: data}); err != nil {
		oltLogger.Errorf("Failed to send Indication_OltInd: %v", err)
		return
	}

	oltLogger.WithFields(log.Fields{
		"OperState": msg.OperState,
	}).Debug("Sent Indication_OltInd")
}

func (o *OltDevice) sendNniIndication(msg NniIndicationMessage, stream openolt.Openolt_EnableIndicationServer) {
	nni, _ := o.getNniById(msg.NniPortID)
	if msg.OperState == UP {
		if err := nni.OperState.Event("enable"); err != nil {
			log.WithFields(log.Fields{
				"Type":      nni.Type,
				"IntfId":    nni.ID,
				"OperState": nni.OperState.Current(),
			}).Errorf("Can't move NNI Port to enabled state: %v", err)
		}
	} else if msg.OperState == DOWN {
		if err := nni.OperState.Event("disable"); err != nil {
			log.WithFields(log.Fields{
				"Type":      nni.Type,
				"IntfId":    nni.ID,
				"OperState": nni.OperState.Current(),
			}).Errorf("Can't move NNI Port to disable state: %v", err)
		}
	}
	// NOTE Operstate may need to be an integer
	operData := &openolt.Indication_IntfOperInd{IntfOperInd: &openolt.IntfOperIndication{
		Type:      nni.Type,
		IntfId:    nni.ID,
		OperState: nni.OperState.Current(),
	}}

	if err := stream.Send(&openolt.Indication{Data: operData}); err != nil {
		oltLogger.Errorf("Failed to send Indication_IntfOperInd for NNI: %v", err)
		return
	}

	oltLogger.WithFields(log.Fields{
		"Type":      nni.Type,
		"IntfId":    nni.ID,
		"OperState": nni.OperState.Current(),
	}).Debug("Sent Indication_IntfOperInd for NNI")
}

func (o *OltDevice) sendPonIndication(ponPortID uint32) {

	stream := o.OpenoltStream
	pon, _ := o.GetPonById(ponPortID)
	// Send IntfIndication for PON port
	discoverData := &openolt.Indication_IntfInd{IntfInd: &openolt.IntfIndication{
		IntfId:    pon.ID,
		OperState: pon.OperState.Current(),
	}}

	if err := stream.Send(&openolt.Indication{Data: discoverData}); err != nil {
		oltLogger.Errorf("Failed to send Indication_IntfInd: %v", err)
		return
	}

	oltLogger.WithFields(log.Fields{
		"IntfId":    pon.ID,
		"OperState": pon.OperState.Current(),
	}).Debug("Sent Indication_IntfInd for PON")

	// Send IntfOperIndication for PON port
	operData := &openolt.Indication_IntfOperInd{IntfOperInd: &openolt.IntfOperIndication{
		Type:      pon.Type,
		IntfId:    pon.ID,
		OperState: pon.OperState.Current(),
	}}

	if err := stream.Send(&openolt.Indication{Data: operData}); err != nil {
		oltLogger.Errorf("Failed to send Indication_IntfOperInd for PON: %v", err)
		return
	}

	oltLogger.WithFields(log.Fields{
		"Type":      pon.Type,
		"IntfId":    pon.ID,
		"OperState": pon.OperState.Current(),
	}).Debug("Sent Indication_IntfOperInd for PON")
}

func (o *OltDevice) sendPortStatsIndication(stats *openolt.PortStatistics, portID uint32, portType string) {
	if o.InternalState.Current() == "enabled" {
		oltLogger.WithFields(log.Fields{
			"Type":   portType,
			"IntfId": portID,
		}).Trace("Sending port stats")
		stats.IntfId = InterfaceIDToPortNo(portID, portType)
		data := &openolt.Indication_PortStats{
			PortStats: stats,
		}
		stream := o.OpenoltStream
		if err := stream.Send(&openolt.Indication{Data: data}); err != nil {
			oltLogger.Errorf("Failed to send PortStats: %v", err)
			return
		}
	}
}

// processOltMessages handles messages received over the OpenOLT interface
func (o *OltDevice) processOltMessages(ctx context.Context, stream openolt.Openolt_EnableIndicationServer, wg *sync.WaitGroup) {
	oltLogger.Debug("Starting OLT Indication Channel")
	ch := o.channel

loop:
	for {
		select {
		case <-ctx.Done():
			oltLogger.Debug("OLT Indication processing canceled via context")
			break loop
		case message, ok := <-ch:
			if !ok || ctx.Err() != nil {
				oltLogger.Debug("OLT Indication processing canceled via closed channel")
				break loop
			}

			oltLogger.WithFields(log.Fields{
				"oltId":       o.ID,
				"messageType": message.Type,
			}).Trace("Received message")

			switch message.Type {
			case OltIndication:
				msg, _ := message.Data.(OltIndicationMessage)
				if msg.OperState == UP {
					_ = o.InternalState.Event("enable")
					_ = o.OperState.Event("enable")
				} else if msg.OperState == DOWN {
					_ = o.InternalState.Event("disable")
					_ = o.OperState.Event("disable")
				}
				o.sendOltIndication(msg, stream)
			case AlarmIndication:
				alarmInd, _ := message.Data.(*openolt.AlarmIndication)
				o.sendAlarmIndication(alarmInd, stream)
			case NniIndication:
				msg, _ := message.Data.(NniIndicationMessage)
				o.sendNniIndication(msg, stream)
			case PonIndication:
				msg, _ := message.Data.(PonIndicationMessage)
				pon, _ := o.GetPonById(msg.PonPortID)
				if msg.OperState == UP {
					if err := pon.OperState.Event("enable"); err != nil {
						oltLogger.WithFields(log.Fields{
							"IntfId": msg.PonPortID,
							"Err":    err,
						}).Error("Can't Enable Oper state for PON Port")
					}
					if err := pon.InternalState.Event("enable"); err != nil {
						oltLogger.WithFields(log.Fields{
							"IntfId": msg.PonPortID,
							"Err":    err,
						}).Error("Can't Enable Internal state for PON Port")
					}
				} else if msg.OperState == DOWN {
					if err := pon.OperState.Event("disable"); err != nil {
						oltLogger.WithFields(log.Fields{
							"IntfId": msg.PonPortID,
							"Err":    err,
						}).Error("Can't Disable Oper state for PON Port")
					}
					if err := pon.InternalState.Event("disable"); err != nil {
						oltLogger.WithFields(log.Fields{
							"IntfId": msg.PonPortID,
							"Err":    err,
						}).Error("Can't Disable Internal state for PON Port")
					}
				}
			default:
				oltLogger.Warnf("Received unknown message data %v for type %v in OLT Channel", message.Data, message.Type)
			}
		}
	}
	wg.Done()
	oltLogger.Warn("Stopped handling OLT Indication Channel")
}

// returns an ONU with a given Serial Number
func (o *OltDevice) FindOnuBySn(serialNumber string) (*Onu, error) {
	// TODO this function can be a performance bottleneck when we have many ONUs,
	// memoizing it will remove the bottleneck
	for _, pon := range o.Pons {
		for _, onu := range pon.Onus {
			if onu.Sn() == serialNumber {
				return onu, nil
			}
		}
	}

	return &Onu{}, fmt.Errorf("cannot-find-onu-by-serial-number-%s", serialNumber)
}

// returns an ONU with a given interface/Onu Id
func (o *OltDevice) FindOnuById(intfId uint32, onuId uint32) (*Onu, error) {
	// TODO this function can be a performance bottleneck when we have many ONUs,
	// memoizing it will remove the bottleneck
	for _, pon := range o.Pons {
		if pon.ID == intfId {
			for _, onu := range pon.Onus {
				if onu.ID == onuId {
					return onu, nil
				}
			}
		}
	}
	return &Onu{}, fmt.Errorf("cannot-find-onu-by-id-%v-%v", intfId, onuId)
}

// returns a Service with a given Mac Address
func (o *OltDevice) FindServiceByMacAddress(mac net.HardwareAddr) (ServiceIf, error) {
	// TODO this function can be a performance bottleneck when we have many ONUs,
	// memoizing it will remove the bottleneck
	for _, pon := range o.Pons {
		for _, onu := range pon.Onus {
			s, err := onu.findServiceByMacAddress(mac)
			if err == nil {
				return s, nil
			}
		}
	}

	return nil, fmt.Errorf("cannot-find-service-by-mac-address-%s", mac)
}

// GRPC Endpoints

func (o *OltDevice) ActivateOnu(context context.Context, onu *openolt.Onu) (*openolt.Empty, error) {
	oltLogger.WithFields(log.Fields{
		"OnuSn": onuSnToString(onu.SerialNumber),
	}).Info("Received ActivateOnu call from VOLTHA")
	publishEvent("ONU-activate-indication-received", int32(onu.IntfId), int32(onu.OnuId), onuSnToString(onu.SerialNumber))

	pon, _ := o.GetPonById(onu.IntfId)
	_onu, _ := pon.GetOnuBySn(onu.SerialNumber)
	_onu.SetID(onu.OnuId)

	if err := _onu.OperState.Event("enable"); err != nil {
		oltLogger.WithFields(log.Fields{
			"IntfId": _onu.PonPortID,
			"OnuSn":  _onu.Sn(),
			"OnuId":  _onu.ID,
		}).Infof("Failed to transition ONU.OperState to enabled state: %s", err.Error())
	}
	if err := _onu.InternalState.Event("enable"); err != nil {
		oltLogger.WithFields(log.Fields{
			"IntfId": _onu.PonPortID,
			"OnuSn":  _onu.Sn(),
			"OnuId":  _onu.ID,
		}).Infof("Failed to transition ONU to enabled state: %s", err.Error())
	}

	// NOTE we need to immediately activate the ONU or the OMCI state machine won't start

	return new(openolt.Empty), nil
}

func (o *OltDevice) DeactivateOnu(context.Context, *openolt.Onu) (*openolt.Empty, error) {
	oltLogger.Error("DeactivateOnu not implemented")
	return new(openolt.Empty), nil
}

func (o *OltDevice) DeleteOnu(_ context.Context, onu *openolt.Onu) (*openolt.Empty, error) {
	oltLogger.WithFields(log.Fields{
		"IntfId": onu.IntfId,
		"OnuId":  onu.OnuId,
	}).Info("Received DeleteOnu call from VOLTHA")

	pon, err := o.GetPonById(onu.IntfId)
	if err != nil {
		oltLogger.WithFields(log.Fields{
			"OnuId":  onu.OnuId,
			"IntfId": onu.IntfId,
			"err":    err,
		}).Error("Can't find PonPort")
	}
	_onu, err := pon.GetOnuById(onu.OnuId)
	if err != nil {
		oltLogger.WithFields(log.Fields{
			"OnuId":  onu.OnuId,
			"IntfId": onu.IntfId,
			"err":    err,
		}).Error("Can't find Onu")
	}

	if err := _onu.InternalState.Event("disable"); err != nil {
		oltLogger.WithFields(log.Fields{
			"IntfId": _onu.PonPortID,
			"OnuSn":  _onu.Sn(),
			"OnuId":  _onu.ID,
		}).Infof("Failed to transition ONU to disabled state: %s", err.Error())
	}

	// ONU Re-Discovery
	if o.InternalState.Current() == "enabled" && pon.InternalState.Current() == "enabled" {
		go _onu.ReDiscoverOnu()
	}

	return new(openolt.Empty), nil
}

func (o *OltDevice) DisableOlt(context.Context, *openolt.Empty) (*openolt.Empty, error) {
	// NOTE when we disable the OLT should we disable NNI, PONs and ONUs altogether?
	oltLogger.WithFields(log.Fields{
		"oltId": o.ID,
	}).Info("Disabling OLT")
	publishEvent("OLT-disable-received", -1, -1, "")

	for _, pon := range o.Pons {
		if pon.InternalState.Current() == "enabled" {
			// disable PONs
			msg := Message{
				Type: PonIndication,
				Data: PonIndicationMessage{
					OperState: DOWN,
					PonPortID: pon.ID,
				},
			}
			o.channel <- msg
		}
	}

	// Note that we are not disabling the NNI as the real OLT does not.
	// The reason for that is in-band management

	// disable OLT
	oltMsg := Message{
		Type: OltIndication,
		Data: OltIndicationMessage{
			OperState: DOWN,
		},
	}
	o.channel <- oltMsg
	return new(openolt.Empty), nil
}

func (o *OltDevice) DisablePonIf(_ context.Context, intf *openolt.Interface) (*openolt.Empty, error) {
	oltLogger.Infof("DisablePonIf request received for PON %d", intf.IntfId)
	ponID := intf.GetIntfId()
	pon, _ := o.GetPonById(intf.IntfId)

	msg := Message{
		Type: PonIndication,
		Data: PonIndicationMessage{
			OperState: DOWN,
			PonPortID: ponID,
		},
	}
	o.channel <- msg

	for _, onu := range pon.Onus {

		onuIndication := OnuIndicationMessage{
			OperState: DOWN,
			PonPortID: ponID,
			OnuID:     onu.ID,
			OnuSN:     onu.SerialNumber,
		}
		onu.sendOnuIndication(onuIndication, o.OpenoltStream)

	}

	return new(openolt.Empty), nil
}

func (o *OltDevice) EnableIndication(_ *openolt.Empty, stream openolt.Openolt_EnableIndicationServer) error {
	oltLogger.WithField("oltId", o.ID).Info("OLT receives EnableIndication call from VOLTHA")
	publishEvent("OLT-enable-received", -1, -1, "")
	o.Enable(stream)
	return nil
}

func (o *OltDevice) EnablePonIf(_ context.Context, intf *openolt.Interface) (*openolt.Empty, error) {
	oltLogger.Infof("EnablePonIf request received for PON %d", intf.IntfId)
	ponID := intf.GetIntfId()
	pon, _ := o.GetPonById(intf.IntfId)

	msg := Message{
		Type: PonIndication,
		Data: PonIndicationMessage{
			OperState: UP,
			PonPortID: ponID,
		},
	}
	o.channel <- msg

	for _, onu := range pon.Onus {

		onuIndication := OnuIndicationMessage{
			OperState: UP,
			PonPortID: ponID,
			OnuID:     onu.ID,
			OnuSN:     onu.SerialNumber,
		}
		onu.sendOnuIndication(onuIndication, o.OpenoltStream)

	}

	return new(openolt.Empty), nil
}

func (o *OltDevice) FlowAdd(ctx context.Context, flow *openolt.Flow) (*openolt.Empty, error) {
	oltLogger.WithFields(log.Fields{
		"IntfId":    flow.AccessIntfId,
		"OnuId":     flow.OnuId,
		"EthType":   fmt.Sprintf("%x", flow.Classifier.EthType),
		"InnerVlan": flow.Classifier.IVid,
		"OuterVlan": flow.Classifier.OVid,
		"FlowType":  flow.FlowType,
		"FlowId":    flow.FlowId,
		"UniID":     flow.UniId,
		"PortNo":    flow.PortNo,
	}).Tracef("OLT receives FlowAdd")

	flowKey := FlowKey{}
	if !o.enablePerf {
		flowKey = FlowKey{ID: flow.FlowId, Direction: flow.FlowType}
		olt.Flows[flowKey] = *flow
	}

	if flow.AccessIntfId == -1 {
		oltLogger.WithFields(log.Fields{
			"FlowId": flow.FlowId,
		}).Debug("Adding OLT flow")
	} else if flow.FlowType == "multicast" {
		oltLogger.WithFields(log.Fields{
			"Cookie":           flow.Cookie,
			"DstPort":          flow.Classifier.DstPort,
			"EthType":          fmt.Sprintf("%x", flow.Classifier.EthType),
			"FlowId":           flow.FlowId,
			"FlowType":         flow.FlowType,
			"GemportId":        flow.GemportId,
			"InnerVlan":        flow.Classifier.IVid,
			"IntfId":           flow.AccessIntfId,
			"IpProto":          flow.Classifier.IpProto,
			"OnuId":            flow.OnuId,
			"OuterVlan":        flow.Classifier.OVid,
			"PortNo":           flow.PortNo,
			"SrcPort":          flow.Classifier.SrcPort,
			"UniID":            flow.UniId,
			"ClassifierOPbits": flow.Classifier.OPbits,
		}).Debug("Adding OLT multicast flow")
	} else {
		pon, err := o.GetPonById(uint32(flow.AccessIntfId))
		if err != nil {
			oltLogger.WithFields(log.Fields{
				"OnuId":  flow.OnuId,
				"IntfId": flow.AccessIntfId,
				"err":    err,
			}).Error("Can't find PonPort")
		}
		onu, err := pon.GetOnuById(uint32(flow.OnuId))
		if err != nil {
			oltLogger.WithFields(log.Fields{
				"OnuId":  flow.OnuId,
				"IntfId": flow.AccessIntfId,
				"err":    err,
			}).Error("Can't find Onu")
			return nil, err
		}
		if !o.enablePerf {
			onu.Flows = append(onu.Flows, flowKey)
			// Generate event on first flow for ONU
			if len(onu.Flows) == 1 {
				publishEvent("Flow-add-received", int32(onu.PonPortID), int32(onu.ID), onuSnToString(onu.SerialNumber))
			}
		}

		msg := Message{
			Type: FlowAdd,
			Data: OnuFlowUpdateMessage{
				PonPortID: pon.ID,
				OnuID:     onu.ID,
				Flow:      flow,
			},
		}
		onu.Channel <- msg
	}

	return new(openolt.Empty), nil
}

// FlowRemove request from VOLTHA
func (o *OltDevice) FlowRemove(_ context.Context, flow *openolt.Flow) (*openolt.Empty, error) {

	oltLogger.WithFields(log.Fields{
		"FlowId":   flow.FlowId,
		"FlowType": flow.FlowType,
	}).Debug("OLT receives FlowRemove")

	if !o.enablePerf { // remove only if flow were stored
		flowKey := FlowKey{
			ID:        flow.FlowId,
			Direction: flow.FlowType,
		}

		// Check if flow exists
		storedFlow, ok := o.Flows[flowKey]
		if !ok {
			oltLogger.Errorf("Flow %v not found", flow)
			return new(openolt.Empty), status.Errorf(codes.NotFound, "Flow not found")
		}

		// if its ONU flow remove it from ONU also
		if storedFlow.AccessIntfId != -1 {
			pon := o.Pons[uint32(storedFlow.AccessIntfId)]
			onu, err := pon.GetOnuById(uint32(storedFlow.OnuId))
			if err != nil {
				oltLogger.WithFields(log.Fields{
					"OnuId":  storedFlow.OnuId,
					"IntfId": storedFlow.AccessIntfId,
					"err":    err,
				}).Error("ONU not found")
				return new(openolt.Empty), nil
			}
			onu.DeleteFlow(flowKey)
			publishEvent("Flow-remove-received", int32(onu.PonPortID), int32(onu.ID), onuSnToString(onu.SerialNumber))
		}

		// delete from olt flows
		delete(o.Flows, flowKey)
	}

	if flow.AccessIntfId == -1 {
		oltLogger.WithFields(log.Fields{
			"FlowId": flow.FlowId,
		}).Debug("Removing OLT flow")
	} else if flow.FlowType == "multicast" {
		oltLogger.WithFields(log.Fields{
			"FlowId": flow.FlowId,
		}).Debug("Removing OLT multicast flow")
	} else {

		onu, err := o.GetOnuByFlowId(flow.FlowId)
		if err != nil {
			oltLogger.WithFields(log.Fields{
				"OnuId":  flow.OnuId,
				"IntfId": flow.AccessIntfId,
				"err":    err,
			}).Error("Can't find Onu")
			return nil, err
		}

		msg := Message{
			Type: FlowRemoved,
			Data: OnuFlowUpdateMessage{
				Flow: flow,
			},
		}
		onu.Channel <- msg
	}

	return new(openolt.Empty), nil
}

func (o *OltDevice) HeartbeatCheck(context.Context, *openolt.Empty) (*openolt.Heartbeat, error) {
	res := openolt.Heartbeat{HeartbeatSignature: uint32(time.Now().Unix())}
	oltLogger.WithFields(log.Fields{
		"signature": res.HeartbeatSignature,
	}).Trace("HeartbeatCheck")
	return &res, nil
}

func (o *OltDevice) GetOnuByFlowId(flowId uint64) (*Onu, error) {
	for _, pon := range o.Pons {
		for _, onu := range pon.Onus {
			for _, fId := range onu.FlowIds {
				if fId == flowId {
					return onu, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("Cannot find Onu by flowId %d", flowId)
}

func (o *OltDevice) GetDeviceInfo(context.Context, *openolt.Empty) (*openolt.DeviceInfo, error) {

	oltLogger.WithFields(log.Fields{
		"oltId":    o.ID,
		"PonPorts": o.NumPon,
	}).Info("OLT receives GetDeviceInfo call from VOLTHA")
	devinfo := new(openolt.DeviceInfo)
	devinfo.Vendor = common.Config.Olt.Vendor
	devinfo.Model = common.Config.Olt.Model
	devinfo.HardwareVersion = common.Config.Olt.HardwareVersion
	devinfo.FirmwareVersion = common.Config.Olt.FirmwareVersion
	devinfo.Technology = common.Config.Olt.Technology
	devinfo.PonPorts = uint32(o.NumPon)
	devinfo.OnuIdStart = 1
	devinfo.OnuIdEnd = 255
	devinfo.AllocIdStart = 1024
	devinfo.AllocIdEnd = 16383
	devinfo.GemportIdStart = 1024
	devinfo.GemportIdEnd = 65535
	devinfo.FlowIdStart = 1
	devinfo.FlowIdEnd = 16383
	devinfo.DeviceSerialNumber = o.SerialNumber
	devinfo.DeviceId = common.Config.Olt.DeviceId

	return devinfo, nil
}

func (o *OltDevice) OmciMsgOut(ctx context.Context, omci_msg *openolt.OmciMsg) (*openolt.Empty, error) {
	pon, err := o.GetPonById(omci_msg.IntfId)
	if err != nil {
		oltLogger.WithFields(log.Fields{
			"error":  err,
			"onu_id": omci_msg.OnuId,
			"pon_id": omci_msg.IntfId,
		}).Error("pon ID not found")
		return nil, err
	}

	onu, err := pon.GetOnuById(omci_msg.OnuId)
	if err != nil {
		oltLogger.WithFields(log.Fields{
			"error":  err,
			"onu_id": omci_msg.OnuId,
			"pon_id": omci_msg.IntfId,
		}).Error("onu ID not found")
		return nil, err
	}

	oltLogger.WithFields(log.Fields{
		"IntfId": onu.PonPortID,
		"OnuId":  onu.ID,
		"OnuSn":  onu.Sn(),
	}).Tracef("Received OmciMsgOut")
	msg := Message{
		Type: OMCI,
		Data: OmciMessage{
			OnuSN:   onu.SerialNumber,
			OnuID:   onu.ID,
			omciMsg: omci_msg,
		},
	}
	onu.Channel <- msg
	return new(openolt.Empty), nil
}

func (o *OltDevice) OnuPacketOut(ctx context.Context, onuPkt *openolt.OnuPacket) (*openolt.Empty, error) {
	pon, err := o.GetPonById(onuPkt.IntfId)
	if err != nil {
		oltLogger.WithFields(log.Fields{
			"OnuId":  onuPkt.OnuId,
			"IntfId": onuPkt.IntfId,
			"err":    err,
		}).Error("Can't find PonPort")
	}
	onu, err := pon.GetOnuById(onuPkt.OnuId)
	if err != nil {
		oltLogger.WithFields(log.Fields{
			"OnuId":  onuPkt.OnuId,
			"IntfId": onuPkt.IntfId,
			"err":    err,
		}).Error("Can't find Onu")
	}

	oltLogger.WithFields(log.Fields{
		"IntfId": onu.PonPortID,
		"OnuId":  onu.ID,
		"OnuSn":  onu.Sn(),
		"Packet": hex.EncodeToString(onuPkt.Pkt),
	}).Trace("Received OnuPacketOut")

	rawpkt := gopacket.NewPacket(onuPkt.Pkt, layers.LayerTypeEthernet, gopacket.Default)

	pktType, err := packetHandlers.GetPktType(rawpkt)
	if err != nil {
		onuLogger.WithFields(log.Fields{
			"IntfId": onu.PonPortID,
			"OnuId":  onu.ID,
			"OnuSn":  onu.Sn(),
			"Pkt":    hex.EncodeToString(rawpkt.Data()),
		}).Error("Can't find pktType in packet, droppint it")
		return new(openolt.Empty), nil
	}

	pktMac, err := packetHandlers.GetDstMacAddressFromPacket(rawpkt)
	if err != nil {
		onuLogger.WithFields(log.Fields{
			"IntfId": onu.PonPortID,
			"OnuId":  onu.ID,
			"OnuSn":  onu.Sn(),
			"Pkt":    rawpkt.Data(),
		}).Error("Can't find Dst MacAddress in packet, droppint it")
		return new(openolt.Empty), nil
	}

	msg := Message{
		Type: OnuPacketOut,
		Data: OnuPacketMessage{
			IntfId:     onuPkt.IntfId,
			OnuId:      onuPkt.OnuId,
			Packet:     rawpkt,
			Type:       pktType,
			MacAddress: pktMac,
		},
	}

	onu.Channel <- msg

	return new(openolt.Empty), nil
}

func (o *OltDevice) Reboot(context.Context, *openolt.Empty) (*openolt.Empty, error) {

	// OLT Reboot is called in two cases:
	// - when an OLT is being removed (voltctl device disable -> voltctl device delete are called, then a new voltctl device create -> voltctl device enable will be issued)
	// - when an OLT needs to be rebooted (voltcl device reboot)

	oltLogger.WithFields(log.Fields{
		"oltId": o.ID,
	}).Info("Shutting down")
	publishEvent("OLT-reboot-received", -1, -1, "")
	go func() { _ = o.RestartOLT() }()
	return new(openolt.Empty), nil
}

func (o *OltDevice) ReenableOlt(context.Context, *openolt.Empty) (*openolt.Empty, error) {
	oltLogger.WithFields(log.Fields{
		"oltId": o.ID,
	}).Info("Received ReenableOlt request from VOLTHA")
	publishEvent("OLT-reenable-received", -1, -1, "")

	// enable OLT
	oltMsg := Message{
		Type: OltIndication,
		Data: OltIndicationMessage{
			OperState: UP,
		},
	}
	o.channel <- oltMsg

	for _, pon := range o.Pons {
		if pon.InternalState.Current() == "disabled" {
			msg := Message{
				Type: PonIndication,
				Data: PonIndicationMessage{
					OperState: UP,
					PonPortID: pon.ID,
				},
			}
			o.channel <- msg
		}
	}

	return new(openolt.Empty), nil
}

func (o *OltDevice) UplinkPacketOut(context context.Context, packet *openolt.UplinkPacket) (*openolt.Empty, error) {
	pkt := gopacket.NewPacket(packet.Pkt, layers.LayerTypeEthernet, gopacket.Default)

	err := o.Nnis[0].handleNniPacket(pkt) // FIXME we are assuming we have only one NNI

	if err != nil {
		return nil, err
	}
	return new(openolt.Empty), nil
}

func (o *OltDevice) CollectStatistics(context.Context, *openolt.Empty) (*openolt.Empty, error) {
	oltLogger.Error("CollectStatistics not implemented")
	return new(openolt.Empty), nil
}

func (o *OltDevice) GetOnuInfo(context context.Context, packet *openolt.Onu) (*openolt.OnuIndication, error) {
	oltLogger.Error("GetOnuInfo not implemented")
	return new(openolt.OnuIndication), nil
}

func (o *OltDevice) GetPonIf(context context.Context, packet *openolt.Interface) (*openolt.IntfIndication, error) {
	oltLogger.Error("GetPonIf not implemented")
	return new(openolt.IntfIndication), nil
}

func (s *OltDevice) CreateTrafficQueues(context.Context, *tech_profile.TrafficQueues) (*openolt.Empty, error) {
	oltLogger.Info("received CreateTrafficQueues")
	return new(openolt.Empty), nil
}

func (s *OltDevice) RemoveTrafficQueues(context.Context, *tech_profile.TrafficQueues) (*openolt.Empty, error) {
	oltLogger.Info("received RemoveTrafficQueues")
	return new(openolt.Empty), nil
}

func (s *OltDevice) CreateTrafficSchedulers(context context.Context, trafficSchedulers *tech_profile.TrafficSchedulers) (*openolt.Empty, error) {
	oltLogger.WithFields(log.Fields{
		"OnuId":     trafficSchedulers.OnuId,
		"IntfId":    trafficSchedulers.IntfId,
		"OnuPortNo": trafficSchedulers.PortNo,
	}).Info("received CreateTrafficSchedulers")

	if !s.enablePerf {
		pon, err := s.GetPonById(trafficSchedulers.IntfId)
		if err != nil {
			oltLogger.Errorf("Error retrieving PON by IntfId: %v", err)
			return new(openolt.Empty), err
		}
		onu, err := pon.GetOnuById(trafficSchedulers.OnuId)
		if err != nil {
			oltLogger.Errorf("Error retrieving ONU from pon by OnuId: %v", err)
			return new(openolt.Empty), err
		}
		onu.TrafficSchedulers = trafficSchedulers
	}
	return new(openolt.Empty), nil
}

func (s *OltDevice) RemoveTrafficSchedulers(context context.Context, trafficSchedulers *tech_profile.TrafficSchedulers) (*openolt.Empty, error) {
	oltLogger.WithFields(log.Fields{
		"OnuId":     trafficSchedulers.OnuId,
		"IntfId":    trafficSchedulers.IntfId,
		"OnuPortNo": trafficSchedulers.PortNo,
	}).Info("received RemoveTrafficSchedulers")
	if !s.enablePerf {
		pon, err := s.GetPonById(trafficSchedulers.IntfId)
		if err != nil {
			oltLogger.Errorf("Error retrieving PON by IntfId: %v", err)
			return new(openolt.Empty), err
		}
		onu, err := pon.GetOnuById(trafficSchedulers.OnuId)
		if err != nil {
			oltLogger.Errorf("Error retrieving ONU from pon by OnuId: %v", err)
			return new(openolt.Empty), err
		}

		onu.TrafficSchedulers = nil
	}
	return new(openolt.Empty), nil
}

// assumes caller has properly formulated an openolt.AlarmIndication
func (o *OltDevice) SendAlarmIndication(context context.Context, ind *openolt.AlarmIndication) error {
	msg := Message{
		Type: AlarmIndication,
		Data: ind,
	}

	o.channel <- msg
	return nil
}

func (o *OltDevice) PerformGroupOperation(ctx context.Context, group *openolt.Group) (*openolt.Empty, error) {
	oltLogger.WithFields(log.Fields{
		"GroupId": group.GroupId,
		"Command": group.Command,
		"Members": group.Members,
		"Action":  group.Action,
	}).Debug("received PerformGroupOperation")
	return &openolt.Empty{}, nil
}

func (o *OltDevice) DeleteGroup(ctx context.Context, group *openolt.Group) (*openolt.Empty, error) {
	oltLogger.WithFields(log.Fields{
		"GroupId": group.GroupId,
		"Command": group.Command,
		"Members": group.Members,
		"Action":  group.Action,
	}).Debug("received PerformGroupOperation")
	return &openolt.Empty{}, nil
}

func (o *OltDevice) GetExtValue(ctx context.Context, in *openolt.ValueParam) (*common_protos.ReturnValues, error) {
	return &common_protos.ReturnValues{}, nil
}

func (o *OltDevice) OnuItuPonAlarmSet(ctx context.Context, in *config.OnuItuPonAlarm) (*openolt.Empty, error) {
	return &openolt.Empty{}, nil
}

func (o *OltDevice) GetLogicalOnuDistanceZero(ctx context.Context, in *openolt.Onu) (*openolt.OnuLogicalDistance, error) {
	return &openolt.OnuLogicalDistance{}, nil
}

func (o *OltDevice) GetLogicalOnuDistance(ctx context.Context, in *openolt.Onu) (*openolt.OnuLogicalDistance, error) {
	return &openolt.OnuLogicalDistance{}, nil
}
