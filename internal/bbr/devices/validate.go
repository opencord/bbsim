/*
 * Copyright 2018-2023 Open Networking Foundation (ONF) and the ONF Contributors

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
	"time"

	"github.com/opencord/bbsim/api/bbsim"
	pb "github.com/opencord/bbsim/api/bbsim"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func ValidateAndClose(olt *OltMock) {

	// connect to the BBSim control APIs to check that all the ONUs are in the correct state
	client, conn := ApiConnect(olt.BBSimIp, olt.BBSimApiPort)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := pb.UNIRequest{
		OnuSerialNumber: "",
		UniID:           "",
	}

	services, err := client.GetServices(ctx, &req)

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatalf("Can't reach BBSim API")
	}

	expectedEapolState := "eap_response_success_received"
	expectedDhcpState := "dhcp_ack_received"

	res := true
	for _, service := range services.Items {
		if service.UniId != 0 {
			// BBR only interacts with the first UNI, rightfully so services for different UNIs
			// won't reach the desired state
			continue
		}
		if service.DhcpState != expectedDhcpState || service.EapolState != expectedEapolState {
			res = false
			log.WithFields(log.Fields{
				"OnuSN":              service.OnuSn,
				"ServiceName":        service.Name,
				"UniId":              service.UniId,
				"DhcpState":          service.DhcpState,
				"EapolState":         service.EapolState,
				"ExpectedDhcpState":  expectedDhcpState,
				"ExpectedEapolState": expectedEapolState,
			}).Fatal("Not matching expected state")
		}
	}

	if res {
		// NOTE that in BBR we expect to have a single service but this is not always the case
		log.WithFields(log.Fields{
			"ExpectedState": expectedDhcpState,
		}).Infof("%d ONUs matching expected state", len(services.Items)/4) // for now BBSim has 4 UNIs per ONU (each UNI has a single service in BBR)
	}

	olt.conn.Close()
}

func ApiConnect(ip string, port string) (bbsim.BBSimClient, *grpc.ClientConn) {
	server := fmt.Sprintf("%s:%s", ip, port)
	conn, err := grpc.Dial(server, grpc.WithInsecure())

	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return nil, conn
	}
	return bbsim.NewBBSimClient(conn), conn
}
