// Copyright 2018-present Open Networking Foundation
// SPDX-FileCopyrightText: 2019 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package devices

import (
	"context"
	"fmt"
	"github.com/opencord/bbsim/api/bbsim"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"time"
)

func ValidateAndClose(olt *OltMock) {

	// connect to the BBSim control APIs to check that all the ONUs are in the correct state
	client, conn := ApiConnect(olt.BBSimIp, olt.BBSimApiPort)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	onus, err := client.GetONUs(ctx, &bbsim.Empty{})

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatalf("Can't reach BBSim API")
	}

	expectedState := "dhcp_ack_received"

	res := true
	for _, onu := range onus.Items {
		if onu.InternalState != expectedState {
			res = false
			log.WithFields(log.Fields{
				"OnuSN":         onu.SerialNumber,
				"OnuId":         onu.ID,
				"InternalState": onu.InternalState,
				"ExpectedSatte": expectedState,
			}).Error("Not matching expected state")
		}
	}

	if res == true {
		log.WithFields(log.Fields{
			"ExpectedState": expectedState,
		}).Infof("%d ONUs matching expected state", len(onus.Items))
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
