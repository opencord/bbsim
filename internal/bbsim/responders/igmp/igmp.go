package igmp

import (
	//	"context"
	//	"crypto/md5"
	//	"errors"
        "encoding/binary"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	bbsim "github.com/opencord/bbsim/internal/bbsim/types"
	"net"
	"time"
)

func SendIGMPMembershipReportV2(ponPortId uint32, onuId uint32, macAddress net.HardwareAddr, stream bbsim.Stream) {
	igmp := createIgmpV2Packet()
	pkt, err := serializeIgmpPacket(ponPortId, onuId, macAddress, igmp)

	if err != nil {
		//TODO : Error Handling
	}
	if pkt != nil {
	}
	//TODO : create bbsim message and send to stream

        //      msg := bbsim.ByteMsg{
        //              IntfId: ponPortId,
        //              OnuId:  onuId,
        //              Bytes:  pkt,
        //      }

        //TODO: Check how to create openolt.Indication packet from bbsim msg, i.e. what will be port, gemID values
        //      err = stream.Send(&bbsim.ByteMsg{Data: msg})

}

func createIgmpV2Packet() *layers.IGMP {

	igmpDefault := layers.IGMP{
		Type:            layers.IGMPMembershipReportV2,
		MaxResponseTime: time.Duration(1),
		Checksum:        0,
		GroupAddress:    net.IPv4zero,
		Version:         2,
	}

	//returning igmp packet
	return &igmpDefault
}

func serializeIgmpPacket(intfId uint32, onuId uint32, srcMac net.HardwareAddr, igmp *layers.IGMP) ([]byte, error) {
	buffer := gopacket.NewSerializeBuffer()
	options := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	ethernetLayer := &layers.Ethernet{
		SrcMAC:       srcMac,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeIPv4,
	}

	ipLayer := &layers.IPv4{
		Version:  4,
		TOS:      0x10,
		TTL:      128,
		SrcIP:    []byte{0, 0, 0, 0},
		DstIP:    []byte{255, 255, 255, 255},
		Protocol: layers.IPProtocolIGMP,
	}

	if err := gopacket.SerializeLayers(buffer, options, ethernetLayer, ipLayer); err != nil {
		return nil, err
	}

	bytes := buffer.Bytes()
	return bytes, nil
}

func SendIGMPLeaveGroup(stream bbsim.Stream) {
	//TODO : Implement IGMPLeave
}

//-----------------------------------------***********************---------------------------------
//TODO : Below code for testing changes only. Do not submit. This is already available or to be added in gopacket

// BaseLayer is a convenience struct which implements the LayerData and
// LayerPayload functions of the Layer interface.
type BaseLayer struct {
        // Contents is the set of bytes that make up this layer.  IE: for an
        // Ethernet packet, this would be the set of bytes making up the
        // Ethernet frame.
        Contents []byte
        // Payload is the set of bytes contained by (but not part of) this
        // Layer.  Again, to take Ethernet as an example, this would be the
        // set of bytes encapsulated by the Ethernet protocol.
        Payload []byte
}

type IGMPType uint8

type IGMP struct {
        BaseLayer
        Type                    IGMPType
        MaxResponseTime         time.Duration
        Checksum                uint16
        GroupAddress            net.IP
        SupressRouterProcessing bool
        RobustnessValue         uint8
        IntervalTime            time.Duration
        SourceAddresses         []net.IP
        NumberOfGroupRecords    uint16
        NumberOfSources         uint16
        //      GroupRecords            []IGMPv3GroupRecord
        Version uint8 // IGMP protocol version
}

// SerializeTo writes the serialized form of this layer into the
// SerializationBuffer, implementing gopacket.SerializableLayer.
// See the docs for gopacket.SerializableLayer for more info.
// SerializeTo writes the serialized form of this layer into the
// SerializationBuffer, implementing gopacket.SerializableLayer.
// See the docs for gopacket.SerializableLayer for more info.
func (igmp *IGMP) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
        //      plen := len(igmp)
        //TODO - add  length check here
        data, err := b.PrependBytes(12)
        if err != nil {
                return err
        }
        data[0] = byte(igmp.Type)
        data[1] = byte(igmp.MaxResponseTime)
        if opts.ComputeChecksums {
                data[2] = 0
                data[3] = 0
                igmp.Checksum = tcpipChecksum(b.Bytes(), 0)
        }
        binary.BigEndian.PutUint16(data[2:4], igmp.Checksum)
        copy(data[5:], igmp.GroupAddress)
        return nil
}

// Calculate the TCP/IP checksum defined in rfc1071.  The passed-in csum is any
// initial checksum data that's already been computed.
func tcpipChecksum(data []byte, csum uint32) uint16 {
        // to handle odd lengths, we loop to length - 1, incrementing by 2, then
        // handle the last byte specifically by checking against the original
        // length.
        length := len(data) - 1
        for i := 0; i < length; i += 2 {
                // For our test packet, doing this manually is about 25% faster
                // (740 ns vs. 1000ns) than doing it by calling binary.BigEndian.Uint16.
                csum += uint32(data[i]) << 8
                csum += uint32(data[i+1])
        }
        if len(data)%2 == 1 {
                csum += uint32(data[length]) << 8
        }
        for csum > 0xffff {
                csum = (csum >> 16) + (csum & 0xffff)
        }
        return ^uint16(csum)
}

