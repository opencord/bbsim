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

package omci

import (
	"encoding/hex"
	"github.com/cboling/omci"
	"github.com/google/gopacket"
	log "github.com/sirupsen/logrus"
)

var omciLogger = log.WithFields(log.Fields{
	"module": "OMCI",
})

func stringToPacket(input string) ([]byte, error) {
	var p []byte

	p, err := hex.DecodeString(input)
	if err != nil {
		omciLogger.Error(err)
		return nil, err
	}
	return p, nil
}

func CreateMibResetRequest() []byte {

	mibResetRequest := "00014F0A000200000000000000000000" +
		"00000000000000000000000000000000" +
		"000000000000000000000028"

	data, err := stringToPacket(mibResetRequest)

	if err != nil {
		return nil
	}

	packet := gopacket.NewPacket(data, omci.LayerTypeOMCI, gopacket.NoCopy)

	//omciLayer := packet.Layer(omci.LayerTypeOMCI)

	//omciMsg, ok := omciLayer.(*omci.OMCI)

	//if !ok {
	//	omciLogger.Error("can't get omci msg")
	//}

	msgLayer := packet.Layer(omci.LayerTypeMibResetRequest)

	omciMsg2, ok2 := msgLayer.(*omci.MibResetRequest)

	if !ok2 {
		omciLogger.Error("can't get omciMsg2 msg")
	}

	// Test serialization back to former string
	//var options gopacket.SerializeOptions
	//options.FixLengths = true
	//
	//buffer := gopacket.NewSerializeBuffer()
	//err = gopacket.SerializeLayers(buffer, options, omciMsg, omciMsg2)
	//
	//outgoingPacket := buffer.Bytes()

	return omciMsg2.Payload
}
