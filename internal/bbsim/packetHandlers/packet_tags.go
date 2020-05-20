// Copyright 2018-present Open Networking Foundation
// SPDX-FileCopyrightText: 2019 Open Networking Foundation <info@opennetworking.org>
// SPDX-FileCopyrightText: 2020 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package packetHandlers

import (
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func PushSingleTag(tag int, pkt gopacket.Packet) (gopacket.Packet, error) {
	// TODO can this method be semplified?
	if eth := getEthernetLayer(pkt); eth != nil {
		ethernetLayer := &layers.Ethernet{
			SrcMAC:       eth.SrcMAC,
			DstMAC:       eth.DstMAC,
			EthernetType: 0x8100,
		}

		dot1qLayer := &layers.Dot1Q{
			Type:           eth.EthernetType,
			VLANIdentifier: uint16(tag),
		}

		buffer := gopacket.NewSerializeBuffer()
		gopacket.SerializeLayers(
			buffer,
			gopacket.SerializeOptions{
				FixLengths: false,
			},
			ethernetLayer,
			dot1qLayer,
			gopacket.Payload(eth.Payload),
		)
		ret := gopacket.NewPacket(
			buffer.Bytes(),
			layers.LayerTypeEthernet,
			gopacket.Default,
		)

		return ret, nil
	}
	return nil, errors.New("Couldn't extract LayerTypeEthernet from packet")
}

func PushDoubleTag(stag int, ctag int, pkt gopacket.Packet) (gopacket.Packet, error) {

	singleTaggedPkt, err := PushSingleTag(ctag, pkt)
	if err != nil {
		return nil, err
	}
	doubleTaggedPkt, err := PushSingleTag(stag, singleTaggedPkt)
	if err != nil {
		return nil, err
	}

	return doubleTaggedPkt, nil
}

func PopSingleTag(pkt gopacket.Packet) (gopacket.Packet, error) {
	layer, err := getDot1QLayer(pkt)
	if err != nil {
		return nil, err
	}

	if eth := getEthernetLayer(pkt); eth != nil {
		ethernetLayer := &layers.Ethernet{
			SrcMAC:       eth.SrcMAC,
			DstMAC:       eth.DstMAC,
			EthernetType: layer.Type,
		}
		buffer := gopacket.NewSerializeBuffer()
		gopacket.SerializeLayers(buffer, gopacket.SerializeOptions{},
			ethernetLayer,
			gopacket.Payload(layer.Payload),
		)
		retpkt := gopacket.NewPacket(
			buffer.Bytes(),
			layers.LayerTypeEthernet,
			gopacket.Default,
		)

		return retpkt, nil
	}
	return nil, errors.New("no-ethernet-layer")
}

func PopDoubleTag(pkt gopacket.Packet) (gopacket.Packet, error) {
	packet, err := PopSingleTag(pkt)
	if err != nil {
		return nil, err
	}
	packet, err = PopSingleTag(packet)
	if err != nil {
		return nil, err
	}
	return packet, nil
}

func getEthernetLayer(pkt gopacket.Packet) *layers.Ethernet {
	eth := &layers.Ethernet{}
	if ethLayer := pkt.Layer(layers.LayerTypeEthernet); ethLayer != nil {
		eth, _ = ethLayer.(*layers.Ethernet)
	}
	return eth
}

func getDot1QLayer(pkt gopacket.Packet) (*layers.Dot1Q, error) {
	dot1q := &layers.Dot1Q{}
	if dot1qLayer := pkt.Layer(layers.LayerTypeDot1Q); dot1qLayer != nil {
		dot1q = dot1qLayer.(*layers.Dot1Q)
		return dot1q, nil
	}
	return nil, errors.New("no-dot1q-layer-in-packet")
}

func GetVlanTag(pkt gopacket.Packet) (uint16, error) {
	dot1q, err := getDot1QLayer(pkt)
	if err != nil {
		return 0, err
	}
	return dot1q.VLANIdentifier, nil
}
