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

package devices

import (
	"bytes"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/looplab/fsm"
	"github.com/opencord/bbsim/internal/bbsim/packetHandlers"
	"github.com/opencord/bbsim/internal/bbsim/types"
	log "github.com/sirupsen/logrus"
	"os/exec"
)

var (
	nniLogger    = log.WithFields(log.Fields{"module": "NNI"})
	nniVeth      = "nni"
	upstreamVeth = "upstream"
	dhcpServerIp = "182.21.0.128"
)

func CreateNNI(olt *OltDevice) (NniPort, error) {
	nniPort := NniPort{
		ID: uint32(0),
		OperState: getOperStateFSM(func(e *fsm.Event) {
			oltLogger.Debugf("Changing NNI OperState from %s to %s", e.Src, e.Dst)
		}),
		Type: "nni",
	}
	createNNIPair(olt)
	return nniPort, nil
}

// sendNniPacket will send a packet out of the NNI interface.
// We will send upstream only DHCP packets and drop anything else
func sendNniPacket(packet gopacket.Packet) error {
	if isDhcp := packetHandlers.IsDhcpPacket(packet); !isDhcp {
		nniLogger.WithFields(log.Fields{
			"packet": packet,
		}).Trace("Dropping NNI packet as it's not DHCP")
	}

	packet, err := packetHandlers.PopDoubleTag(packet)
	if err != nil {
		nniLogger.WithFields(log.Fields{
			"packet": packet,
		}).Errorf("Can't remove double tags from packet: %v", err)
		return err
	}

	handle, err := getVethHandler(nniVeth)
	if err != nil {
		return err
	}

	err = handle.WritePacketData(packet.Data())
	if err != nil {
		nniLogger.WithFields(log.Fields{
			"packet": packet,
		}).Errorf("Failed to send packet out of the NNI: %s", err)
		return err
	}

	nniLogger.Infof("Sent packet out of NNI")
	return nil
}

// createNNIBridge will create a veth bridge to fake the connection between the NNI port
// and something upstream, in this case a DHCP server.
// It is also responsible to start the DHCP server itself
func createNNIPair(olt *OltDevice) error {

	if err := exec.Command("ip", "link", "add", nniVeth, "type", "veth", "peer", "name", upstreamVeth).Run(); err != nil {
		nniLogger.Errorf("Couldn't create veth pair between %s and %s", nniVeth, upstreamVeth)
		return err
	}

	if err := setVethUp(nniVeth); err != nil {
		return err
	}

	if err := setVethUp(upstreamVeth); err != nil {
		return err
	}

	if err := startDHCPServer(); err != nil {
		return err
	}

	ch, err := listenOnVeth(nniVeth)
	if err != nil {
		return err
	}
	olt.nniPktInChannel = ch
	return nil
}

// setVethUp is responsible to activate a virtual interface
func setVethUp(vethName string) error {
	if err := exec.Command("ip", "link", "set", vethName, "up").Run(); err != nil {
		nniLogger.Errorf("Couldn't change interface %s state to up: %v", vethName, err)
		return err
	}
	return nil
}

func startDHCPServer() error {
	if err := exec.Command("ip", "addr", "add", dhcpServerIp, "dev", upstreamVeth).Run(); err != nil {
		nniLogger.Errorf("Couldn't assing ip %s to interface %s: %v", dhcpServerIp, upstreamVeth, err)
		return err
	}

	if err := setVethUp(upstreamVeth); err != nil {
		return err
	}

	dhcp := "/usr/local/bin/dhcpd"
	conf := "/etc/dhcp/dhcpd.conf" // copied in the container from configs/dhcpd.conf
	logfile := "/tmp/dhcplog"
	var stderr bytes.Buffer
	cmd := exec.Command(dhcp, "-cf", conf, upstreamVeth, "-tf", logfile)
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		nniLogger.Errorf("Fail to start DHCP Server: %s, %s", err, stderr.String())
		return err
	}
	nniLogger.Info("Successfully activated DHCP Server")
	return nil
}

func getVethHandler(vethName string) (*pcap.Handle, error) {
	var (
		device            = vethName
		snapshotLen int32 = 1518
		promiscuous       = false
		timeout           = pcap.BlockForever
	)
	handle, err := pcap.OpenLive(device, snapshotLen, promiscuous, timeout)
	if err != nil {
		nniLogger.Errorf("Can't retrieve handler for interface %s", vethName)
		return nil, err
	}
	return handle, nil
}

func listenOnVeth(vethName string) (chan *types.PacketMsg, error) {

	handle, err := getVethHandler(vethName)
	if err != nil {
		return nil, err
	}

	channel := make(chan *types.PacketMsg, 32)

	go func() {
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for packet := range packetSource.Packets() {

			if !packetHandlers.IsIncomingPacket(packet) {
				nniLogger.Tracef("Ignoring packet as it's going out")
				continue
			}

			nniLogger.WithFields(log.Fields{
				"packet": packet.Dump(),
			}).Tracef("Received packet on NNI Port")
			pkt := types.PacketMsg{
				Pkt: packet,
			}
			channel <- &pkt
		}
	}()

	return channel, nil
}
