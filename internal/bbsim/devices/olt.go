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
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/looplab/fsm"
	"github.com/opencord/bbsim/internal/bbsim/packetHandlers"
	bbsim "github.com/opencord/bbsim/internal/bbsim/types"
	omcisim "github.com/opencord/omci-sim"
	"github.com/opencord/voltha-protos/v2/go/openolt"
	"github.com/opencord/voltha-protos/v2/go/tech_profile"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var oltLogger = log.WithFields(log.Fields{
	"module": "OLT",
})

type OltDevice struct {
	// BBSIM Internals
	ID              int
	SerialNumber    string
	NumNni          int
	NumPon          int
	NumOnuPerPon    int
	InternalState   *fsm.FSM
	channel         chan Message
	nniPktInChannel chan *bbsim.PacketMsg // packets coming in from the NNI and going to VOLTHA

	Delay int

	Pons []*PonPort
	Nnis []*NniPort

	// OLT Attributes
	OperState *fsm.FSM
}

var olt OltDevice
var oltServer *grpc.Server

func GetOLT() *OltDevice {
	return &olt
}

func CreateOLT(oltId int, nni int, pon int, onuPerPon int, sTag int, cTagInit int, auth bool, dhcp bool, delay int, isMock bool) *OltDevice {
	oltLogger.WithFields(log.Fields{
		"ID":           oltId,
		"NumNni":       nni,
		"NumPon":       pon,
		"NumOnuPerPon": onuPerPon,
	}).Debug("CreateOLT")

	olt = OltDevice{
		ID:           oltId,
		SerialNumber: fmt.Sprintf("BBSIM_OLT_%d", oltId),
		OperState: getOperStateFSM(func(e *fsm.Event) {
			oltLogger.Debugf("Changing OLT OperState from %s to %s", e.Src, e.Dst)
		}),
		NumNni:       nni,
		NumPon:       pon,
		NumOnuPerPon: onuPerPon,
		Pons:         []*PonPort{},
		Nnis:         []*NniPort{},
		Delay:        delay,
	}

	// OLT State machine
	// NOTE do we need 2 state machines for the OLT? (InternalState and OperState)
	olt.InternalState = fsm.NewFSM(
		"created",
		fsm.Events{
			{Name: "initialize", Src: []string{"disabled", "created"}, Dst: "initialized"},
			{Name: "enable", Src: []string{"initialized", "disabled"}, Dst: "enabled"},
			{Name: "disable", Src: []string{"enabled"}, Dst: "disabled"},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) {
				oltLogger.Debugf("Changing OLT InternalState from %s to %s", e.Src, e.Dst)
			},
			"enter_disabled":    func(e *fsm.Event) { olt.disableOlt() },
			"enter_initialized": func(e *fsm.Event) { olt.InitOlt() },
		},
	)

	if isMock != true {
		// create NNI Port
		nniPort, err := CreateNNI(&olt)
		if err != nil {
			oltLogger.Fatalf("Couldn't create NNI Port: %v", err)
		}

		olt.Nnis = append(olt.Nnis, &nniPort)
	}

	// create PON ports
	availableCTag := cTagInit
	for i := 0; i < pon; i++ {
		p := PonPort{
			NumOnu: olt.NumOnuPerPon,
			ID:     uint32(i),
			Type:   "pon",
			Olt:    olt,
			Onus:   []*Onu{},
		}
		p.OperState = getOperStateFSM(func(e *fsm.Event) {
			oltLogger.WithFields(log.Fields{
				"ID": p.ID,
			}).Debugf("Changing PON Port OperState from %s to %s", e.Src, e.Dst)
		})

		// create ONU devices
		for j := 0; j < onuPerPon; j++ {
			o := CreateONU(olt, p, uint32(j+1), sTag, availableCTag, auth, dhcp)
			p.Onus = append(p.Onus, o)
			availableCTag = availableCTag + 1
		}

		olt.Pons = append(olt.Pons, &p)
	}

	if err := olt.InternalState.Event("initialize"); err != nil {
		log.Errorf("Error initializing OLT: %v", err)
		return nil
	}

	return &olt
}

func (o *OltDevice) InitOlt() error {

	if oltServer == nil {
		oltServer, _ = newOltServer()
	} else {
		oltLogger.Warn("OLT server already running.")
	}

	// create new channel for processOltMessages Go routine
	o.channel = make(chan Message)

	o.nniPktInChannel = make(chan *bbsim.PacketMsg, 1024)
	// FIXME we are assuming we have only one NNI
	if o.Nnis[0] != nil {
		ch, err := o.Nnis[0].NewVethChan()
		if err == nil {
			o.nniPktInChannel = ch
		} else {
			log.Errorf("Error getting NNI channel: %v", err)
		}
	}

	for i := range olt.Pons {
		for _, onu := range olt.Pons[i].Onus {
			if err := onu.InternalState.Event("initialize"); err != nil {
				log.Errorf("Error initializing ONU: %v", err)
				return err
			}
		}
	}

	return nil
}

// callback for disable state entry
func (o *OltDevice) disableOlt() error {

	// disable all onus
	for i := range o.Pons {
		for _, onu := range o.Pons[i].Onus {
			// NOTE order of these is important.
			onu.OperState.Event("disable")
			onu.InternalState.Event("disable")
		}
	}

	// TODO handle hard poweroff (i.e. no indications sent to Voltha) vs soft poweroff
	if err := StopOltServer(); err != nil {
		return err
	}

	// terminate the OLT's processOltMessages go routine
	close(o.channel)
	// terminate the OLT's processNniPacketIns go routine
	close(o.nniPktInChannel)
	return nil
}

func (o *OltDevice) RestartOLT() error {
	oltLogger.Infof("Simulating OLT restart... (%ds)", OltRebootDelay)

	// transition internal state to disable
	if !o.InternalState.Is("disabled") {
		if err := o.InternalState.Event("disable"); err != nil {
			log.Errorf("Error disabling OLT: %v", err)
			return err
		}
	}

	time.Sleep(OltRebootDelay * time.Second)

	if err := o.InternalState.Event("initialize"); err != nil {
		log.Errorf("Error initializing OLT: %v", err)
		return err
	}
	oltLogger.Info("OLT restart completed")
	return nil
}

// newOltServer launches a new grpc server for OpenOLT
func newOltServer() (*grpc.Server, error) {
	// TODO make configurable
	address := "0.0.0.0:50060"
	lis, err := net.Listen("tcp", address)
	if err != nil {
		oltLogger.Fatalf("OLT failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()

	o := GetOLT()
	openolt.RegisterOpenoltServer(grpcServer, o)

	reflection.Register(grpcServer)

	go grpcServer.Serve(lis)
	oltLogger.Debugf("OLT Listening on: %v", address)

	return grpcServer, nil
}

// StopOltServer stops the OpenOLT grpc server
func StopOltServer() error {
	// TODO handle poweroff vs graceful shutdown
	if oltServer != nil {
		log.Warnf("Stopping OLT gRPC server")
		oltServer.Stop()
		oltServer = nil
	}
	return nil
}

// Device Methods

// Enable implements the OpenOLT EnableIndicationServer functionality
func (o *OltDevice) Enable(stream openolt.Openolt_EnableIndicationServer) error {
	oltLogger.Debug("Enable OLT called")

	wg := sync.WaitGroup{}
	wg.Add(2)

	// create Go routine to process all OLT events
	go o.processOltMessages(stream, &wg)
	go o.processNniPacketIns(stream, &wg)

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

	go o.processOmciMessages()

	// send PON Port indications
	for i, pon := range o.Pons {
		msg := Message{
			Type: PonIndication,
			Data: PonIndicationMessage{
				OperState: UP,
				PonPortID: pon.ID,
			},
		}
		o.channel <- msg

		for _, onu := range o.Pons[i].Onus {
			go onu.ProcessOnuMessages(stream, nil)
			if err := onu.InternalState.Event("discover"); err != nil {
				log.Errorf("Error discover ONU: %v", err)
				return err
			}
		}
	}

	oltLogger.Warn("Enable OLT Done")
	wg.Wait()
	return nil
}

func (o *OltDevice) processOmciMessages() {
	ch := omcisim.GetChannel()

	oltLogger.Debug("Starting OMCI Indication Channel")

	for message := range ch {
		onuId := message.Data.OnuId
		intfId := message.Data.IntfId
		onu, err := o.FindOnuById(intfId, onuId)
		if err != nil {
			oltLogger.Errorf("Failed to find onu: %v", err)
		}
		go onu.processOmciMessage(message)
	}
}

// Helpers method

func (o OltDevice) GetPonById(id uint32) (*PonPort, error) {
	for _, pon := range o.Pons {
		if pon.ID == id {
			return pon, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Cannot find PonPort with id %d in OLT %d", id, o.ID))
}

func (o OltDevice) getNniById(id uint32) (*NniPort, error) {
	for _, nni := range o.Nnis {
		if nni.ID == id {
			return nni, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Cannot find NniPort with id %d in OLT %d", id, o.ID))
}

func (o *OltDevice) sendOltIndication(msg OltIndicationMessage, stream openolt.Openolt_EnableIndicationServer) {
	data := &openolt.Indication_OltInd{OltInd: &openolt.OltIndication{OperState: msg.OperState.String()}}
	if err := stream.Send(&openolt.Indication{Data: data}); err != nil {
		oltLogger.Errorf("Failed to send Indication_OltInd: %v", err)
	}

	oltLogger.WithFields(log.Fields{
		"OperState": msg.OperState,
	}).Debug("Sent Indication_OltInd")
}

func (o *OltDevice) sendNniIndication(msg NniIndicationMessage, stream openolt.Openolt_EnableIndicationServer) {
	nni, _ := o.getNniById(msg.NniPortID)
	nni.OperState.Event("enable")
	// NOTE Operstate may need to be an integer
	operData := &openolt.Indication_IntfOperInd{IntfOperInd: &openolt.IntfOperIndication{
		Type:      nni.Type,
		IntfId:    nni.ID,
		OperState: nni.OperState.Current(),
	}}

	if err := stream.Send(&openolt.Indication{Data: operData}); err != nil {
		oltLogger.Errorf("Failed to send Indication_IntfOperInd for NNI: %v", err)
	}

	oltLogger.WithFields(log.Fields{
		"Type":      nni.Type,
		"IntfId":    nni.ID,
		"OperState": nni.OperState.Current(),
	}).Debug("Sent Indication_IntfOperInd for NNI")
}

func (o *OltDevice) sendPonIndication(msg PonIndicationMessage, stream openolt.Openolt_EnableIndicationServer) {
	pon, _ := o.GetPonById(msg.PonPortID)
	pon.OperState.Event("enable")
	discoverData := &openolt.Indication_IntfInd{IntfInd: &openolt.IntfIndication{
		IntfId:    pon.ID,
		OperState: pon.OperState.Current(),
	}}

	if err := stream.Send(&openolt.Indication{Data: discoverData}); err != nil {
		oltLogger.Errorf("Failed to send Indication_IntfInd: %v", err)
	}

	oltLogger.WithFields(log.Fields{
		"IntfId":    pon.ID,
		"OperState": pon.OperState.Current(),
	}).Debug("Sent Indication_IntfInd")

	operData := &openolt.Indication_IntfOperInd{IntfOperInd: &openolt.IntfOperIndication{
		Type:      pon.Type,
		IntfId:    pon.ID,
		OperState: pon.OperState.Current(),
	}}

	if err := stream.Send(&openolt.Indication{Data: operData}); err != nil {
		oltLogger.Errorf("Failed to send Indication_IntfOperInd for PON: %v", err)
	}

	oltLogger.WithFields(log.Fields{
		"Type":      pon.Type,
		"IntfId":    pon.ID,
		"OperState": pon.OperState.Current(),
	}).Debug("Sent Indication_IntfOperInd for PON")
}

// processOltMessages handles messages received over the OpenOLT interface
func (o *OltDevice) processOltMessages(stream openolt.Openolt_EnableIndicationServer, wg *sync.WaitGroup) {
	oltLogger.Debug("Starting OLT Indication Channel")
	for message := range o.channel {

		oltLogger.WithFields(log.Fields{
			"oltId":       o.ID,
			"messageType": message.Type,
		}).Trace("Received message")

		switch message.Type {
		case OltIndication:
			msg, _ := message.Data.(OltIndicationMessage)
			if msg.OperState == UP {
				o.InternalState.Event("enable")
				o.OperState.Event("enable")
			} else if msg.OperState == DOWN {
				o.InternalState.Event("disable")
				o.OperState.Event("disable")
			}
			o.sendOltIndication(msg, stream)
		case NniIndication:
			msg, _ := message.Data.(NniIndicationMessage)
			o.sendNniIndication(msg, stream)
		case PonIndication:
			msg, _ := message.Data.(PonIndicationMessage)
			o.sendPonIndication(msg, stream)
		default:
			oltLogger.Warnf("Received unknown message data %v for type %v in OLT Channel", message.Data, message.Type)
		}

	}
	wg.Done()
	oltLogger.Warn("Stopped handling OLT Indication Channel")
}

// processNniPacketIns handles messages received over the NNI interface
func (o *OltDevice) processNniPacketIns(stream openolt.Openolt_EnableIndicationServer, wg *sync.WaitGroup) {
	oltLogger.WithFields(log.Fields{
		"nniChannel": o.nniPktInChannel,
	}).Debug("Started NNI Channel")
	nniId := o.Nnis[0].ID // FIXME we are assuming we have only one NNI
	for message := range o.nniPktInChannel {
		oltLogger.Tracef("Received packets on NNI Channel")

		onuMac, err := packetHandlers.GetDstMacAddressFromPacket(message.Pkt)

		if err != nil {
			log.WithFields(log.Fields{
				"IntfType": "nni",
				"IntfId":   nniId,
				"Pkt":      message.Pkt.Data(),
			}).Error("Can't find Dst MacAddress in packet")
			return
		}

		onu, err := o.FindOnuByMacAddress(onuMac)
		if err != nil {
			log.WithFields(log.Fields{
				"IntfType":   "nni",
				"IntfId":     nniId,
				"Pkt":        message.Pkt.Data(),
				"MacAddress": onuMac.String(),
			}).Error("Can't find ONU with MacAddress")
			return
		}

		doubleTaggedPkt, err := packetHandlers.PushDoubleTag(onu.STag, onu.CTag, message.Pkt)
		if err != nil {
			log.Error("Fail to add double tag to packet")
		}

		data := &openolt.Indication_PktInd{PktInd: &openolt.PacketIndication{
			IntfType: "nni",
			IntfId:   nniId,
			Pkt:      doubleTaggedPkt.Data()}}
		if err := stream.Send(&openolt.Indication{Data: data}); err != nil {
			oltLogger.WithFields(log.Fields{
				"IntfType": data.PktInd.IntfType,
				"IntfId":   nniId,
				"Pkt":      doubleTaggedPkt.Data(),
			}).Errorf("Fail to send PktInd indication: %v", err)
		}
		oltLogger.WithFields(log.Fields{
			"IntfType": data.PktInd.IntfType,
			"IntfId":   nniId,
			"Pkt":      doubleTaggedPkt.Data(),
			"OnuSn":    onu.Sn(),
		}).Tracef("Sent PktInd indication")
	}
	wg.Done()
	oltLogger.WithFields(log.Fields{
		"nniChannel": o.nniPktInChannel,
	}).Warn("Stopped handling NNI Channel")
}

// returns an ONU with a given Serial Number
func (o OltDevice) FindOnuBySn(serialNumber string) (*Onu, error) {
	// TODO this function can be a performance bottleneck when we have many ONUs,
	// memoizing it will remove the bottleneck
	for _, pon := range o.Pons {
		for _, onu := range pon.Onus {
			if onu.Sn() == serialNumber {
				return onu, nil
			}
		}
	}

	return &Onu{}, errors.New(fmt.Sprintf("cannot-find-onu-by-serial-number-%s", serialNumber))
}

// returns an ONU with a given interface/Onu Id
func (o OltDevice) FindOnuById(intfId uint32, onuId uint32) (*Onu, error) {
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
	return &Onu{}, errors.New(fmt.Sprintf("cannot-find-onu-by-id-%v-%v", intfId, onuId))
}

// returns an ONU with a given Mac Address
func (o OltDevice) FindOnuByMacAddress(mac net.HardwareAddr) (*Onu, error) {
	// TODO this function can be a performance bottleneck when we have many ONUs,
	// memoizing it will remove the bottleneck
	for _, pon := range o.Pons {
		for _, onu := range pon.Onus {
			if onu.HwAddress.String() == mac.String() {
				return onu, nil
			}
		}
	}

	return &Onu{}, errors.New(fmt.Sprintf("cannot-find-onu-by-mac-address-%s", mac))
}

// GRPC Endpoints

func (o OltDevice) ActivateOnu(context context.Context, onu *openolt.Onu) (*openolt.Empty, error) {
	oltLogger.WithFields(log.Fields{
		"OnuSn": onuSnToString(onu.SerialNumber),
	}).Info("Received ActivateOnu call from VOLTHA")

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

func (o OltDevice) DeactivateOnu(context.Context, *openolt.Onu) (*openolt.Empty, error) {
	oltLogger.Error("DeactivateOnu not implemented")
	return new(openolt.Empty), nil
}

func (o OltDevice) DeleteOnu(context.Context, *openolt.Onu) (*openolt.Empty, error) {
	oltLogger.Error("DeleteOnu not implemented")
	return new(openolt.Empty), nil
}

func (o OltDevice) DisableOlt(context.Context, *openolt.Empty) (*openolt.Empty, error) {
	// NOTE when we disable the OLT should we disable NNI, PONs and ONUs altogether?
	oltMsg := Message{
		Type: OltIndication,
		Data: OltIndicationMessage{
			OperState: DOWN,
		},
	}
	o.channel <- oltMsg
	return new(openolt.Empty), nil
}

func (o OltDevice) DisablePonIf(context.Context, *openolt.Interface) (*openolt.Empty, error) {
	oltLogger.Error("DisablePonIf not implemented")
	return new(openolt.Empty), nil
}

func (o *OltDevice) EnableIndication(_ *openolt.Empty, stream openolt.Openolt_EnableIndicationServer) error {
	oltLogger.WithField("oltId", o.ID).Info("OLT receives EnableIndication call from VOLTHA")
	o.Enable(stream)
	return nil
}

func (o OltDevice) EnablePonIf(context.Context, *openolt.Interface) (*openolt.Empty, error) {
	oltLogger.Error("EnablePonIf not implemented")
	return new(openolt.Empty), nil
}

func (o OltDevice) FlowAdd(ctx context.Context, flow *openolt.Flow) (*openolt.Empty, error) {
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
	}).Tracef("OLT receives Flow")
	// TODO optionally store flows somewhere

	if flow.AccessIntfId == -1 {
		oltLogger.WithFields(log.Fields{
			"FlowId": flow.FlowId,
		}).Debugf("This is an OLT flow")
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
		}

		msg := Message{
			Type: FlowUpdate,
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

func (o OltDevice) FlowRemove(context.Context, *openolt.Flow) (*openolt.Empty, error) {
	oltLogger.Tracef("received FlowRemove")
	// TODO store flows somewhere
	return new(openolt.Empty), nil
}

func (o OltDevice) HeartbeatCheck(context.Context, *openolt.Empty) (*openolt.Heartbeat, error) {
	oltLogger.Error("HeartbeatCheck not implemented")
	return new(openolt.Heartbeat), nil
}

func (o OltDevice) GetDeviceInfo(context.Context, *openolt.Empty) (*openolt.DeviceInfo, error) {

	oltLogger.WithFields(log.Fields{
		"oltId":    o.ID,
		"PonPorts": o.NumPon,
	}).Info("OLT receives GetDeviceInfo call from VOLTHA")
	devinfo := new(openolt.DeviceInfo)
	devinfo.Vendor = "BBSim"
	devinfo.Model = "asfvolt16"
	devinfo.HardwareVersion = "emulated"
	devinfo.FirmwareVersion = ""
	devinfo.Technology = "xgspon"
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
	devinfo.DeviceId = net.HardwareAddr{0xA, 0xA, 0xA, 0xA, 0xA, byte(o.ID)}.String()

	return devinfo, nil
}

func (o OltDevice) OmciMsgOut(ctx context.Context, omci_msg *openolt.OmciMsg) (*openolt.Empty, error) {
	pon, _ := o.GetPonById(omci_msg.IntfId)
	onu, _ := pon.GetOnuById(omci_msg.OnuId)
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

func (o OltDevice) OnuPacketOut(ctx context.Context, onuPkt *openolt.OnuPacket) (*openolt.Empty, error) {
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
	}).Tracef("Received OnuPacketOut")

	rawpkt := gopacket.NewPacket(onuPkt.Pkt, layers.LayerTypeEthernet, gopacket.Default)
	pktType, err := packetHandlers.IsEapolOrDhcp(rawpkt)

	msg := Message{
		Type: OnuPacketOut,
		Data: OnuPacketMessage{
			IntfId: onuPkt.IntfId,
			OnuId:  onuPkt.OnuId,
			Packet: rawpkt,
			Type:   pktType,
		},
	}
	onu.Channel <- msg

	return new(openolt.Empty), nil
}

func (o OltDevice) Reboot(context.Context, *openolt.Empty) (*openolt.Empty, error) {
	oltLogger.Info("Shutting down")
	o.RestartOLT()
	return new(openolt.Empty), nil
}

func (o OltDevice) ReenableOlt(context.Context, *openolt.Empty) (*openolt.Empty, error) {
	oltLogger.Error("ReenableOlt not implemented")
	return new(openolt.Empty), nil
}

func (o OltDevice) UplinkPacketOut(context context.Context, packet *openolt.UplinkPacket) (*openolt.Empty, error) {
	pkt := gopacket.NewPacket(packet.Pkt, layers.LayerTypeEthernet, gopacket.Default)

	o.Nnis[0].sendNniPacket(pkt) // FIXME we are assuming we have only one NNI
	// NOTE should we return an error if sendNniPakcet fails?
	return new(openolt.Empty), nil
}

func (o OltDevice) CollectStatistics(context.Context, *openolt.Empty) (*openolt.Empty, error) {
	oltLogger.Error("CollectStatistics not implemented")
	return new(openolt.Empty), nil
}

func (o OltDevice) GetOnuInfo(context context.Context, packet *openolt.Onu) (*openolt.OnuIndication, error) {
	oltLogger.Error("GetOnuInfo not implemented")
	return new(openolt.OnuIndication), nil
}

func (o OltDevice) GetPonIf(context context.Context, packet *openolt.Interface) (*openolt.IntfIndication, error) {
	oltLogger.Error("GetPonIf not implemented")
	return new(openolt.IntfIndication), nil
}

func (s OltDevice) CreateTrafficQueues(context.Context, *tech_profile.TrafficQueues) (*openolt.Empty, error) {
	oltLogger.Info("received CreateTrafficQueues")
	return new(openolt.Empty), nil
}

func (s OltDevice) RemoveTrafficQueues(context.Context, *tech_profile.TrafficQueues) (*openolt.Empty, error) {
	oltLogger.Info("received RemoveTrafficQueues")
	return new(openolt.Empty), nil
}

func (s OltDevice) CreateTrafficSchedulers(context.Context, *tech_profile.TrafficSchedulers) (*openolt.Empty, error) {
	oltLogger.Info("received CreateTrafficSchedulers")
	return new(openolt.Empty), nil
}

func (s OltDevice) RemoveTrafficSchedulers(context.Context, *tech_profile.TrafficSchedulers) (*openolt.Empty, error) {
	oltLogger.Info("received RemoveTrafficSchedulers")
	return new(openolt.Empty), nil
}
