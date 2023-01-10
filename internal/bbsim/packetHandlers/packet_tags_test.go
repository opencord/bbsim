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

package packetHandlers_test

import (
	"net"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/opencord/bbsim/internal/bbsim/packetHandlers"
	"gotest.tools/assert"
)

func TestPushSingleTag(t *testing.T) {
	rawBytes := []byte{10, 20, 30}
	srcMac := net.HardwareAddr{0xff, 0xff, 0xff, 0xff, byte(1), byte(1)}
	dstMac := net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	ethernetLayer := &layers.Ethernet{
		SrcMAC:       srcMac,
		DstMAC:       dstMac,
		EthernetType: 0x8100,
	}

	buffer := gopacket.NewSerializeBuffer()
	_ = gopacket.SerializeLayers(
		buffer,
		gopacket.SerializeOptions{
			FixLengths: false,
		},
		ethernetLayer,
		gopacket.Payload(rawBytes),
	)

	untaggedPkt := gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeEthernet, gopacket.Default)

	taggedPkt, err := packetHandlers.PushSingleTag(111, untaggedPkt, 7)
	if err != nil {
		t.Fail()
		t.Logf("Error in PushSingleTag: %v", err)
	}

	vlan, _ := packetHandlers.GetVlanTag(taggedPkt)
	pbit, _ := packetHandlers.GetPbit(taggedPkt)

	assert.Equal(t, vlan, uint16(111))
	assert.Equal(t, pbit, uint8(7))
}

func TestPushDoubleTag(t *testing.T) {
	rawBytes := []byte{10, 20, 30}
	srcMac := net.HardwareAddr{0xff, 0xff, 0xff, 0xff, byte(1), byte(1)}
	dstMac := net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	ethernetLayer := &layers.Ethernet{
		SrcMAC:       srcMac,
		DstMAC:       dstMac,
		EthernetType: 0x8100,
	}

	buffer := gopacket.NewSerializeBuffer()
	_ = gopacket.SerializeLayers(
		buffer,
		gopacket.SerializeOptions{
			FixLengths: false,
		},
		ethernetLayer,
		gopacket.Payload(rawBytes),
	)

	untaggedPkt := gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeEthernet, gopacket.Default)
	taggedPkt, err := packetHandlers.PushDoubleTag(900, 800, untaggedPkt, 0)
	if err != nil {
		t.Fail()
		t.Logf("Error in PushSingleTag: %v", err)
	}

	sTag, err := packetHandlers.GetVlanTag(taggedPkt)
	if err != nil {
		t.Fatalf(err.Error())
	}
	singleTagPkt, err := packetHandlers.PopSingleTag(taggedPkt)
	if err != nil {
		t.Fatalf(err.Error())
	}
	cTag, err := packetHandlers.GetVlanTag(singleTagPkt)
	if err != nil {
		t.Fatalf(err.Error())
	}

	assert.Equal(t, sTag, uint16(900))
	assert.Equal(t, cTag, uint16(800))
}

func TestPopSingleTag(t *testing.T) {
	rawBytes := []byte{10, 20, 30}
	srcMac := net.HardwareAddr{0xff, 0xff, 0xff, 0xff, byte(1), byte(1)}
	dstMac := net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	ethernetLayer := &layers.Ethernet{
		SrcMAC:       srcMac,
		DstMAC:       dstMac,
		EthernetType: 0x8100,
	}

	dot1qLayer := &layers.Dot1Q{
		Type:           0x8100,
		VLANIdentifier: uint16(111),
	}

	buffer := gopacket.NewSerializeBuffer()
	_ = gopacket.SerializeLayers(
		buffer,
		gopacket.SerializeOptions{
			FixLengths: false,
		},
		ethernetLayer,
		dot1qLayer,
		gopacket.Payload(rawBytes),
	)

	untaggedPkt := gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeEthernet, gopacket.Default)
	taggedPkt, err := packetHandlers.PopSingleTag(untaggedPkt)
	if err != nil {
		t.Fail()
		t.Logf("Error in PushSingleTag: %v", err)
	}

	vlan, _ := packetHandlers.GetVlanTag(taggedPkt)
	assert.Equal(t, vlan, uint16(2580)) // FIXME where dows 2056 comes from??
}

func TestPopDoubleTag(t *testing.T) {
	rawBytes := []byte{10, 20, 30}
	srcMac := net.HardwareAddr{0xff, 0xff, 0xff, 0xff, byte(1), byte(1)}
	dstMac := net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	ethernetLayer := &layers.Ethernet{
		SrcMAC:       srcMac,
		DstMAC:       dstMac,
		EthernetType: 0x8100,
	}

	dot1qLayer := &layers.Dot1Q{
		Type:           0x8100,
		VLANIdentifier: uint16(111),
	}

	buffer := gopacket.NewSerializeBuffer()
	_ = gopacket.SerializeLayers(
		buffer,
		gopacket.SerializeOptions{
			FixLengths: false,
		},
		ethernetLayer,
		dot1qLayer,
		gopacket.Payload(rawBytes),
	)

	untaggedPkt := gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeEthernet, gopacket.Default)
	taggedPkt, err := packetHandlers.PopDoubleTag(untaggedPkt)
	if err != nil {
		t.Fail()
		t.Logf("Error in PushSingleTag: %v", err)
	}

	vlan, err := packetHandlers.GetVlanTag(taggedPkt)
	assert.Equal(t, vlan, uint16(0))
	assert.Equal(t, err.Error(), "no-dot1q-layer-in-packet")
}

func TestGetTagsFromPacket(t *testing.T) {
	rawBytes := []byte{10, 20, 30}
	srcMac := net.HardwareAddr{0xff, 0xff, 0xff, 0xff, byte(1), byte(1)}
	dstMac := net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	ethernetLayer := &layers.Ethernet{
		SrcMAC:       srcMac,
		DstMAC:       dstMac,
		EthernetType: 0x800,
	}

	buffer := gopacket.NewSerializeBuffer()
	_ = gopacket.SerializeLayers(
		buffer,
		gopacket.SerializeOptions{
			FixLengths: false,
		},
		ethernetLayer,
		gopacket.Payload(rawBytes),
	)
	untaggedPkt := gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeEthernet, gopacket.Default)
	singleTaggedPkt, err := packetHandlers.PushSingleTag(111, untaggedPkt, 0)
	assert.NilError(t, err)
	doubleTaggedPkt, err := packetHandlers.PushDoubleTag(111, 222, untaggedPkt, 0)
	assert.NilError(t, err)

	_, _, err = packetHandlers.GetTagsFromPacket(untaggedPkt)
	assert.Equal(t, err.Error(), "no-dot1q-layer-in-packet")

	oTag, iTag, err := packetHandlers.GetTagsFromPacket(singleTaggedPkt)
	assert.NilError(t, err)
	assert.Equal(t, uint16(111), oTag)
	assert.Equal(t, uint16(0), iTag)

	oTag, iTag, err = packetHandlers.GetTagsFromPacket(doubleTaggedPkt)
	assert.NilError(t, err)
	assert.Equal(t, uint16(111), oTag)
	assert.Equal(t, uint16(222), iTag)
}
