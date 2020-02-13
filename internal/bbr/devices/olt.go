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
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/opencord/bbsim/internal/bbsim/devices"
	"github.com/opencord/bbsim/internal/bbsim/packetHandlers"
	"github.com/opencord/bbsim/internal/common"
	"github.com/opencord/voltha-protos/v2/go/openolt"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"io"
	"reflect"
	"time"
)

type OltMock struct {
	LastUsedOnuId map[uint32]uint32
	Olt           *devices.OltDevice
	BBSimIp       string
	BBSimPort     string
	BBSimApiPort  string

	conn *grpc.ClientConn

	TargetOnus    int
	CompletedOnus int // Number of ONUs that have received a DHCPAck
}

// trigger an enable call and start the same listeners on the gRPC stream that VOLTHA would create
// this method is blocking
func (o *OltMock) Start() {
	log.Info("Starting Mock OLT")

	for _, pon := range o.Olt.Pons {
		for _, onu := range pon.Onus {
			if err := onu.InternalState.Event("initialize"); err != nil {
				log.Fatalf("Error initializing ONU: %v", err)
			}
			log.Debugf("Created ONU: %s (%d:%d)", onu.Sn(), onu.STag, onu.CTag)
		}
	}

	client, conn := Connect(o.BBSimIp, o.BBSimPort)
	o.conn = conn
	defer conn.Close()

	deviceInfo, err := o.getDeviceInfo(client)

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Can't read device info")
	}

	log.WithFields(log.Fields{
		"Vendor":             deviceInfo.Vendor,
		"Model":              deviceInfo.Model,
		"DeviceSerialNumber": deviceInfo.DeviceSerialNumber,
		"PonPorts":           deviceInfo.PonPorts,
	}).Info("Retrieved device info")

	o.readIndications(client)

}

func (o *OltMock) getDeviceInfo(client openolt.OpenoltClient) (*openolt.DeviceInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return client.GetDeviceInfo(ctx, new(openolt.Empty))
}

func (o *OltMock) getOnuByTags(sTag int, cTag int) (*devices.Onu, error) {

	for _, pon := range o.Olt.Pons {
		for _, onu := range pon.Onus {
			if onu.STag == sTag && onu.CTag == cTag {
				return onu, nil
			}
		}
	}

	return nil, errors.New("cant-find-onu-by-c-s-tags")
}

func (o *OltMock) readIndications(client openolt.OpenoltClient) {
	defer func() {
		log.Info("OLT readIndications done")
	}()

	// Tell the OLT to start sending indications
	indications, err := client.EnableIndication(context.Background(), new(openolt.Empty))
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

			// the connection is closed once we have sent the DHCP_ACK packet to all of the ONUs
			// it means BBR completed, it's not an error

			log.WithFields(log.Fields{
				"error": err,
			}).Debug("Failed to read from indications")
			break
		}

		o.handleIndication(client, indication)
	}
}

func (o *OltMock) handleIndication(client openolt.OpenoltClient, indication *openolt.Indication) {
	switch indication.Data.(type) {
	case *openolt.Indication_OltInd:
		log.Info("Received Indication_OltInd")
	case *openolt.Indication_IntfInd:
		log.Info("Received Indication_IntfInd")
	case *openolt.Indication_IntfOperInd:
		log.Info("Received Indication_IntfOperInd")
	case *openolt.Indication_OnuDiscInd:
		onuDiscInd := indication.GetOnuDiscInd()
		o.handleOnuDiscIndication(client, onuDiscInd)
	case *openolt.Indication_OnuInd:
		onuInd := indication.GetOnuInd()
		o.handleOnuIndication(client, onuInd)
	case *openolt.Indication_OmciInd:
		omciIndication := indication.GetOmciInd()
		o.handleOmciIndication(client, omciIndication)
	case *openolt.Indication_PktInd:
		pktIndication := indication.GetPktInd()
		o.handlePktIndication(client, pktIndication)
	case *openolt.Indication_PortStats:
	case *openolt.Indication_FlowStats:
	case *openolt.Indication_AlarmInd:
	default:
		log.WithFields(log.Fields{
			"data": indication.Data,
			"type": reflect.TypeOf(indication.Data),
		}).Warn("Indication unsupported")
	}
}

func (o *OltMock) handleOnuDiscIndication(client openolt.OpenoltClient, onuDiscInd *openolt.OnuDiscIndication) {
	log.WithFields(log.Fields{
		"IntfId":       onuDiscInd.IntfId,
		"SerialNumber": common.OnuSnToString(onuDiscInd.SerialNumber),
	}).Info("Received Onu discovery indication")

	onu, err := o.Olt.FindOnuBySn(common.OnuSnToString(onuDiscInd.SerialNumber))

	if err != nil {
		log.WithFields(log.Fields{
			"IntfId":       onuDiscInd.IntfId,
			"SerialNumber": common.OnuSnToString(onuDiscInd.SerialNumber),
			"Err":          err,
		}).Fatal("Cannot find ONU")
	}

	// creating and storing ONU IDs
	id := o.LastUsedOnuId[onuDiscInd.IntfId] + 1
	o.LastUsedOnuId[onuDiscInd.IntfId] = o.LastUsedOnuId[onuDiscInd.IntfId] + 1
	onu.SetID(id)

	var pir uint32 = 1000000
	Onu := openolt.Onu{
		IntfId:       onu.PonPortID,
		OnuId:        id,
		SerialNumber: onu.SerialNumber,
		Pir:          pir,
	}

	if _, err := client.ActivateOnu(context.Background(), &Onu); err != nil {
		log.WithFields(log.Fields{
			"IntfId":       onuDiscInd.IntfId,
			"SerialNumber": common.OnuSnToString(onuDiscInd.SerialNumber),
		}).Error("Failed to activate ONU")
	}
}

func (o *OltMock) handleOnuIndication(client openolt.OpenoltClient, onuInd *openolt.OnuIndication) {
	log.WithFields(log.Fields{
		"IntfId":       onuInd.IntfId,
		"SerialNumber": common.OnuSnToString(onuInd.SerialNumber),
	}).Info("Received Onu indication")

	onu, err := o.Olt.FindOnuBySn(common.OnuSnToString(onuInd.SerialNumber))

	if err != nil {
		log.WithFields(log.Fields{
			"IntfId":       onuInd.IntfId,
			"SerialNumber": common.OnuSnToString(onuInd.SerialNumber),
		}).Fatal("Cannot find ONU")
	}

	ctx, cancel := context.WithCancel(context.TODO())
	go onu.ProcessOnuMessages(ctx, nil, client)

	go func() {

		defer func() {
			log.WithFields(log.Fields{
				"onuSn":         common.OnuSnToString(onuInd.SerialNumber),
				"CompletedOnus": o.CompletedOnus,
				"TargetOnus":    o.TargetOnus,
			}).Debugf("Onu done")

			// close the ONU channel
			cancel()
		}()

		for message := range onu.DoneChannel {
			if message == true {
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

	// TODO change the state instead of calling an ONU method from here
	onu.StartOmci(client)
}

func (o *OltMock) handleOmciIndication(client openolt.OpenoltClient, omciInd *openolt.OmciIndication) {

	pon, err := o.Olt.GetPonById(omciInd.IntfId)
	if err != nil {
		log.WithFields(log.Fields{
			"OnuId":  omciInd.OnuId,
			"IntfId": omciInd.IntfId,
			"err":    err,
		}).Fatal("Can't find PonPort")
	}
	onu, _ := pon.GetOnuById(omciInd.OnuId)
	if err != nil {
		log.WithFields(log.Fields{
			"OnuId":  omciInd.OnuId,
			"IntfId": omciInd.IntfId,
			"err":    err,
		}).Fatal("Can't find Onu")
	}

	log.WithFields(log.Fields{
		"IntfId": onu.PonPortID,
		"OnuId":  onu.ID,
		"OnuSn":  onu.Sn(),
		"Pkt":    omciInd.Pkt,
	}).Trace("Received Onu omci indication")

	msg := devices.Message{
		Type: devices.OmciIndication,
		Data: devices.OmciIndicationMessage{
			OnuSN:   onu.SerialNumber,
			OnuID:   onu.ID,
			OmciInd: omciInd,
		},
	}
	onu.Channel <- msg
}

func (o *OltMock) handlePktIndication(client openolt.OpenoltClient, pktIndication *openolt.PacketIndication) {

	pkt := gopacket.NewPacket(pktIndication.Pkt, layers.LayerTypeEthernet, gopacket.Default)

	pktType, err := packetHandlers.IsEapolOrDhcp(pkt)

	if err != nil {
		log.Warnf("Ignoring packet as it's neither EAPOL or DHCP")
		return
	}

	log.WithFields(log.Fields{
		"IntfType":  pktIndication.IntfType,
		"IntfId":    pktIndication.IntfId,
		"GemportId": pktIndication.GemportId,
		"FlowId":    pktIndication.FlowId,
		"PortNo":    pktIndication.PortNo,
		"Cookie":    pktIndication.Cookie,
		"pktType":   pktType,
	}).Trace("Received PktIndication")

	msg := devices.Message{}
	if pktIndication.IntfType == "nni" {
		// This is an packet that is arriving from the NNI and needs to be sent to an ONU
		// in this case we need to fin the ONU from the C/S tags
		// TODO: handle errors in the untagging process
		sTag, _ := packetHandlers.GetVlanTag(pkt)
		singleTagPkt, _ := packetHandlers.PopSingleTag(pkt)
		cTag, _ := packetHandlers.GetVlanTag(singleTagPkt)

		onu, err := o.getOnuByTags(int(sTag), int(cTag))

		if err != nil {
			log.WithFields(log.Fields{
				"sTag": sTag,
				"cTag": cTag,
			}).Fatalf("Can't find ONU from c/s tags")
		}

		msg = devices.Message{
			Type: devices.OnuPacketIn,
			Data: devices.OnuPacketMessage{
				IntfId: pktIndication.IntfId,
				OnuId:  onu.ID,
				Packet: pkt,
				Type:   pktType,
			},
		}
		// NOTE we send it on the ONU channel so that is handled as all the others packets in a separate thread
		onu.Channel <- msg
	} else {
		// TODO a very similar construct is used in many places,
		// abstract this in an OLT method
		pon, err := o.Olt.GetPonById(pktIndication.IntfId)
		if err != nil {
			log.WithFields(log.Fields{
				"OnuId":  pktIndication.PortNo,
				"IntfId": pktIndication.IntfId,
				"err":    err,
			}).Fatal("Can't find PonPort")
		}
		onu, err := pon.GetOnuById(pktIndication.PortNo)
		if err != nil {
			log.WithFields(log.Fields{
				"OnuId":  pktIndication.PortNo,
				"IntfId": pktIndication.IntfId,
				"err":    err,
			}).Fatal("Can't find Onu")
		}
		// NOTE when we push the EAPOL flow we set the PortNo = OnuId for convenience sake
		// BBsim responds setting the port number that was sent with the flow
		msg = devices.Message{
			Type: devices.OnuPacketIn,
			Data: devices.OnuPacketMessage{
				IntfId: pktIndication.IntfId,
				OnuId:  pktIndication.PortNo,
				Packet: pkt,
				Type:   pktType,
			},
		}
		onu.Channel <- msg
	}
}

// TODO Move in a different file
func Connect(ip string, port string) (openolt.OpenoltClient, *grpc.ClientConn) {
	server := fmt.Sprintf("%s:%s", ip, port)
	conn, err := grpc.Dial(server, grpc.WithInsecure())

	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return nil, conn
	}
	return openolt.NewOpenoltClient(conn), conn
}
