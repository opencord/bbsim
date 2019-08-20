package devices

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/opencord/voltha-protos/go/openolt"
	"github.com/looplab/fsm"
	omci "github.com/opencord/omci-sim"
	log "github.com/sirupsen/logrus"
	"net"
)

var onuLogger = log.WithFields(log.Fields{
	"module": "ONU",
})

func CreateONU(olt OltDevice, pon PonPort, id uint32) Onu {
		o := Onu{
			ID: id,
			PonPortID: pon.ID,
			PonPort: pon,
			channel: make(chan Message),
		}
		o.SerialNumber = o.NewSN(olt.ID, pon.ID, o.ID)

		// NOTE this state machine is used to track the operational
		// state as requested by VOLTHA
		o.OperState = getOperStateFSM(func(e *fsm.Event) {
			onuLogger.WithFields(log.Fields{
				"ID": o.ID,
			}).Debugf("Changing ONU OperState from %s to %s", e.Src, e.Dst)
		})

		// NOTE this state machine is used to activate the OMCI, EAPOL and DHCP clients
		o.InternalState = fsm.NewFSM(
			"created",
			fsm.Events{
				{Name: "discover", Src: []string{"created"}, Dst: "discovered"},
				{Name: "enable", Src: []string{"discovered"}, Dst: "enabled"},
				{Name: "start_omci", Src: []string{"enabled"}, Dst: "starting_openomci"},
				{Name: "start_auth", Src: []string{"starting_openomci"}, Dst: "starting_auth"},
			},
			fsm.Callbacks{
				"enter_state": func(e *fsm.Event) {
					onuLogger.WithFields(log.Fields{
						"OnuId": o.ID,
						"IntfId": o.PonPortID,
						"OnuSn": o.SerialNumber,
					}).Debugf("Changing ONU InternalState from %s to %s", e.Src, e.Dst)
				},
				"enter_starting_auth": func(e *fsm.Event) {
					log.Infof("[TEO] Starting Auth")
					//msg := Message{
					//	Type:      StartEAPOL,
					//}
					//o.channel <- msg
				},
			},
		)
		return o
}

func (o Onu) processOnuMessages(stream openolt.Openolt_EnableIndicationServer)  {
	onuLogger.WithFields(log.Fields{
		"onuID": o.ID,
		"onuSN": o.SerialNumber,
	}).Debug("Started ONU Indication Channel")

	o.stream = stream

	for message := range o.channel {
		onuLogger.WithFields(log.Fields{
			"onuID": o.ID,
			"onuSN": o.SerialNumber,
			"messageType": message.Type,
		}).Trace("Received message")

		switch message.Type {
		case OnuDiscIndication:
			msg, _ := message.Data.(OnuDiscIndicationMessage)
			o.sendOnuDiscIndication(msg, stream)
		case OnuIndication:
			msg, _ := message.Data.(OnuIndicationMessage)
			o.sendOnuIndication(msg, stream)
		case OMCI:
			msg, _ := message.Data.(OmciMessage)
			o.InternalState.Event("start_omci")
			// TODO track when OMCI is done
			o.handleOmciMessage(msg, stream)
		case FlowUpdate:
			msg, _ := message.Data.(OnuFlowUpdateMessage)
			o.handleFlowUpdate(msg, stream)
		case StartEAPOL:
			o.startWpaSupplicant(stream)
		default:
			onuLogger.Warnf("Received unknown message data %v for type %v in OLT channel", message.Data, message.Type)
		}
	}
}

func (o Onu) NewSN(oltid int, intfid uint32, onuid uint32) *openolt.SerialNumber {

	sn := new(openolt.SerialNumber)

	sn = new(openolt.SerialNumber)
	sn.VendorId = []byte("BBSM")
	sn.VendorSpecific = []byte{0, byte(oltid % 256), byte(intfid), byte(onuid)}

	return sn
}

func (o Onu) sendOnuDiscIndication(msg OnuDiscIndicationMessage, stream openolt.Openolt_EnableIndicationServer) {
	discoverData := &openolt.Indication_OnuDiscInd{OnuDiscInd: &openolt.OnuDiscIndication{
		IntfId: msg.Onu.PonPortID,
		SerialNumber: msg.Onu.SerialNumber,
	}}
	if err := stream.Send(&openolt.Indication{Data: discoverData}); err != nil {
		log.Error("Failed to send Indication_OnuDiscInd: %v", err)
	}
	o.InternalState.Event("discover")
	onuLogger.WithFields(log.Fields{
		"IntfId": msg.Onu.PonPortID,
		"SerialNumber": msg.Onu.SerialNumber,
	}).Debug("Sent Indication_OnuDiscInd")
}

func (o Onu) sendOnuIndication(msg OnuIndicationMessage, stream openolt.Openolt_EnableIndicationServer) {
	// NOTE voltha returns an ID, but if we use that ID then it complains:
	// expected_onu_id: 1, received_onu_id: 1024, event: ONU-id-mismatch, can happen if both voltha and the olt rebooted
	// so we're using the internal ID that is 1
	// o.ID = msg.OnuID
	o.OperState.Event("enable")

	indData := &openolt.Indication_OnuInd{OnuInd: &openolt.OnuIndication{
		IntfId: o.PonPortID,
		OnuId: o.ID,
		OperState: o.OperState.Current(),
		AdminState: o.OperState.Current(),
		SerialNumber: o.SerialNumber,
	}}
	if err := stream.Send(&openolt.Indication{Data: indData}); err != nil {
		log.Error("Failed to send Indication_OnuInd: %v", err)
	}
	o.InternalState.Event("enable")
	onuLogger.WithFields(log.Fields{
		"IntfId": o.PonPortID,
		"OnuId": o.ID,
		"OperState": msg.OperState.String(),
		"AdminState": msg.OperState.String(),
		"SerialNumber": o.SerialNumber,
	}).Debug("Sent Indication_OnuInd")
}

func (o Onu) handleOmciMessage(msg OmciMessage, stream openolt.Openolt_EnableIndicationServer) {

	onuLogger.WithFields(log.Fields{
		"IntfId": o.PonPortID,
		"SerialNumber": o.SerialNumber,
		"omciPacket": msg.omciMsg.Pkt,
	}).Tracef("Received OMCI message")

	var omciInd openolt.OmciIndication
	respPkt, err := omci.OmciSim(o.PonPortID, o.ID, HexDecode(msg.omciMsg.Pkt))
	if err != nil {
		onuLogger.Errorf("Error handling OMCI message %v", msg)
	}

	omciInd.IntfId = o.PonPortID
	omciInd.OnuId = o.ID
	omciInd.Pkt = respPkt

	omci := &openolt.Indication_OmciInd{OmciInd: &omciInd}
	if err := stream.Send(&openolt.Indication{Data: omci}); err != nil {
		onuLogger.Error("send omci indication failed: %v", err)
	}
	onuLogger.WithFields(log.Fields{
		"IntfId": o.PonPortID,
		"SerialNumber": o.SerialNumber,
		"omciPacket": omciInd.Pkt,
	}).Tracef("Sent OMCI message")
}

func (o Onu) handleFlowUpdate(msg OnuFlowUpdateMessage, stream openolt.Openolt_EnableIndicationServer) {
	if msg.Flow.Classifier.EthType == uint32(layers.EthernetTypeEAPOL) && msg.Flow.Classifier.OVid == 4091 {
		o.InternalState.Event("start_auth")
	}
}

func (o Onu) startWpaSupplicant(stream openolt.Openolt_EnableIndicationServer) {
//	TODO add a configurable delay?

	onuLogger.WithFields(log.Fields{
		"OnuId": o.ID,
		"IntfId": o.PonPortID,
		"OnuSn": o.SerialNumber,
	}).Infof("EAPOL State Machine starting")

	// send the packet (hacked together)
	gemid, err := omci.GetGemPortId(o.PonPortID, o.ID)
	if err != nil {
		onuLogger.WithFields(log.Fields{
			"OnuId": o.ID,
			"IntfId": o.PonPortID,
			"OnuSn": o.SerialNumber,
		}).Errorf("Can't retrieve GemPortId: %s", err)
		return
	}

	buffer := gopacket.NewSerializeBuffer()
	options := gopacket.SerializeOptions{}

	ethernetLayer := &layers.Ethernet{
		SrcMAC:       net.HardwareAddr{0x2e, 0x60, 0x70, 0x13, 0x07, byte(o.ID)},
		DstMAC:       net.HardwareAddr{0x01, 0x80, 0xC2, 0x00, 0x00, 0x03},
		EthernetType: layers.EthernetTypeEAPOL,
	}

	gopacket.SerializeLayers(buffer, options,
		ethernetLayer,
		&layers.EAPOL{Version: 1, Type: 1, Length: 0},
	)

	msg := buffer.Bytes()

	data := &openolt.Indication_PktInd{
		PktInd: &openolt.PacketIndication{
			IntfType: "pon",
			IntfId: o.PonPortID,
			GemportId: uint32(gemid),
			Pkt: msg,
		},
	}
	// end of hacked (move in an EAPOL state machine)
	if err := stream.Send(&openolt.Indication{Data: data}); err != nil {
		onuLogger.Error("Fail to send EAPOL PktInd indication. %v", err)
	}
}

// HexDecode converts the hex encoding to binary
func HexDecode(pkt []byte) []byte {
	p := make([]byte, len(pkt)/2)
	for i, j := 0, 0; i < len(pkt); i, j = i+2, j+1 {
		// Go figure this ;)
		u := (pkt[i] & 15) + (pkt[i]>>6)*9
		l := (pkt[i+1] & 15) + (pkt[i+1]>>6)*9
		p[j] = u<<4 + l
	}
	onuLogger.Tracef("Omci decoded: %x.", p)
	return p
}