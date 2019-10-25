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
	"github.com/opencord/bbsim/internal/bbsim/responders/dhcp"
	"github.com/opencord/bbsim/internal/bbsim/types"
	log "github.com/sirupsen/logrus"
	"os/exec"
)

var (
	nniLogger    = log.WithFields(log.Fields{"module": "NNI"})
	nniVeth      = "nni"
	upstreamVeth = "upstream"
	dhcpServerIp = "192.168.254.1"
)

type Executor interface {
	Command(name string, arg ...string) Runnable
}

type DefaultExecutor struct{}

func (d DefaultExecutor) Command(name string, arg ...string) Runnable {
	return exec.Command(name, arg...)
}

type Runnable interface {
	Run() error
}

var executor = DefaultExecutor{}

type NniPort struct {
	// BBSIM Internals
	ID uint32

	// PON Attributes
	OperState  *fsm.FSM
	Type       string
	DhcpServer *dhcp.DhcpServer
}

func CreateNNI(olt *OltDevice) (NniPort, error) {
	dhcpServer, _ := dhcp.CreateDhcpServer("192.168.0.0/24")
	nniPort := NniPort{
		ID: uint32(0),
		OperState: getOperStateFSM(func(e *fsm.Event) {
			oltLogger.Debugf("Changing NNI OperState from %s to %s", e.Src, e.Dst)
		}),
		Type:       "nni",
		DhcpServer: dhcpServer,
	}
	// FIXME this should be create in the CreateOLT method
	olt.nniPktInChannel = make(chan *types.PacketMsg, 32)

	//createNNIPair(executor, olt)
	return nniPort, nil
}

// The only packets we expect BBSim to handle are DHCP packets
// this function will forge a response based on the packet and send it back on the NNI channel
func sendNniPacket(packet gopacket.Packet, olt OltDevice) error {
	if isDhcp := packetHandlers.IsDhcpPacket(packet); !isDhcp {
		nniLogger.WithFields(log.Fields{
			"packet": packet,
		}).Trace("Dropping NNI packet as it's not DHCP")
		return nil
	}

	responsePacket, err := olt.Nnis[0].DhcpServer.HandleDhcpReq(packet)

	if err != nil {
		nniLogger.WithFields(log.Fields{
			"packet": packet,
		}).Errorf("Failed to process packet on the NNI port: %s", err)
	}

	pkt := types.PacketMsg{
		Pkt: responsePacket,
	}
	olt.nniPktInChannel <- &pkt

	//packet, err := packetHandlers.PopDoubleTag(packet)
	//if err != nil {
	//	nniLogger.WithFields(log.Fields{
	//		"packet": packet,
	//	}).Errorf("Can't remove double tags from packet: %v", err)
	//	return err
	//}
	//
	//handle, err := getVethHandler(nniVeth)
	//if err != nil {
	//	return err
	//}
	//
	//err = handle.WritePacketData(packet.Data())
	//if err != nil {
	//	nniLogger.WithFields(log.Fields{
	//		"packet": packet,
	//	}).Errorf("Failed to send packet out of the NNI: %s", err)
	//	return err
	//}
	//
	//nniLogger.Infof("Sent packet out of NNI")
	return nil
}

//createNNIBridge will create a veth bridge to fake the connection between the NNI port
//and something upstream, in this case a DHCP server.
//It is also responsible to start the DHCP server itself
func createNNIPair(executor Executor, olt *OltDevice) error {

	if err := executor.Command("ip", "link", "add", nniVeth, "type", "veth", "peer", "name", upstreamVeth).Run(); err != nil {
		nniLogger.Errorf("Couldn't create veth pair between %s and %s", nniVeth, upstreamVeth)
		return err
	}

	if err := setVethUp(executor, nniVeth); err != nil {
		return err
	}

	if err := setVethUp(executor, upstreamVeth); err != nil {
		return err
	}

	if err := startDHCPServer(); err != nil {
		return err
	}

	_, err := listenOnVeth(nniVeth)
	if err != nil {
		return err
	}

	return nil
}

// setVethUp is responsible to activate a virtual interface
func setVethUp(executor Executor, vethName string) error {
	if err := executor.Command("ip", "link", "set", vethName, "up").Run(); err != nil {
		nniLogger.Errorf("Couldn't change interface %s state to up: %v", vethName, err)
		return err
	}
	return nil
}

var startDHCPServer = func() error {
	if err := exec.Command("ip", "addr", "add", dhcpServerIp, "dev", upstreamVeth).Run(); err != nil {
		nniLogger.Errorf("Couldn't assing ip %s to interface %s: %v", dhcpServerIp, upstreamVeth, err)
		return err
	}

	if err := setVethUp(executor, upstreamVeth); err != nil {
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

var listenOnVeth = func(vethName string) (chan *types.PacketMsg, error) {

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
