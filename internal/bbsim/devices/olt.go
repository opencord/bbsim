package devices

import (
	"context"
	"errors"
	"fmt"
	"gerrit.opencord.org/bbsim/api"
	"github.com/looplab/fsm"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"sync"
)

func init() {
	//log.SetReportCaller(true)
	log.SetLevel(log.DebugLevel)
}

func CreateOLT(seq int, nni int, pon int, onuPerPon int) OltDevice {
	log.WithFields(log.Fields{
		"ID": seq,
		"NumNni":nni,
		"NumPon":pon,
		"NumOnuPerPon":onuPerPon,
	}).Debug("CreateOLT")

	olt := OltDevice{
		ID: seq,
		NumNni:nni,
		NumPon:pon,
		NumOnuPerPon:onuPerPon,
		Pons: []PonPort{},
		Nnis: []NniPort{},
		channel: make(chan interface{}, 32),
	}

	// OLT State machine
	olt.InternalState = fsm.NewFSM(
		"created",
		fsm.Events{
			{Name: "enable", Src: []string{"created"}, Dst: "enabled"},
			{Name: "disable", Src: []string{"enabled"}, Dst: "disabled"},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) {
				olt.stateChange(e)
			},
		},
	)

	// create NNI Port
	nniPort := NniPort{
		ID: uint32(0),
		OperState: DOWN,
		Type: "nni",
	}
	olt.Nnis = append(olt.Nnis, nniPort)

	// create PON ports
	for i := 0; i < pon; i++ {
		p := PonPort{
			NumOnu: olt.NumOnuPerPon,
			ID: uint32(i),
			OperState: DOWN,
			Type: "pon",
		}

		// create ONU devices
		for j := 0; j < onuPerPon; j++ {
			o := CreateONU(olt, p, uint32(j + 1))
			p.Onus = append(p.Onus, o)
		}

		olt.Pons = append(olt.Pons, p)
	}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go newOltServer(olt)
	wg.Wait()
	return olt
}

func newOltServer(o OltDevice) error {
	// TODO make configurable
	address :=  "0.0.0.0:50060"
	log.Debugf("OLT Listening on: %v", address)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	openolt.RegisterOpenoltServer(grpcServer, o)

	go grpcServer.Serve(lis)

	return nil
}

// Device Methods

func (o OltDevice) Enable (stream openolt.Openolt_EnableIndicationServer) error {

	wg := sync.WaitGroup{}
	wg.Add(1)

	// create a channel for all the OLT events
	go o.oltChannels(stream)

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
			msg := Message{
				Type:      OnuDiscIndication,
				Data: OnuDiscIndicationMessage{
					Onu:     onu,
					OperState: UP,
				},
			}
			o.channel <- msg
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

func (o OltDevice) stateChange(e *fsm.Event) {
	log.WithFields(log.Fields{
		"oltId": o.ID,
		"dstState": e.Dst,
		"srcState": e.Src,
	}).Debugf("OLT state has changed")
}

func (o OltDevice) sendOltIndication(msg OltIndicationMessage, stream openolt.Openolt_EnableIndicationServer) {
	data := &openolt.Indication_OltInd{OltInd: &openolt.OltIndication{OperState: msg.OperState.String()}}
	if err := stream.Send(&openolt.Indication{Data: data}); err != nil {
		log.Error("Failed to send Indication_OltInd: %v", err)
	}

	log.WithFields(log.Fields{
		"OperState": msg.OperState,
	}).Debug("Sent Indication_OltInd")
}

func (o OltDevice) sendNniIndication(msg NniIndicationMessage, stream openolt.Openolt_EnableIndicationServer) {
	nni, _ := o.getNniById(msg.NniPortID)
	nni.OperState = UP
	operData := &openolt.Indication_IntfOperInd{IntfOperInd: &openolt.IntfOperIndication{
		Type: nni.Type,
		IntfId: nni.ID,
		OperState: nni.OperState.String(),
	}}

	if err := stream.Send(&openolt.Indication{Data: operData}); err != nil {
		log.Error("Failed to send Indication_IntfOperInd for NNI: %v", err)
	}

	log.WithFields(log.Fields{
		"Type": nni.Type,
		"IntfId": nni.ID,
		"OperState": nni.OperState.String(),
	}).Debug("Sent Indication_IntfOperInd for NNI")
}

func (o OltDevice) sendPonIndication(msg PonIndicationMessage, stream openolt.Openolt_EnableIndicationServer) {
	pon, _ := o.getPonById(msg.PonPortID)
	pon.OperState = UP
	discoverData := &openolt.Indication_IntfInd{IntfInd: &openolt.IntfIndication{
		IntfId: pon.ID,
		OperState: pon.OperState.String(),
	}}

	if err := stream.Send(&openolt.Indication{Data: discoverData}); err != nil {
		log.Error("Failed to send Indication_IntfInd: %v", err)
	}

	log.WithFields(log.Fields{
		"IntfId": pon.ID,
		"OperState": pon.OperState.String(),
	}).Debug("Sent Indication_IntfInd")

	operData := &openolt.Indication_IntfOperInd{IntfOperInd: &openolt.IntfOperIndication{
		Type: pon.Type,
		IntfId: pon.ID,
		OperState: pon.OperState.String(),
	}}

	if err := stream.Send(&openolt.Indication{Data: operData}); err != nil {
		log.Error("Failed to send Indication_IntfOperInd for PON: %v", err)
	}

	log.WithFields(log.Fields{
		"Type": pon.Type,
		"IntfId": pon.ID,
		"OperState": pon.OperState.String(),
	}).Debug("Sent Indication_IntfOperInd for PON")
}

func (o OltDevice) oltChannels(stream openolt.Openolt_EnableIndicationServer) {

	for message := range o.channel {

		_msg, _ok := message.(Message)
		if _ok {
			log.WithFields(log.Fields{
				"oltId": o.ID,
				"messageType": _msg.Type,
			}).Debug("Received message")

			switch _msg.Data.(type) {
			case OltIndicationMessage:
				msg, _ := _msg.Data.(OltIndicationMessage)
				o.InternalState.Event("enable")
				o.sendOltIndication(msg, stream)
			case NniIndicationMessage:
				msg, _ := _msg.Data.(NniIndicationMessage)
				o.sendNniIndication(msg, stream)
			case PonIndicationMessage:
				msg, _ := _msg.Data.(PonIndicationMessage)
				o.sendPonIndication(msg, stream)
			case OnuDiscIndicationMessage:
				msg, _ := _msg.Data.(OnuDiscIndicationMessage)
				msg.Onu.InternalState.Event("discover")
				msg.Onu.sendOnuDiscIndication(msg, stream)
			case OnuIndicationMessage:
				msg, _ := _msg.Data.(OnuIndicationMessage)
				pon, _ := o.getPonById(msg.PonPortID)
				onu, _ := pon.getOnuBySn(msg.OnuSN)
				onu.InternalState.Event("enable")
				onu.sendOnuIndication(msg, stream)
			default:
				log.Warnf("Received unkown message data %v for type %v", _msg.Data, _msg.Type)
			}
		} else {
			log.Warnf("Received unkown message %v", message)
		}

	}
}

// GRPC Endpoints

func (o OltDevice) ActivateOnu(context context.Context, onu *openolt.Onu) (*openolt.Empty, error)  {
	log.WithFields(log.Fields{
		"onuSerialNumber": onu.SerialNumber,
	}).Info("Received ActivateOnu call from VOLTHA")
	msg := Message{
		Type:      OnuIndication,
		Data:      OnuIndicationMessage{
			OnuSN:     onu.SerialNumber,
			PonPortID: onu.IntfId,
			OperState: UP,
		},
	}
	o.channel <- msg
	return new(openolt.Empty) , nil
}

func (o OltDevice) DeactivateOnu(context.Context, *openolt.Onu) (*openolt.Empty, error)  {
	log.Error("DeactivateOnu not implemented")
	return new(openolt.Empty) , nil
}

func (o OltDevice) DeleteOnu(context.Context, *openolt.Onu) (*openolt.Empty, error)  {
	log.Error("DeleteOnu not implemented")
	return new(openolt.Empty) , nil
}

func (o OltDevice) DisableOlt(context.Context, *openolt.Empty) (*openolt.Empty, error)  {
	log.Error("DisableOlt not implemented")
	return new(openolt.Empty) , nil
}

func (o OltDevice) DisablePonIf(context.Context, *openolt.Interface) (*openolt.Empty, error)  {
	log.Error("DisablePonIf not implemented")
	return new(openolt.Empty) , nil
}

func (o OltDevice) EnableIndication(_ *openolt.Empty, stream openolt.Openolt_EnableIndicationServer) error  {
	log.WithField("oltId", o.ID).Info("OLT receives EnableIndication call from VOLTHA")
	o.Enable(stream)
	return nil
}

func (o OltDevice) EnablePonIf(context.Context, *openolt.Interface) (*openolt.Empty, error)  {
	log.Error("EnablePonIf not implemented")
	return new(openolt.Empty) , nil
}

func (o OltDevice) FlowAdd(context.Context, *openolt.Flow) (*openolt.Empty, error)  {
	log.Error("FlowAdd not implemented")
	return new(openolt.Empty) , nil
}

func (o OltDevice) FlowRemove(context.Context, *openolt.Flow) (*openolt.Empty, error)  {
	log.Error("FlowRemove not implemented")
	return new(openolt.Empty) , nil
}

func (o OltDevice) HeartbeatCheck(context.Context, *openolt.Empty) (*openolt.Heartbeat, error)  {
	log.Error("HeartbeatCheck not implemented")
	return new(openolt.Heartbeat) , nil
}

func (o OltDevice) GetDeviceInfo(context.Context, *openolt.Empty) (*openolt.DeviceInfo, error)  {

	log.WithField("oltId", o.ID).Info("OLT receives GetDeviceInfo call from VOLTHA")
	devinfo := new(openolt.DeviceInfo)
	devinfo.Vendor = "BBSim"
	devinfo.Model = "asfvolt16"
	devinfo.HardwareVersion = ""
	devinfo.FirmwareVersion = ""
	devinfo.Technology = "xgspon"
	devinfo.PonPorts = 1
	devinfo.OnuIdStart = 1
	devinfo.OnuIdEnd = 255
	devinfo.AllocIdStart = 1024
	devinfo.AllocIdEnd = 16383
	devinfo.GemportIdStart = 1024
	devinfo.GemportIdEnd = 65535
	devinfo.FlowIdStart = 1
	devinfo.FlowIdEnd = 16383

	return devinfo, nil
}

func (o OltDevice) OmciMsgOut(context.Context, *openolt.OmciMsg) (*openolt.Empty, error)  {
	log.Error("OmciMsgOut not implemented")
	return new(openolt.Empty) , nil
}

func (o OltDevice) OnuPacketOut(context.Context, *openolt.OnuPacket) (*openolt.Empty, error)  {
	log.Error("OnuPacketOut not implemented")
	return new(openolt.Empty) , nil
}

func (o OltDevice) Reboot(context.Context, *openolt.Empty) (*openolt.Empty, error)  {
	log.Error("Reboot not implemented")
	return new(openolt.Empty) , nil
}

func (o OltDevice) ReenableOlt(context.Context, *openolt.Empty) (*openolt.Empty, error) {
	log.Error("ReenableOlt not implemented")
	return new(openolt.Empty) , nil
}

func (o OltDevice) UplinkPacketOut(context context.Context, packet *openolt.UplinkPacket) (*openolt.Empty, error) {
	log.Error("UplinkPacketOut not implemented")
	return new(openolt.Empty) , nil
}

func (o OltDevice) CollectStatistics(context.Context, *openolt.Empty) (*openolt.Empty, error)  {
	log.Error("CollectStatistics not implemented")
	return new(openolt.Empty) , nil
}

func (o OltDevice) CreateTconts(context context.Context, packet *openolt.Tconts) (*openolt.Empty, error) {
	log.Error("CreateTconts not implemented")
	return new(openolt.Empty) , nil
}

func (o OltDevice) RemoveTconts(context context.Context, packet *openolt.Tconts) (*openolt.Empty, error) {
	log.Error("RemoveTconts not implemented")
	return new(openolt.Empty) , nil
}

func (o OltDevice) GetOnuInfo(context context.Context, packet *openolt.Onu) (*openolt.OnuIndication, error) {
	log.Error("GetOnuInfo not implemented")
	return new(openolt.OnuIndication) , nil
}

func (o OltDevice) GetPonIf(context context.Context, packet *openolt.Interface) (*openolt.IntfIndication, error) {
	log.Error("GetPonIf not implemented")
	return new(openolt.IntfIndication) , nil
}