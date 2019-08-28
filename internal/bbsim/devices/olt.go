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
	bbsim "github.com/opencord/bbsim/internal/bbsim/types"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/looplab/fsm"
	"github.com/opencord/voltha-protos/go/openolt"
	"github.com/opencord/voltha-protos/go/tech_profile"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"sync"
)

var oltLogger = log.WithFields(log.Fields{
	"module": "OLT",
})

func init() {
	//log.SetReportCaller(true)
	log.SetLevel(log.DebugLevel)
}

var olt = OltDevice{}

func GetOLT() OltDevice  {
	return olt
}

func CreateOLT(seq int, nni int, pon int, onuPerPon int, oltDoneChannel *chan bool, apiDoneChannel *chan bool, group *sync.WaitGroup) OltDevice {
	oltLogger.WithFields(log.Fields{
		"ID": seq,
		"NumNni":nni,
		"NumPon":pon,
		"NumOnuPerPon":onuPerPon,
	}).Debug("CreateOLT")

	olt = OltDevice{
		ID: seq,
		OperState: getOperStateFSM(func(e *fsm.Event) {
			oltLogger.Debugf("Changing OLT OperState from %s to %s", e.Src, e.Dst)
		}),
		NumNni:nni,
		NumPon:pon,
		NumOnuPerPon:onuPerPon,
		Pons: []PonPort{},
		Nnis: []NniPort{},
		channel: make(chan Message),
		oltDoneChannel: oltDoneChannel,
		apiDoneChannel: apiDoneChannel,
	}

	// OLT State machine
	// NOTE do we need 2 state machines for the OLT? (InternalState and OperState)
	olt.InternalState = fsm.NewFSM(
		"created",
		fsm.Events{
			{Name: "enable", Src: []string{"created"}, Dst: "enabled"},
			{Name: "disable", Src: []string{"enabled"}, Dst: "disabled"},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) {
				oltLogger.Debugf("Changing OLT InternalState from %s to %s", e.Src, e.Dst)
			},
		},
	)

	// create NNI Port
	nniPort := NniPort{
		ID: uint32(0),
		OperState: getOperStateFSM(func(e *fsm.Event) {
			oltLogger.Debugf("Changing NNI OperState from %s to %s", e.Src, e.Dst)
		}),
		Type: "nni",
	}
	olt.Nnis = append(olt.Nnis, nniPort)

	// create PON ports
	//onuId := 1
	for i := 0; i < pon; i++ {
		p := PonPort{
			NumOnu: olt.NumOnuPerPon,
			ID: uint32(i),
			Type: "pon",
		}
		p.OperState = getOperStateFSM(func(e *fsm.Event) {
			oltLogger.WithFields(log.Fields{
				"ID": p.ID,
			}).Debugf("Changing PON Port OperState from %s to %s", e.Src, e.Dst)
		})

		// create ONU devices
		for j := 0; j < onuPerPon; j++ {
			//o := CreateONU(olt, p, uint32(onuId))
			o := CreateONU(olt, p, uint32(j + 1))
			p.Onus = append(p.Onus, o)
			//onuId = onuId + 1
		}

		olt.Pons = append(olt.Pons, p)
	}

	newOltServer(olt)

	group.Done()
	return olt
}

func newOltServer(o OltDevice) error {
	// TODO make configurable
	address :=  "0.0.0.0:50060"
	lis, err := net.Listen("tcp", address)
	if err != nil {
		oltLogger.Fatalf("OLT failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	openolt.RegisterOpenoltServer(grpcServer, o)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go grpcServer.Serve(lis)
	oltLogger.Debugf("OLT Listening on: %v", address)

	for {
		_, ok := <- *o.oltDoneChannel
		if !ok {
			// if the olt channel is closed, stop the gRPC server
			log.Warnf("Stopping OLT gRPC server")
			grpcServer.Stop()
			wg.Done()
			break
		}
	}

	wg.Wait()

	return nil
}

// Device Methods

func (o OltDevice) Enable (stream openolt.Openolt_EnableIndicationServer) error {

	oltLogger.Debug("Enable OLT called")

	wg := sync.WaitGroup{}
	wg.Add(1)

	// create a channel for all the OLT events
	go o.processOltMessages(stream)

	// enable the OLT
	olt_msg := Message{
		Type: OltIndication,
		Data: OltIndicationMessage{
			OperState: UP,
		},
	}
	o.channel <- olt_msg

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

		for _, onu := range pon.Onus {
			go onu.processOnuMessages(stream)
			go onu.processOmciMessages(stream)
			msg := Message{
				Type:      OnuDiscIndication,
				Data: OnuDiscIndicationMessage{
					Onu:     onu,
					OperState: UP,
				},
			}
			onu.channel <- msg
		}
	}

	wg.Wait()
	return nil
}

// Helpers method

func (o OltDevice) getPonById(id uint32) (*PonPort, error) {
	for _, pon := range o.Pons {
		if pon.ID == id {
			return &pon, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Cannot find PonPort with id %d in OLT %d", id, o.ID))
}

func (o OltDevice) getNniById(id uint32) (*NniPort, error) {
	for _, nni := range o.Nnis {
		if nni.ID == id {
			return &nni, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Cannot find NniPort with id %d in OLT %d", id, o.ID))
}

func (o OltDevice) sendOltIndication(msg OltIndicationMessage, stream openolt.Openolt_EnableIndicationServer) {
	data := &openolt.Indication_OltInd{OltInd: &openolt.OltIndication{OperState: msg.OperState.String()}}
	if err := stream.Send(&openolt.Indication{Data: data}); err != nil {
		oltLogger.Errorf("Failed to send Indication_OltInd: %v", err)
	}

	oltLogger.WithFields(log.Fields{
		"OperState": msg.OperState,
	}).Debug("Sent Indication_OltInd")
}

func (o OltDevice) sendNniIndication(msg NniIndicationMessage, stream openolt.Openolt_EnableIndicationServer) {
	nni, _ := o.getNniById(msg.NniPortID)
	nni.OperState.Event("enable")
	// NOTE Operstate may need to be an integer
	operData := &openolt.Indication_IntfOperInd{IntfOperInd: &openolt.IntfOperIndication{
		Type: nni.Type,
		IntfId: nni.ID,
		OperState: nni.OperState.Current(),
	}}

	if err := stream.Send(&openolt.Indication{Data: operData}); err != nil {
		oltLogger.Errorf("Failed to send Indication_IntfOperInd for NNI: %v", err)
	}

	oltLogger.WithFields(log.Fields{
		"Type": nni.Type,
		"IntfId": nni.ID,
		"OperState": nni.OperState.Current(),
	}).Debug("Sent Indication_IntfOperInd for NNI")
}

func (o OltDevice) sendPonIndication(msg PonIndicationMessage, stream openolt.Openolt_EnableIndicationServer) {
	pon, _ := o.getPonById(msg.PonPortID)
	pon.OperState.Event("enable")
	discoverData := &openolt.Indication_IntfInd{IntfInd: &openolt.IntfIndication{
		IntfId: pon.ID,
		OperState: pon.OperState.Current(),
	}}

	if err := stream.Send(&openolt.Indication{Data: discoverData}); err != nil {
		oltLogger.Errorf("Failed to send Indication_IntfInd: %v", err)
	}

	oltLogger.WithFields(log.Fields{
		"IntfId": pon.ID,
		"OperState": pon.OperState.Current(),
	}).Debug("Sent Indication_IntfInd")

	operData := &openolt.Indication_IntfOperInd{IntfOperInd: &openolt.IntfOperIndication{
		Type: pon.Type,
		IntfId: pon.ID,
		OperState: pon.OperState.Current(),
	}}

	if err := stream.Send(&openolt.Indication{Data: operData}); err != nil {
		oltLogger.Errorf("Failed to send Indication_IntfOperInd for PON: %v", err)
	}

	oltLogger.WithFields(log.Fields{
		"Type": pon.Type,
		"IntfId": pon.ID,
		"OperState": pon.OperState.Current(),
	}).Debug("Sent Indication_IntfOperInd for PON")
}

func (o OltDevice) processOltMessages(stream openolt.Openolt_EnableIndicationServer) {
	oltLogger.Debug("Started OLT Indication Channel")
	for message := range o.channel {


		oltLogger.WithFields(log.Fields{
			"oltId": o.ID,
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
			oltLogger.Warnf("Received unknown message data %v for type %v in OLT channel", message.Data, message.Type)
		}

	}
}

// GRPC Endpoints

func (o OltDevice) ActivateOnu(context context.Context, onu *openolt.Onu) (*openolt.Empty, error)  {
	oltLogger.WithFields(log.Fields{
		"onuSerialNumber": onu.SerialNumber,
	}).Info("Received ActivateOnu call from VOLTHA")

	pon, _ := o.getPonById(onu.IntfId)
	_onu, _ := pon.getOnuBySn(onu.SerialNumber)

	// NOTE we need to immediately activate the ONU or the OMCI state machine won't start
	msg := Message{
		Type:      OnuIndication,
		Data:      OnuIndicationMessage{
			OnuSN:     onu.SerialNumber,
			PonPortID: onu.IntfId,
			OperState: UP,
		},
	}
	_onu.channel <- msg
	return new(openolt.Empty) , nil
}

func (o OltDevice) DeactivateOnu(context.Context, *openolt.Onu) (*openolt.Empty, error)  {
	oltLogger.Error("DeactivateOnu not implemented")
	return new(openolt.Empty) , nil
}

func (o OltDevice) DeleteOnu(context.Context, *openolt.Onu) (*openolt.Empty, error)  {
	oltLogger.Error("DeleteOnu not implemented")
	return new(openolt.Empty) , nil
}

func (o OltDevice) DisableOlt(context.Context, *openolt.Empty) (*openolt.Empty, error)  {
	// NOTE when we disable the OLT should we disable NNI, PONs and ONUs altogether?
	olt_msg := Message{
		Type: OltIndication,
		Data: OltIndicationMessage{
			OperState: DOWN,
		},
	}
	o.channel <- olt_msg
	return new(openolt.Empty) , nil
}

func (o OltDevice) DisablePonIf(context.Context, *openolt.Interface) (*openolt.Empty, error)  {
	oltLogger.Error("DisablePonIf not implemented")
	return new(openolt.Empty) , nil
}

func (o OltDevice) EnableIndication(_ *openolt.Empty, stream openolt.Openolt_EnableIndicationServer) error  {
	oltLogger.WithField("oltId", o.ID).Info("OLT receives EnableIndication call from VOLTHA")
	o.Enable(stream)
	return nil
}

func (o OltDevice) EnablePonIf(context.Context, *openolt.Interface) (*openolt.Empty, error)  {
	oltLogger.Error("EnablePonIf not implemented")
	return new(openolt.Empty) , nil
}

func (o OltDevice) FlowAdd(ctx context.Context, flow *openolt.Flow) (*openolt.Empty, error)  {
	oltLogger.WithFields(log.Fields{
		"IntfId": flow.AccessIntfId,
		"OnuId": flow.OnuId,
		"EthType": fmt.Sprintf("%x", flow.Classifier.EthType),
		"InnerVlan": flow.Classifier.IVid,
		"OuterVlan": flow.Classifier.OVid,
		"FlowType": flow.FlowType,
		"FlowId": flow.FlowId,
		"UniID": flow.UniId,
		"PortNo": flow.PortNo,
	}).Infof("OLT receives Flow")
	// TODO optionally store flows somewhere

	if flow.AccessIntfId == -1 {
		oltLogger.WithFields(log.Fields{
			"FlowId": flow.FlowId,
		}).Debugf("This is an OLT flow")
	} else {
		pon, _ := o.getPonById(uint32(flow.AccessIntfId))
		onu, _ := pon.getOnuById(uint32(flow.OnuId))

		msg := Message{
			Type:      FlowUpdate,
			Data:      OnuFlowUpdateMessage{
				PonPortID:  pon.ID,
				OnuID: 	onu.ID,
				Flow: 	flow,
			},
		}
		onu.channel <- msg
	}

	return new(openolt.Empty) , nil
}

func (o OltDevice) FlowRemove(context.Context, *openolt.Flow) (*openolt.Empty, error)  {
	oltLogger.Info("received FlowRemove")
	// TODO store flows somewhere
	return new(openolt.Empty) , nil
}

func (o OltDevice) HeartbeatCheck(context.Context, *openolt.Empty) (*openolt.Heartbeat, error)  {
	oltLogger.Error("HeartbeatCheck not implemented")
	return new(openolt.Heartbeat) , nil
}

func (o OltDevice) GetDeviceInfo(context.Context, *openolt.Empty) (*openolt.DeviceInfo, error)  {

	oltLogger.WithFields(log.Fields{
		"oltId": o.ID,
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
	devinfo.DeviceSerialNumber = fmt.Sprintf("BBSIM_OLT_%d", o.ID)

	return devinfo, nil
}

func (o OltDevice) OmciMsgOut(ctx context.Context, omci_msg *openolt.OmciMsg) (*openolt.Empty, error)  {
	pon, _ := o.getPonById(omci_msg.IntfId)
	onu, _ := pon.getOnuById(omci_msg.OnuId)
	msg := Message{
		Type:      OMCI,
		Data:      OmciMessage{
			OnuSN:  onu.SerialNumber,
			OnuID: 	onu.ID,
			omciMsg: 	omci_msg,
		},
	}
	onu.channel <- msg
	return new(openolt.Empty) , nil
}

func (o OltDevice) OnuPacketOut(ctx context.Context, onuPkt *openolt.OnuPacket) (*openolt.Empty, error)  {
	pon, _ := o.getPonById(onuPkt.IntfId)
	onu, _ := pon.getOnuById(onuPkt.OnuId)

	rawpkt := gopacket.NewPacket(onuPkt.Pkt, layers.LayerTypeEthernet, gopacket.Default)

	// NOTE is this the best way to the to the ethertype?
	etherType := rawpkt.Layer(layers.LayerTypeEthernet).(*layers.Ethernet).EthernetType

	if etherType == layers.EthernetTypeEAPOL {
		eapolPkt := bbsim.ByteMsg{IntfId: onuPkt.IntfId, OnuId: onuPkt.OnuId, Bytes: rawpkt.Data()}
		onu.eapolPktOutCh <- &eapolPkt
	}
	return new(openolt.Empty) , nil
}

func (o OltDevice) Reboot(context.Context, *openolt.Empty) (*openolt.Empty, error)  {
	oltLogger.Info("Shutting Down")
	close(*o.oltDoneChannel)
	close(*o.apiDoneChannel)
	return new(openolt.Empty) , nil
}

func (o OltDevice) ReenableOlt(context.Context, *openolt.Empty) (*openolt.Empty, error) {
	oltLogger.Error("ReenableOlt not implemented")
	return new(openolt.Empty) , nil
}

func (o OltDevice) UplinkPacketOut(context context.Context, packet *openolt.UplinkPacket) (*openolt.Empty, error) {
	oltLogger.Warn("UplinkPacketOut not implemented")
	return new(openolt.Empty) , nil
}

func (o OltDevice) CollectStatistics(context.Context, *openolt.Empty) (*openolt.Empty, error)  {
	oltLogger.Error("CollectStatistics not implemented")
	return new(openolt.Empty) , nil
}

func (o OltDevice) GetOnuInfo(context context.Context, packet *openolt.Onu) (*openolt.OnuIndication, error) {
	oltLogger.Error("GetOnuInfo not implemented")
	return new(openolt.OnuIndication) , nil
}

func (o OltDevice) GetPonIf(context context.Context, packet *openolt.Interface) (*openolt.IntfIndication, error) {
	oltLogger.Error("GetPonIf not implemented")
	return new(openolt.IntfIndication) , nil
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