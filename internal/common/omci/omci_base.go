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

package omci

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/gopacket"
	"github.com/opencord/omci-lib-go/v2"
	log "github.com/sirupsen/logrus"
)

const DeviceIdentifierPos = 3

// ParseOpenOltOmciPacket receive an OMCI packet in the openolt format and returns
// an OMCI Layer as per omci-lib-go
func ParseOpenOltOmciPacket(pkt []byte) (gopacket.Packet, *omci.OMCI, error) {
	rxMsg := HexDecode(pkt)
	if len(rxMsg) < 10 {
		omciLogger.WithFields(log.Fields{"Length": len(rxMsg)}).Error("received omci message is generally too small - abort")
		return nil, nil, errors.New("received omci message is generally too small - abort")
	}
	if rxMsg[DeviceIdentifierPos] == byte(omci.BaselineIdent) {
		// NOTE this is may not be needed, VOLTHA sends the correct message
		if len(rxMsg) >= 44 {
			trailerLenData := rxMsg[42:44]
			trailerLen := binary.BigEndian.Uint16(trailerLenData)
			if trailerLen != 40 { // invalid base Format entry -> autocorrect
				binary.BigEndian.PutUint16(rxMsg[42:44], 40)
				omciLogger.Trace("cc-corrected-omci-message: trailer len inserted")
			}
		} else {
			omciLogger.WithFields(log.Fields{"Length": len(rxMsg)}).Error("received omci-message too small for OmciBaseFormat - abort")
			return nil, nil, errors.New("received omci-message too small for OmciBaseFormat - abort")
		}
	} else if rxMsg[DeviceIdentifierPos] == byte(omci.ExtendedIdent) {
		if len(rxMsg) > 1980 {
			omciLogger.WithFields(log.Fields{"Length": len(rxMsg)}).Error("received omci-message with wrong length for OmciExtendedFormat - abort")
			return nil, nil, errors.New("received omci-message with wrong length for OmciExtendedFormat - abort")
		}
	} else {
		omciLogger.WithFields(log.Fields{"DeviceIdent": rxMsg[DeviceIdentifierPos]}).Error("received omci-message with wrong Device Identifier - abort")
		return nil, nil, errors.New("received omci-message with wrong Device Identifier - abort")
	}

	packet := gopacket.NewPacket(rxMsg, omci.LayerTypeOMCI, gopacket.NoCopy)
	if packet == nil {
		omciLogger.Error("omci-message could not be decoded")
		return nil, nil, errors.New("omci-message could not be decoded")
	}
	omciLayer := packet.Layer(omci.LayerTypeOMCI)
	if omciLayer == nil {
		omciLogger.Error("omci-message could not decode omci layer")
		return nil, nil, fmt.Errorf("omci-message could not decode omci layer")
	}
	parsed, ok := omciLayer.(*omci.OMCI)
	if !ok {
		omciLogger.Error("omci-message could not assign omci layer")
		return nil, nil, errors.New("omci-message could not assign omci layer")
	}

	return packet, parsed, nil
}

// HexDecode converts the hex encoding to binary
func HexDecode(pkt []byte) []byte {
	p := make([]byte, len(pkt)/2)
	for i, j := 0, 0; i < len(pkt); i, j = i+2, j+1 {
		// Go figure this ;)
		u := (pkt[i] & 15) + (pkt[i]>>6)*9
		l := (pkt[i+1] & 15) + (pkt[i+1]>>6)*9
		p[j] = u<<4 + l
	}
	return p
}

// HexEncode convert binary to hex
func HexEncode(omciPkt []byte) ([]byte, error) {
	dst := make([]byte, hex.EncodedLen(len(omciPkt)))
	hex.Encode(dst, omciPkt)
	return dst, nil
}
