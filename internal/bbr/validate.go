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
	"github.com/opencord/bbsim/api/bbsim"
	log "github.com/sirupsen/logrus"
	"time"
)

func ValidateAndClose(olt *OLTMock) {

	// connect to the BBSim control APIs to check that all the ONUs are in the correct state
	client, conn := ApiConnect(olt.TargetIp, olt.TargetApiPort)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	onus, err := client.GetONUs(ctx, &bbsim.Empty{})

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatalf("Can't reach BBSim API")
	}

	expectedState := "eap_start_sent"

	for _, onu := range onus.Items {
		if onu.InternalState != expectedState {
			log.WithFields(log.Fields{
				"OnuSN":         onu.SerialNumber,
				"OnuId":         onu.ID,
				"InternalState": onu.InternalState,
				"ExpectedSatte": expectedState,
			}).Error("Not matching expected state")
		}
	}
	log.WithFields(log.Fields{
		"ExpectedSatte": expectedState,
	}).Infof("%d ONUs matching expected state", len(onus.Items))

	olt.conn.Close()
}
