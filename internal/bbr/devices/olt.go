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
	"fmt"
	"github.com/opencord/bbsim/internal/bbsim/devices"
	"github.com/opencord/voltha-protos/go/openolt"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"io"
	"time"
)

type OltMock struct {
	Olt          *devices.OltDevice
	BBSimIp      string
	BBSimPort    string
	BBSimApiPort string

	conn          *grpc.ClientConn
	completedOnus int // Number of ONUs that have received a DHCPAck
}

// trigger an enable call and start the same listeners on the gRPC stream that VOLTHA would create
// this method is blocking
func (o *OltMock) Start() {
	log.Info("Starting Mock OLT")

	for _, pon := range o.Olt.Pons {
		for _, onu := range pon.Onus {
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
		log.Infof("Received Indication %v", indication)
		//o.handleIndication(client, indication)
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
