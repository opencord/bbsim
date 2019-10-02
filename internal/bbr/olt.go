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

package bbr

import (
	"context"
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/opencord/bbsim/internal/bbsim/packetHandlers"
	"github.com/opencord/bbsim/internal/bbsim/responders/dhcp"
	eapolresponder "github.com/opencord/bbsim/internal/bbsim/responders/eapol"
	"github.com/opencord/bbsim/internal/common"
	pb "github.com/opencord/voltha-protos/go/openolt"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"io"
	"reflect"
	"time"
)

type OLTMock struct {
	Onus          map[string]ONUMock
	TargetOnus    int
	CompletedOnus int // Number of ONUs that have received a DHCPAck
	conn          *grpc.ClientConn
	TargetIp      string
	TargetPort    string
	TargetApiPort string
	STag          int
	NextCtag      int // this is the next available C-Tag, used when creating ONUs
}

func (o *OLTMock) Enable() {

	startTime := time.Now()

	client, conn := Connect(o.TargetIp, o.TargetPort)
	o.conn = conn
	defer conn.Close()

	deviceInfo, err := o.getDeviceInfo(client)

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Can't read device info: %v")
	}

	log.WithFields(log.Fields{
		"Vendor":             deviceInfo.Vendor,
		"Model":              deviceInfo.Model,
		"DeviceSerialNumber": deviceInfo.DeviceSerialNumber,
		"PonPorts":           deviceInfo.PonPorts,
	}).Info("Retrieved device info")

	// TODO how do we understand when it's complete? when all the ONU channels are done
	o.readIndications(client)

	defer func() {
		endTime := time.Now()
		runTime := endTime.Sub(startTime)
		log.WithField("Duration", runTime).Info("BBR done!")
	}()
}

func (o OLTMock) getDeviceInfo(client pb.OpenoltClient) (*pb.DeviceInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return client.GetDeviceInfo(ctx, new(pb.Empty))
}

func (o *OLTMock) readIndications(client pb.OpenoltClient) {

	defer func() {
		log.Info("OLT readIndications done")
	}()

	// Tell the OLT to start sending indications
	indications, err := client.EnableIndication(context.Background(), new(pb.Empty))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to enable indication stream")
		return
	}

	// listen for indications
	for {
		indication, err := indications.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {

			// TODO if code = Canceled desc = grpc: the client connection is closing
			// it means BBR completed, it's not an error

			log.WithFields(log.Fields{
				"error": err,
			}).Error("Failed to read from indications")
			break
		}
		o.handleIndication(client, indication)
	}
}

func (o *OLTMock) handleIndication(client pb.OpenoltClient, indication *pb.Indication) {
	switch indication.Data.(type) {
	case *pb.Indication_OltInd:
		log.Info("Received Indication_OltInd")
	case *pb.Indication_IntfInd:
		log.Info("Received Indication_IntfInd")
	case *pb.Indication_IntfOperInd:
		log.Info("Received Indication_IntfOperInd")
	case *pb.Indication_OnuDiscInd:
		onuDiscInd := indication.GetOnuDiscInd()
		o.handleOnuDiscIndication(client, onuDiscInd)
	case *pb.Indication_OnuInd:
		onuInd := indication.GetOnuInd()
		o.handleOnuIndication(client, onuInd)
		go func() {

			defer func() {
				log.WithFields(log.Fields{
					"onuSn":         common.OnuSnToString(onuInd.SerialNumber),
					"CompletedOnus": o.CompletedOnus,
					"TargetOnus":    o.TargetOnus,
				}).Trace("Onu done")

			}()

			onu := o.getOnuBySn(onuInd.SerialNumber)

			for message := range onu.Done {
				if message == true {
					close(onu.OmciChannel)
					o.CompletedOnus++

					if o.CompletedOnus == o.TargetOnus {
						// NOTE once all the ONUs are completed, exit
						// closing the connection is not the most elegant way,
						// but I haven't found any other way to stop
						// the indications.Recv() infinite loop
						log.Info("Simulation Done")
						ValidateAndClose(o)
					}

					break
				}
			}

		}()
	case *pb.Indication_OmciInd:
		omciIndication := indication.GetOmciInd()

		// FIXME is there a better option that generate the serial number each time?
		sn := GetSN(0, omciIndication.IntfId, omciIndication.OnuId)
		onu := o.getOnuBySn(sn)

		log.WithFields(log.Fields{
			"IntfId": omciIndication.IntfId,
			"OnuId":  omciIndication.OnuId,
			"OnuSn":  common.OnuSnToString(sn),
			"Pkt":    omciIndication.Pkt,
		}).Trace("OLT Receveives OMCI Msg")

		onu.OmciChannel <- omciIndication

	case *pb.Indication_PktInd:
		pktIndication := indication.GetPktInd()
		o.handlePktIndication(client, pktIndication)
	case *pb.Indication_PortStats:
	case *pb.Indication_FlowStats:
	case *pb.Indication_AlarmInd:
	default:
		log.WithFields(log.Fields{
			"data": indication.Data,
			"type": reflect.TypeOf(indication.Data),
		}).Warn("Indication unsupported")
	}
}

func (o *OLTMock) handleOnuDiscIndication(client pb.OpenoltClient, onuDiscInd *pb.OnuDiscIndication) {
	log.WithFields(log.Fields{
		"IntfId":       onuDiscInd.IntfId,
		"SerialNumber": common.OnuSnToString(onuDiscInd.SerialNumber),
	}).Info("Received Onu discovery indication")

	onu := ONUMock{
		OmciChannel:  make(chan *pb.OmciIndication, 1024),
		Done:         make(chan bool, 1),
		SerialNumber: onuDiscInd.SerialNumber,
		IntfId:       onuDiscInd.IntfId,
		ID:           uint32(len(o.Onus)) + 1,
		tid:          0x1,
		hpTid:        0x8000,
		seqNumber:    0,
		CTag:         o.NextCtag,
		STag:         o.STag,
	}

	o.NextCtag++

	o.Onus[common.OnuSnToString(onuDiscInd.SerialNumber)] = onu

	var pir uint32 = 1000000
	Onu := pb.Onu{
		IntfId:       onuDiscInd.IntfId,
		OnuId:        onu.ID,
		SerialNumber: onuDiscInd.SerialNumber,
		Pir:          pir,
	}

	if _, err := client.ActivateOnu(context.Background(), &Onu); err != nil {
		log.WithFields(log.Fields{
			"IntfId":       onuDiscInd.IntfId,
			"SerialNumber": common.OnuSnToString(onuDiscInd.SerialNumber),
		}).Error("Failed to activate ONU")
	}
}

func (o *OLTMock) getOnuBySn(sn *pb.SerialNumber) *ONUMock {
	onu := o.Onus[common.OnuSnToString(sn)]
	return &onu
}

func (o *OLTMock) getOnuByTags(sTag int, cTag int) (*ONUMock, error) {
	//onu := o.Onus[common.OnuSnToString(sn)]

	for _, onu := range o.Onus {
		if onu.STag == sTag && onu.CTag == cTag {
			return &onu, nil
		}
	}

	return nil, errors.New("cant-find-onu-by-c-s-tags")
}

func (o *OLTMock) handleOnuIndication(client pb.OpenoltClient, onuDiscInd *pb.OnuIndication) {
	log.WithFields(log.Fields{
		"IntfId":       onuDiscInd.IntfId,
		"SerialNumber": common.OnuSnToString(onuDiscInd.SerialNumber),
	}).Info("Received Onu indication")

	onu := o.getOnuBySn(onuDiscInd.SerialNumber)
	onu.startOmci(client)
}

func (o *OLTMock) handlePktIndication(client pb.OpenoltClient, pktIndication *pb.PacketIndication) {
	log.WithFields(log.Fields{
		"IntfType":  pktIndication.IntfType,
		"IntfId":    pktIndication.IntfId,
		"GemportId": pktIndication.GemportId,
		"FlowId":    pktIndication.FlowId,
		"PortNo":    pktIndication.PortNo,
		"Cookie":    pktIndication.Cookie,
	}).Debugf("Received PktIndication")

	// NOTE when we push the EAPOL flow we set the PortNo = OnuId for convenience sake
	// BBsim responds setting the port number that was sent with the flow

	pkt := gopacket.NewPacket(pktIndication.Pkt, layers.LayerTypeEthernet, gopacket.Default)

	if pkt.Layer(layers.LayerTypeEAP) != nil || pkt.Layer(layers.LayerTypeEAPOL) != nil {
		o.handleEapolPacket(client, pktIndication.IntfId, pktIndication.PortNo, pkt)
	} else if packetHandlers.IsDhcpPacket(pkt) {
		o.handleDhcpPacket(client, pktIndication.IntfId, pktIndication.PortNo, pkt)
	}
}

// handleEapolPacket
// This functions will respond with the correct packets based on the received one
func (o *OLTMock) handleEapolPacket(client pb.OpenoltClient, intfId uint32, onuId uint32, receivedPkt gopacket.Packet) {

	eap, eapErr := eapolresponder.ExtractEAP(receivedPkt)

	eapol, eapolErr := eapolresponder.ExtractEAPOL(receivedPkt)

	if eapErr != nil && eapolErr != nil {
		log.Errorf("Failed to Extract EAP: %v - %v", eapErr, eapolErr)
		return
	}

	sn := GetSN(0, intfId, onuId)
	onu := o.getOnuBySn(sn)

	if eapol != nil && eapol.Type == layers.EAPOLTypeStart {
		identityRequest := eapolresponder.CreateEAPIdentityRequest(1)
		pkt := eapolresponder.CreateEAPOLPkt(identityRequest, onuId, intfId)

		if err := eapolresponder.SendEapolPktOut(client, intfId, onuId, pkt); err != nil {
			log.WithFields(log.Fields{
				"OnuId":  onuId,
				"IntfId": intfId,
				"OnuSn":  common.OnuSnToString(onu.SerialNumber),
				"error":  err,
			}).Errorf("Error while sending EAPIdentityRequest packet")
			return
		}

		log.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": intfId,
			"OnuSn":  common.OnuSnToString(onu.SerialNumber),
		}).Infof("Sent EAPIdentityRequest packet")
		return
	}

	if eap.Code == layers.EAPCodeResponse && eap.Type == layers.EAPTypeIdentity {
		senddata := eapolresponder.GetMD5Data(eap)
		senddata = append([]byte{0x10}, senddata...)
		challengeRequest := eapolresponder.CreateEAPChallengeRequest(eap.Id, senddata)
		pkt := eapolresponder.CreateEAPOLPkt(challengeRequest, onuId, intfId)

		if err := eapolresponder.SendEapolPktOut(client, intfId, onuId, pkt); err != nil {
			log.WithFields(log.Fields{
				"OnuId":  onuId,
				"IntfId": intfId,
				"OnuSn":  common.OnuSnToString(onu.SerialNumber),
				"error":  err,
			}).Errorf("Error while sending EAPChallengeRequest packet")
			return
		}
		log.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": intfId,
			"OnuSn":  common.OnuSnToString(onu.SerialNumber),
		}).Infof("Sent EAPChallengeRequest packet")
		return
	} else if eap.Code == layers.EAPCodeResponse && eap.Type == layers.EAPTypeOTP {
		eapSuccess := eapolresponder.CreateEAPSuccess(eap.Id)
		pkt := eapolresponder.CreateEAPOLPkt(eapSuccess, onuId, intfId)

		if err := eapolresponder.SendEapolPktOut(client, intfId, onuId, pkt); err != nil {
			log.WithFields(log.Fields{
				"OnuId":  onuId,
				"IntfId": intfId,
				"OnuSn":  common.OnuSnToString(onu.SerialNumber),
				"error":  err,
			}).Errorf("Error while sending EAPSuccess packet")
			return
		}

		log.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": intfId,
			"OnuSn":  common.OnuSnToString(onu.SerialNumber),
		}).Infof("Sent EAP Success packet")

		onu.sendDhcpFlow(client)
	}

}

// handleDhcpPacket
// This functions will simply sends the packets back as UplinkPacketOut as they are destined to the NNI
func (o *OLTMock) handleDhcpPacket(client pb.OpenoltClient, intfId uint32, onuId uint32, receivedPkt gopacket.Packet) {

	// check if the packet is going:
	// - outgouing: toward the DHCP
	// - incoming: toward the ONU
	isIncoming := packetHandlers.IsIncomingPacket(receivedPkt)
	log.Tracef("Is Incoming: %t", isIncoming)

	if isIncoming == true {

		// TODO: handle errors in the untagging process

		sTag, _ := packetHandlers.GetVlanTag(receivedPkt)
		singleTagPkt, _ := packetHandlers.PopSingleTag(receivedPkt)
		cTag, _ := packetHandlers.GetVlanTag(singleTagPkt)

		onu, err := o.getOnuByTags(int(sTag), int(cTag))

		if err != nil {
			log.WithFields(log.Fields{
				"sTag": sTag,
				"cTag": cTag,
				"onus": o.Onus,
			}).Fatalf("Can't find ONU from c/s tags")
		}

		onuPacket := pb.OnuPacket{
			IntfId:    onu.IntfId,
			OnuId:     onu.ID,
			PortNo:    onu.ID,
			GemportId: 1,
			Pkt:       receivedPkt.Data(),
		}

		if _, err := client.OnuPacketOut(context.Background(), &onuPacket); err != nil {
			log.WithFields(log.Fields{
				"OnuId":  onuId,
				"IntfId": intfId,
				"OnuSn":  common.OnuSnToString(onu.SerialNumber),
				"error":  err,
			}).Infof("Failed to send DHCP packet to the ONU")
		}

		log.WithFields(log.Fields{
			"OnuId":  onu.ID,
			"IntfId": onu.IntfId,
			"OnuSn":  common.OnuSnToString(onu.SerialNumber),
		}).Infof("Sent DHCP packet to the ONU")

		// TODO: handle errors (even though we'll only receive DHCP packets here)
		dhcpLayer, _ := dhcp.GetDhcpLayer(receivedPkt)
		dhcpMessageType, _ := dhcp.GetDhcpMessageType(dhcpLayer)
		if dhcpMessageType == layers.DHCPMsgTypeAck {
			log.WithFields(log.Fields{
				"OnuId":  onu.ID,
				"IntfId": onu.IntfId,
				"OnuSn":  common.OnuSnToString(onu.SerialNumber),
			}).Warn("ONU %s DONE!!!!")
			onu.Done <- true
		}

	} else {
		// doble tag the packet and send it to the NNI
		sn := GetSN(0, intfId, onuId)
		onu := o.getOnuBySn(sn)

		// FIXME create PushDoubleTagMethod
		singleTaggedPkt, err := packetHandlers.PushSingleTag(onu.CTag, receivedPkt)
		if err != nil {
			log.Error("Failt to add single tag to packet")
		}
		doubleTaggedPkt, err := packetHandlers.PushSingleTag(onu.STag, singleTaggedPkt)
		if err != nil {
			log.Error("Failt to add double tag to packet")
		}

		pkt := pb.UplinkPacket{
			IntfId: 0, // BBSim does not care about which NNI, it has only one
			Pkt:    doubleTaggedPkt.Data(),
		}
		if _, err := client.UplinkPacketOut(context.Background(), &pkt); err != nil {
			log.WithFields(log.Fields{
				"OnuId":  onuId,
				"IntfId": intfId,
				"OnuSn":  common.OnuSnToString(onu.SerialNumber),
				"error":  err,
			}).Infof("Failed to send DHCP packet out of the NNI Port")
		}
		log.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": intfId,
			"OnuSn":  common.OnuSnToString(onu.SerialNumber),
		}).Infof("Sent DHCP packet out of the NNI Port")
	}

}
