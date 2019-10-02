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
			log.WithFields(log.Fields{
				"error": err,
			}).Info("Failed to read from indications")
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
	}
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
	}).Warn("Received PktIndication")

	// NOTE when we push the EAPOL flow we set the PortNo = OnuId for convenience sake
	// BBsim responds setting the port number that was sent with the flow

	log.WithFields(log.Fields{
		"intfID": pktIndication.IntfId,
		"OnuId":  pktIndication.PortNo,
	}).Warn("Received PktIndication parsed")

	// TODO generate EAPOL response packets and send them via client.OnuPacketOut()
}

//OnuIDFromPortNum returns ONUID derived from portNumber
func OnuIDFromPortNum(portNum uint32) uint32 {
	return (portNum >> 4) & 127
}

//IntfIDFromUniPortNum returns IntfID derived from portNum
func IntfIDFromUniPortNum(portNum uint32) uint32 {
	return (portNum >> 11) & 15
}
