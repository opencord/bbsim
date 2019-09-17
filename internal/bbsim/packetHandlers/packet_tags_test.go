package packetHandlers

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"gotest.tools/assert"
	"net"
	"os"
	"testing"
)

func setUp() {
	fmt.Println("Test Setup")
}

func tearDown() {
	fmt.Println("Test Teardown")
}

func TestMain(m *testing.M) {
	setUp()
	code := m.Run()
	tearDown()
	os.Exit(code)
}

// GO111MODULE=on go test -v -mod vendor ./internal/bbsim/... -run TestPushSingleTag
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
	gopacket.SerializeLayers(
		buffer,
		gopacket.SerializeOptions{
			FixLengths: false,
		},
		ethernetLayer,
		gopacket.Payload(rawBytes),
	)

	untaggedPkt := gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeEthernet, gopacket.Default)
	taggedPkt, err := PushSingleTag(111, untaggedPkt)
	if err != nil {
		t.Fail()
		t.Logf("Error in PushSingleTag: %v", err)
	}

	vlan, _ := getVlanTag(taggedPkt)
	assert.Equal(t, vlan, uint16(111))
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
	gopacket.SerializeLayers(
		buffer,
		gopacket.SerializeOptions{
			FixLengths: false,
		},
		ethernetLayer,
		dot1qLayer,
		gopacket.Payload(rawBytes),
	)

	untaggedPkt := gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeEthernet, gopacket.Default)
	taggedPkt, err := PopSingleTag(untaggedPkt)
	if err != nil {
		t.Fail()
		t.Logf("Error in PushSingleTag: %v", err)
	}

	vlan, err := getVlanTag(taggedPkt)
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
	gopacket.SerializeLayers(
		buffer,
		gopacket.SerializeOptions{
			FixLengths: false,
		},
		ethernetLayer,
		dot1qLayer,
		gopacket.Payload(rawBytes),
	)

	untaggedPkt := gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeEthernet, gopacket.Default)
	taggedPkt, err := PopDoubleTag(untaggedPkt)
	if err != nil {
		t.Fail()
		t.Logf("Error in PushSingleTag: %v", err)
	}

	vlan, err := getVlanTag(taggedPkt)
	assert.Equal(t, vlan, uint16(0))
	assert.Equal(t, err.Error(), "no-dot1q-layer-in-packet")
}
