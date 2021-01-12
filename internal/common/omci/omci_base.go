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
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/opencord/omci-lib-go"
	"github.com/opencord/voltha-protos/v4/go/openolt"
	log "github.com/sirupsen/logrus"
)

// ParseOpenOltOmciPacket receive an OMCI packet in the openolt format and returns
// an OMCI Layer as per omci-lib-go
func ParseOpenOltOmciPacket(omciMsg *openolt.OmciMsg) (*omci.OMCI, error) {
	rxMsg := hexDecode(omciMsg.Pkt)

	// NOTE this is may not be needed, VOLTHA sends the correct message
	if len(rxMsg) >= 44 {
		trailerLenData := rxMsg[42:44]
		trailerLen := binary.BigEndian.Uint16(trailerLenData)
		if trailerLen != 40 { // invalid base Format entry -> autocorrect
			binary.BigEndian.PutUint16(rxMsg[42:44], 40)
			omciLogger.Debug("cc-corrected-omci-message: trailer len inserted")
		}
	} else {
		omciLogger.WithFields(log.Fields{"Length": len(rxMsg)}).Error("received omci-message too small for OmciBaseFormat - abort")
		return nil, errors.New("received omci-message too small for OmciBaseFormat - abort")
	}

	packet := gopacket.NewPacket(rxMsg, omci.LayerTypeOMCI, gopacket.NoCopy)
	if packet == nil {
		omciLogger.Error("omci-message could not be decoded")
		return nil, errors.New("omci-message could not be decoded")
	}
	omciLayer := packet.Layer(omci.LayerTypeOMCI)
	if omciLayer == nil {
		omciLogger.Error("omci-message could not decode omci layer")
		return nil, fmt.Errorf("omci-message could not decode omci layer")
	}
	parsed, ok := omciLayer.(*omci.OMCI)
	if !ok {
		omciLogger.Error("omci-message could not assign omci layer")
		return nil, errors.New("omci-message could not assign omci layer")
	}

	return parsed, nil
}

// HexDecode converts the hex encoding to binary
func hexDecode(pkt []byte) []byte {
	p := make([]byte, len(pkt)/2)
	for i, j := 0, 0; i < len(pkt); i, j = i+2, j+1 {
		// Go figure this ;)
		u := (pkt[i] & 15) + (pkt[i]>>6)*9
		l := (pkt[i+1] & 15) + (pkt[i+1]>>6)*9
		p[j] = u<<4 + l
	}
	omciLogger.Tracef("Omci decoded: %x.", p)
	return p
}
