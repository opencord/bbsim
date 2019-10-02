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
	"io"
	"reflect"
	"time"
)

func Enable(ip string, port string) {
	client, conn := Connect(ip, port)
	defer conn.Close()

	// TODO get device info
	deviceInfo, err := getDeviceInfo(client)

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

	// TODO listen for indications
	readIndications(client)
}

func getDeviceInfo(client pb.OpenoltClient) (*pb.DeviceInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return client.GetDeviceInfo(ctx, new(pb.Empty))
}

func readIndications(client pb.OpenoltClient) {
	indications, err := client.EnableIndication(context.Background(), new(pb.Empty))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to read indications")
		return
	}

	for {
		indication, err := indications.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Info("Failed to read from indications")
			break
		}
		handleIndication(client, indication)
	}
}

func handleIndication(client pb.OpenoltClient, indication *pb.Indication) {
	switch indication.Data.(type) {
	case *pb.Indication_OltInd:
		log.Info("Received Indication_OltInd")
	case *pb.Indication_IntfInd:
		log.Info("Received Indication_IntfInd")
	case *pb.Indication_IntfOperInd:
		log.Info("Received Indication_IntfOperInd")
	case *pb.Indication_OnuDiscInd:
		onuDiscInd := indication.GetOnuDiscInd()
		handleOnuDiscIndication(client, onuDiscInd)
	case *pb.Indication_OnuInd:
		onuInd := indication.GetOnuInd()
		handleOnuIndication(client, onuInd)
	case *pb.Indication_OmciInd:
	case *pb.Indication_PktInd:
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

func handleOnuDiscIndication(client pb.OpenoltClient, onuDiscInd *pb.OnuDiscIndication) {
	log.WithFields(log.Fields{
		"IntfId":       onuDiscInd.IntfId,
		"SerialNumber": common.OnuSnToString(onuDiscInd.SerialNumber),
	}).Info("Received Onu discovery indication")

	// TODO create an ONU (using the BBSim types)
	var pir uint32 = 1000000
	Onu := pb.Onu{
		IntfId:       onuDiscInd.IntfId,
		OnuId:        uint32(1), // FIXME OnuId is hardcoded
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

func handleOnuIndication(client pb.OpenoltClient, onuDiscInd *pb.OnuIndication) {
	log.WithFields(log.Fields{
		"IntfId":       onuDiscInd.IntfId,
		"SerialNumber": common.OnuSnToString(onuDiscInd.SerialNumber),
	}).Info("Received Onu indication")

	// TODO what's next? start OMCI state machine?
}
